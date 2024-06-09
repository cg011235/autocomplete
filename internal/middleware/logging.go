package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs the details of incoming requests and responses.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures the status code
		rw := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(rw, r)

		log.Printf(
			"[%s] %s %s %s %d %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			r.UserAgent(),
			rw.statusCode,
			time.Since(start),
		)
	})
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
