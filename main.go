package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/betterstack-community/go-logging/logger"
)

//go:embed static
var staticFiles embed.FS

//go:embed templates
var indexHTML embed.FS

var tpl *template.Template

func htmlSafe(str string) template.HTML {
	return template.HTML(str)
}

func main() {
	l := logger.Get()

	var err error

	tpl, err = template.New("index.html").Funcs(template.FuncMap{
		"htmlSafe": htmlSafe,
	}).ParseFS(indexHTML, "templates/index.html")
	if err != nil {
		log.Fatalf("failed to initialize HTML templates: %s", err.Error())
	}

	fs := http.FileServer(http.FS(staticFiles))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	mux.Handle("/static/", fs)
	mux.Handle("/search", handlerWithError(searchHandler))
	mux.Handle("/", handlerWithError(indexHandler))

	l.Info().
		Str("port", port).
		Msgf("starting application server on port: %s", port)

	l.Fatal().
		Err(http.ListenAndServe(":"+port, requestLogger(mux))).
		Msg("server closed")
}
