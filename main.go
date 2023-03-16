package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const MAX_INT = int(0 >> 1)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))
	http.HandleFunc("/work", handleWork(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
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
	Title    string   `json:"title"`
	Contents string   `json:"contents"`
	Results  []string `json:"results"`
}

type SearchResponse struct {
	Works []Work `json:"works"`
}

type WorkResponse struct {
	Title    string `json:"title"`
	Contents string `json:"contents"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			res := ErrorResponse{Message: "missing search query in URL params"}
			writeResponse(nil, res, w, true)

			return
		}

		works, err := searcher.Search(query[0])

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("Error", err)
			// fmt.Fprintf(w, "unable to search")

			res := ErrorResponse{Message: err.Error()}
			writeResponse(err, res, w, true)

			return
		}

		res := SearchResponse{
			Works: works,
		}

		writeResponse(err, res, w, false)
	}
}

func handleWork(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		title, ok := r.URL.Query()["t"]
		if !ok || len(title[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			res := ErrorResponse{Message: "missing title in URL params"}
			writeResponse(nil, res, w, true)

			return
		}

		work, err := searcher.GetWork(title[0])

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("Error", err)

			res := ErrorResponse{Message: err.Error()}
			writeResponse(err, res, w, true)

			return
		}

		res := WorkResponse{
			Title:    work.Title,
			Contents: work.Contents,
		}

		writeResponse(err, res, w, false)
	}
}

func writeResponse(err error, res interface{}, w http.ResponseWriter, useBuffer bool) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err = enc.Encode(res)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("encoding failure"))
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if useBuffer {
		w.Write(buf.Bytes())
		return
	}
	w.Header().Set("Content-Encoding", "gzip")

	gz := gzip.NewWriter(w)
	defer gz.Close()
	gz.Write(buf.Bytes())
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

	// build work start indexes
	for idx := 0; idx < len(workTitles); idx++ {
		regex := regexp.MustCompile(workTitles[idx])

		// find second case-sensitive title
		titleIndexes := rawSuffixArray.FindAllIndex(regex, 2)
		s.WorkIndexes[workTitles[idx]] = titleIndexes[1][0]
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

	// case-insensitive, optional surrounding word character of query
	regex, err := regexp.Compile(fmt.Sprintf("(?i)\\w?%s\\w?", query))
	if err != nil {
		return []Work{}, err
	}

	workTitles := getWorkTitles()
	queryIdxs := s.SuffixArray.FindAllIndex(regex, -1)
	firstWorkIdx := s.WorkIndexes[workTitles[0]]
	queryIdx := 0

	// move query index to after first work index
	for queryIdxs[queryIdx][0] < firstWorkIdx {
		queryIdx++
	}

	// loop through works to match with each query index
	for workIdx := 0; workIdx < len(workTitles) && queryIdx < len(queryIdxs); {

		// find first query result after first work
		queryIdxPoints := queryIdxs[queryIdx]
		queryStartIdx := queryIdxPoints[0]
		currWorkIdx := s.WorkIndexes[workTitles[workIdx]]
		nextWorkIdx := MAX_INT

		if workIdx+1 < len(workTitles) {
			nextWorkIdx = s.WorkIndexes[workTitles[workIdx+1]]
		}


		// build works' results and move work start index forward
		if queryStartIdx > nextWorkIdx {

			if len(results) > 0 {
				work := Work{Title: workTitles[workIdx], Results: results}
				works = append(works, work)
				results = []string{}
			}
			workIdx++
		}

		// find query's matching work
		if queryStartIdx > currWorkIdx && queryStartIdx < nextWorkIdx {
			stringBlock := s.markResult(queryIdxPoints, query, results)
			results = append(results, stringBlock)

			// save last works' results
			if queryIdx+1 == len(queryIdxs) {

				work := Work{Title: workTitles[workIdx], Results: results}
				works = append(works, work)

				break
			}

			// move query indexes
			queryIdx++
			queryIdxPoints = queryIdxs[queryIdx]
		}

	}

	return works, nil
}

func (s *Searcher) markResult(idx []int, query string, results []string) string {
	startIdx := idx[0]
	endIdx := idx[1]

	// string of query match
	foundString := s.CompleteWorks[startIdx:endIdx]
	substrIdx := strings.Index(strings.ToLower(foundString), strings.ToLower(query))

	// get and mark query substring
	queryIdx := startIdx + substrIdx
	substr := s.CompleteWorks[queryIdx : queryIdx+len(query)]
	substr = fmt.Sprintf("<mark>%s</mark>", substr)

	stringBlock := s.CompleteWorks[queryIdx-250:queryIdx] + substr + s.CompleteWorks[queryIdx+len(query):queryIdx+250]

	return stringBlock
}

func (s *Searcher) GetWork(title string) (Work, error) {
	var workEndIdx int
	workStartIdx, ok := s.WorkIndexes[title]
	if !ok {
		return Work{}, errors.New("work not found")
	}

	// find end index
	for _, title := range getWorkTitles() {
		workIdx := s.WorkIndexes[title]
		if workIdx > workStartIdx {
			workEndIdx = workIdx
			break
		}
	}

	newLineLen := len(title + "\r\n\r\n")
	var workContents string

	// get work contents
	if workEndIdx == 0 {
		workContents = s.CompleteWorks[workStartIdx+newLineLen:]
	} else {
		workContents = s.CompleteWorks[workStartIdx+newLineLen : workEndIdx]
	}

	work := Work{Title: title, Contents: strings.TrimSpace(workContents)}

	return work, nil
}
