package main

import (
	"bytes"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

type handlerWithError func(w http.ResponseWriter, r *http.Request) error

func (fn handlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) error {
	log.Println("entered indexHandler()")

	if r.URL.Path != "/" {
		log.Printf("path '%s' not found\n", r.URL.Path)
		http.NotFound(w, r)

		return nil
	}

	buf := &bytes.Buffer{}

	err := tpl.Execute(buf, nil)
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)

	return err
}

func searchHandler(w http.ResponseWriter, r *http.Request) error {
	log.Println("entered searchHandler()")

	u, err := url.Parse(r.URL.String())
	if err != nil {
		return err
	}

	params := u.Query()
	searchQuery := params.Get("q")

	pageNum := params.Get("page")
	if pageNum == "" {
		pageNum = "1"
	}

	log.Printf(
		"received incoming search query: '%s', page: '%s'\n",
		searchQuery,
		pageNum,
	)

	nextPage, err := strconv.Atoi(pageNum)
	if err != nil {
		return err
	}

	pageSize := 20

	resultsOffset := (nextPage - 1) * pageSize

	searchResponse, err := searchWikipedia(searchQuery, pageSize, resultsOffset)
	if err != nil {
		return err
	}

	totalHits := searchResponse.Query.SearchInfo.TotalHits

	search := &Search{
		Query:      searchQuery,
		Results:    searchResponse,
		TotalPages: int(math.Ceil(float64(totalHits) / float64(pageSize))),
		NextPage:   nextPage + 1,
	}

	buf := &bytes.Buffer{}

	err = tpl.Execute(buf, search)
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		return err
	}

	log.Printf("search query succeeded without errors\n")

	return nil
}
