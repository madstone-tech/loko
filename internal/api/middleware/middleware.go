// Package middleware provides HTTP middleware for the loko API.
package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// writeJSONError writes a JSON error response with the proper Content-Type header.
func writeJSONError(w http.ResponseWriter, statusCode int, errMsg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintf(w, `{"error":%q,"code":%q}`, errMsg, code)
}

// Auth returns middleware that validates bearer token authentication.
func Auth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing authorization header", "UNAUTHORIZED")
				return
			}

			// Validate bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeJSONError(w, http.StatusUnauthorized, "invalid authorization format", "UNAUTHORIZED")
				return
			}

			if parts[1] != apiKey {
				writeJSONError(w, http.StatusUnauthorized, "invalid api key", "UNAUTHORIZED")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Logger returns middleware that logs HTTP requests.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		_, _ = fmt.Fprintf(os.Stderr, "%s %s %s %d %s\n",
			time.Now().Format(time.RFC3339),
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration.Round(time.Millisecond),
		)
	})
}

// CORS returns middleware that adds CORS headers.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Recovery returns middleware that recovers from panics.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "panic recovered: %v\n", err)
				writeJSONError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
