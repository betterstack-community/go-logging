package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/betterstack-community/go-logging/logger"
)

var HTTPClient = http.Client{
	Timeout: 30 * time.Second,
}

// ResultItem represents each search result item.
type ResultItem struct {
	Timestamp time.Time `json:"timestamp"`
	Title     string    `json:"title"`
	Snippet   string    `json:"snippet"`
	PageID    int       `json:"pageid"`
	Size      int       `json:"size"`
	WordCount int       `json:"wordcount"`
	Ns        int       `json:"ns"`
}

// SearchResponse is the entire JSON response from the Wikipedia API.
type SearchResponse struct {
	BatchComplete string `json:"batchcomplete"`
	Continue      struct {
		Continue string `json:"continue"`
		Sroffset int    `json:"sroffset"`
	} `json:"continue"`
	Query struct {
		Search     []ResultItem `json:"search"`
		SearchInfo struct {
			TotalHits int `json:"totalhits"`
		} `json:"searchinfo"`
	} `json:"query"`
}

type Search struct {
	Results    *SearchResponse
	Query      string
	TotalPages int
	NextPage   int
}

func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}

	return s.NextPage - 1
}

func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}

func searchWikipedia(
	ctx context.Context,
	searchQuery string,
	pageSize, resultsOffset int,
) (*SearchResponse, error) {
	l := logger.FromCtx(ctx)

	resp, err := HTTPClient.Get(
		fmt.Sprintf(
			"https://en.wikipedia.org/w/api.php?action=query&list=search&prop=info&inprop=url&utf8=&format=json&origin=*&srlimit=%d&srsearch=%s&sroffset=%d",
			pageSize,
			url.QueryEscape(searchQuery),
			resultsOffset,
		),
	)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		l.Sugar().Debugf(
			"%d response from Wikipedia: %s",
			resp.StatusCode,
			string(body),
		)

		return nil, fmt.Errorf(
			"unexpected %d response from Wikipedia",
			resp.StatusCode,
		)
	}

	var searchResponse SearchResponse

	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		return nil, err
	}

	return &searchResponse, nil
}
