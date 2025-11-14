package middlewares

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type RequestLogger struct {
	logger *slog.Logger
}

func NewRequestLogger(logger *slog.Logger) *RequestLogger {
	return &RequestLogger{
		logger: logger,
	}
}

func (l *RequestLogger) LogRequest(next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := middleware.GetReqID(r.Context())

		l.logger.Info(
			"request started",
			"http_method", r.Method,
			"http_path", r.URL.Path,
			"request_id", requestID,
		)

		ww := newResponseWriter(w)
		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		var durationStr string
		if duration.Milliseconds() == 0 {
			durationStr = strconv.FormatInt(duration.Microseconds(), 10) + "µs"
		} else {
			durationStr = strconv.FormatInt(duration.Milliseconds(), 10) + "ms"
		}

		l.logger.Info(
			"request finished",
			"duration", durationStr,
			"http_status", ww.status,
			"request_id", requestID,
		)
	})

	return fn
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
