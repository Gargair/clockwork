package http

import (
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// slogMiddleware is an http.Handler that logs requests.
type slogMiddleware struct {
	logger *slog.Logger
	next   stdhttp.Handler
}

func (m slogMiddleware) ServeHTTP(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	start := time.Now()
	ww := &statusWriter{ResponseWriter: w, status: 200}
	m.next.ServeHTTP(ww, r)
	duration := time.Since(start)

	reqID := middleware.GetReqID(r.Context())
	m.logger.Info("http_request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", ww.status),
		slog.Duration("duration", duration),
		slog.String("request_id", reqID),
		slog.String("remote_ip", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
	)
}

// loggingMiddleware wraps next with a slogMiddleware; short wrapper to satisfy chi's signature.
func loggingMiddleware(logger *slog.Logger) func(next stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler { return slogMiddleware{logger: logger, next: next} }
}

type statusWriter struct {
	stdhttp.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
