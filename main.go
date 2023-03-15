package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const RESULT_LIMIT = 10
const MAX_INT = int(0 >> 1)
const FIRST_WORK = "the sonnets"

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
	WorkIndexes   map[string]int
}

type Work struct {
	Title   string   `json:"title"`
	Results []string `json:"results"`
}
type APIResponse struct {
	Works []Work `json:"works"`
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		works, err := searcher.Search(query[0])

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("Error", err)
			fmt.Fprintf(w, "unable to search")
			return
		}

		res := APIResponse{
			Works: works,
		}

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err = enc.Encode(res)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	workTitles := getWorkTitles()

	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	rawSuffixArray := suffixarray.New(dat)
	s.SuffixArray = suffixarray.New(bytes.ToLower(dat))

	s.WorkIndexes = make(map[string]int)

	for idx := 0; idx < len(workTitles); idx++ {
		regex := regexp.MustCompile(workTitles[idx])
		titleIndexes := rawSuffixArray.FindAllIndex(regex, 2)
		s.WorkIndexes[workTitles[idx]] = titleIndexes[1][0]
		fmt.Println(workTitles[idx], titleIndexes[1][0])
	}

	return nil
}

func getWorkTitles() []string {
	return []string{
		"THE SONNETS",
		"ALL’S WELL THAT ENDS WELL",
		"THE TRAGEDY OF ANTONY AND CLEOPATRA",
		"AS YOU LIKE IT",
		"THE COMEDY OF ERRORS",
		"THE TRAGEDY OF CORIOLANUS",
		"CYMBELINE",
		"THE TRAGEDY OF HAMLET, PRINCE OF DENMARK",
		"THE FIRST PART OF KING HENRY THE FOURTH",
		"THE SECOND PART OF KING HENRY THE FOURTH",
		"THE LIFE OF KING HENRY THE FIFTH",
		"THE FIRST PART OF HENRY THE SIXTH",
		"THE SECOND PART OF KING HENRY THE SIXTH",
		"THE THIRD PART OF KING HENRY THE SIXTH",
		"KING HENRY THE EIGHTH",
		"KING JOHN",
		"THE TRAGEDY OF JULIUS CAESAR",
		"THE TRAGEDY OF KING LEAR",
		"LOVE’S LABOUR’S LOST",
		"THE TRAGEDY OF MACBETH",
		"MEASURE FOR MEASURE",
		"THE MERCHANT OF VENICE",
		"THE MERRY WIVES OF WINDSOR",
		"A MIDSUMMER NIGHT’S DREAM",
		"MUCH ADO ABOUT NOTHING",
		"THE TRAGEDY OF OTHELLO, MOOR OF VENICE",
		"PERICLES, PRINCE OF TYRE",
		"KING RICHARD THE SECOND",
		"KING RICHARD THE THIRD",
		"THE TRAGEDY OF ROMEO AND JULIET",
		"THE TAMING OF THE SHREW",
		"THE TEMPEST",
		"THE LIFE OF TIMON OF ATHENS",
		"THE TRAGEDY OF TITUS ANDRONICUS",
		"THE HISTORY OF TROILUS AND CRESSIDA",
		"TWELFTH NIGHT; OR, WHAT YOU WILL",
		"THE TWO GENTLEMEN OF VERONA",
		"THE TWO NOBLE KINSMEN",
		"THE WINTER’S TALE",
		"A LOVER’S COMPLAINT",
		"THE PASSIONATE PILGRIM",
		"THE PHOENIX AND THE TURTLE",
		"THE RAPE OF LUCRECE",
		"VENUS AND ADONIS",
	}
}

func (s *Searcher) Search(query string) ([]Work, error) {
	works := []Work{}
	results := []string{}
	count := 0
	workIdx := 0

	regex, err := regexp.Compile(fmt.Sprintf("(?i)\\w?%s\\w?", query))
	if err != nil {
		return []Work{}, err
	}

	workTitles := getWorkTitles()
	queryIdxs := s.SuffixArray.FindAllIndex(regex, -1)

	for _, queryIdx := range queryIdxs {
		if workIdx == len(workTitles) {
			break
		}

		currWorkIdx := s.WorkIndexes[workTitles[workIdx]]
		nextWorkIdx := MAX_INT

		if workIdx < len(workTitles)-1 {
			nextWorkIdx = s.WorkIndexes[workTitles[workIdx+1]]
		}

		startIdx := queryIdx[0]

		if startIdx > currWorkIdx && startIdx < nextWorkIdx {
			results = s.markResult(queryIdx, query, results)
			count++
		} else if startIdx >= nextWorkIdx {
			if len(results) == 0 {
				continue
			}
			work := Work{Title: workTitles[workIdx], Results: results}
			works = append(works, work)
			count++
			workIdx++
			results = []string{}
		} else if count == RESULT_LIMIT {
			if len(results) == 0 {
				break
			}
			work := Work{Title: workTitles[workIdx], Results: results}
			works = append(works, work)
		}
	}

	return works, nil
}

func (s *Searcher) markResult(idx []int, query string, results []string) []string {
	startIdx := idx[0]
	endIdx := idx[1]

	// string of query match
	foundString := s.CompleteWorks[startIdx:endIdx]
	substrIdx := strings.Index(foundString, query)

	// get and mark query substring
	queryIdx := startIdx + substrIdx
	substr := s.CompleteWorks[queryIdx : queryIdx+len(query)]
	substr = fmt.Sprintf("<mark>%s</mark>", substr)

	stringBlock := s.CompleteWorks[queryIdx-250:queryIdx] + substr + s.CompleteWorks[queryIdx+len(query):queryIdx+250]
	results = append(results, stringBlock)
	return results
}

func paginate(x []string, page int, limit int) []string {
	if page > len(x) {
		page = len(x)
	}

	end := page + limit
	if end > len(x) {
		end = len(x)
	}

	return x[page:end]
}
