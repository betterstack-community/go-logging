package main

import (
	"context"
	"net/http"
	"time"

	"github.com/betterstack-community/go-logging/logger"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

type contextKey string

const (
	correlationIDCtxKey contextKey = "correlation_id"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// requestLogger attaches a correlation ID to each reqwuest
// and logs all incoming requests to the server.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := logger.Get()

		correlationID := xid.New().String()

		ctx := context.WithValue(
			r.Context(),
			correlationIDCtxKey,
			correlationID,
		)

		r = r.WithContext(ctx)

		l.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str(string(correlationIDCtxKey), correlationID)
		})

		w.Header().Add("X-Correlation-ID", correlationID)

		lrw := newLoggingResponseWriter(w)

		r = r.WithContext(l.WithContext(r.Context()))

		defer func(start time.Time) {
			l.
				Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Str("referrer", r.Referer()).
				Str("user_agent", r.UserAgent()).
				Dur("elapsed_ms", time.Since(start)).
				Int("status_code", lrw.statusCode).
				Msgf("%s request to %s completed", r.Method, r.RequestURI)
		}(time.Now())

		next.ServeHTTP(lrw, r)
	})
}
