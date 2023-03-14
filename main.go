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
	"strconv"
	"strings"
)

const RESULT_LIMIT = 10

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
}

type APIResponse struct {
	Results     []string `json:"results"`
	ResultCount int      `json:"resultCount"`
	Limit       int      `json:"resultLimit"`
	Page        int      `json:"page"`
	TotalPages  int      `json:"totalPages"`
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		var err error
		pageStr := r.URL.Query().Get("page")
		page := 1

		if pageStr != "" {
			page, err = strconv.Atoi(pageStr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println("Error", err)
				fmt.Fprintf(w, "unable to parse page number")
				return
			}
		}

		results, err := searcher.Search(query[0])

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("Error", err)
			fmt.Fprintf(w, "unable to search")
			return
		}

		resultLen := len(results)
		totalPages := resultLen / RESULT_LIMIT

		res := APIResponse{
			Results:     paginate(results, page, RESULT_LIMIT),
			ResultCount: resultLen,
			Limit:       RESULT_LIMIT,
			Page:        page,
			TotalPages:  totalPages,
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
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(bytes.ToLower(dat))
	return nil
}

func (s *Searcher) Search(query string) ([]string, error) {
	regex, err := regexp.Compile(fmt.Sprintf("(?i)\\w?%s\\w?", query))
	if err != nil {
		return []string{}, err
	}
	idxs := s.SuffixArray.FindAllIndex(regex, -1)
	results := []string{}
	for _, idx := range idxs {
		results = s.markResult(idx, query, results)
	}
	return results, nil
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
