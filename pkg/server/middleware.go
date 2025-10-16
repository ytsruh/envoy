package server

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// Middleware represents a middleware handler
type Middleware func(http.Handler) http.Handler

// RegisterMiddleware chains all middleware together
func (s *Server) RegisterMiddleware() http.Handler {
	middlewares := []Middleware{
		RecoveryMiddleware,
		LoggingMiddleware,
		SecurityHeadersMiddleware,
		CorsMiddleware,
	}
	return ApplyMiddleware(s.server.Handler, middlewares...)
}

// ApplyMiddleware applies a list of middleware handlers to an http.Handler
func ApplyMiddleware(h http.Handler, middlewares ...Middleware) http.Handler {
	// If final handler is nil, replace with explicit 500 handler to avoid typed-nil issues
	if h == nil {
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("PANIC: nil final handler passed to ApplyMiddleware\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		})
	}
	// Apply middleware in reverse order (so the first middleware in the slice is the outermost/last to execute)
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// CorsMiddleware adds CORS headers to responses
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security-related HTTP headers to responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs information about each request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		rw := NewStatusRecorder(w)

		// Process the request
		next.ServeHTTP(rw, r)

		// Log the request details
		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %d %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			rw.status,
			duration,
		)
	})
}

// TimeoutMiddleware adds a timeout to the request context
func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a channel to signal completion
		done := make(chan bool)

		// Create a timeout of 30 seconds
		timeout := time.After(30 * time.Second)

		// Process the request in a goroutine
		go func() {
			next.ServeHTTP(w, r)
			done <- true
		}()

		// Wait for either completion or timeout
		select {
		case <-done:
			// Request completed normally
			return
		case <-timeout:
			// Request timed out
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Request timed out"))
			return
		}
	})
}

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.Handler) http.Handler {
	// make middleware nil-safe: if next is nil, replace with a no-op handler
	if next == nil {
		// Return explicit 500 handler when next is nil to make behavior deterministic
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("PANIC: nil handler passed to RecoveryMiddleware\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error and stack trace
				log.Printf("PANIC: %v\n%s", err, debug.Stack())

				// Return a 500 Internal Server Error
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			}
		}()

		// Protect against nil next even if earlier check missed typed nil
		if next == nil {
			log.Printf("PANIC: nil handler encountered in RecoveryMiddleware during ServeHTTP\n")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// statusRecorder is a custom ResponseWriter that keeps track of the status code
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func NewStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{w, http.StatusOK}
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
