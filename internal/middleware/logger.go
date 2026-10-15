package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseRecorder) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Flush passes flushes through to the real writer. Without it the reverse proxy's
// FlushInterval = -1 has no effect and SSE responses are buffered until they end.
func (rw *responseRecorder) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap lets http.ResponseController reach the underlying ResponseWriter.
func (rw *responseRecorder) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default to 200
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		slog.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", duration.String(),
			"ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}
