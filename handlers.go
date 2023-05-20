package main

import (
	"bytes"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/rs/zerolog"
)

type handlerWithError func(w http.ResponseWriter, r *http.Request) error

func (fn handlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())

	err := fn(w, r)
	if err != nil {
		l.Error().Err(err).Msg("unexpected error while processing request")
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) error {
	l := zerolog.Ctx(r.Context())

	l.Trace().Msg("entered indexHandler()")

	if r.URL.Path != "/" {
		l.Trace().Msgf("path '%s' not found\n", r.URL.Path)
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
	ctx := r.Context()

	l := zerolog.Ctx(ctx)

	l.Trace().Msg("entered searchHandler()")

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

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("search_query", searchQuery).Str("page_num", pageNum)
	})

	l.Info().Msgf(
		"received incoming search query: '%s', page: '%s'",
		searchQuery,
		pageNum,
	)

	nextPage, err := strconv.Atoi(pageNum)
	if err != nil {
		return err
	}

	pageSize := 20

	resultsOffset := (nextPage - 1) * pageSize

	searchResponse, err := searchWikipedia(
		ctx,
		searchQuery,
		pageSize,
		resultsOffset,
	)
	if err != nil {
		return err
	}

	l.Debug().Interface("search_response", searchResponse).Send()

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

	l.Trace().Msg("search query succeeded without errors")

	return nil
}
