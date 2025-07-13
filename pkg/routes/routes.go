package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Router defines the methods for registering routes that this package needs.
type Router interface {
	Get(path string, handler http.HandlerFunc)
	Post(path string, handler http.HandlerFunc)
	http.Handler // To allow ServeHTTP for testing
}

// DBService defines the methods for database interactions that this package needs.
type DBService interface {
	Health() map[string]string
}

// RegisterRoutes sets up all the application routes.
func RegisterRoutes(router Router, db DBService) {
	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	router.Post("/goodbye", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Goodbye, World!")
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		healthStatus := db.Health()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(healthStatus)
	})

	router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
