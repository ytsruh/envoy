package server

import (
	"net/http"

	"ytsruh.com/envoy/pkg/handlers"
)

// RegisterRoutes sets up all the application routes using Server's dependencies.
func (s *Server) RegisterRoutes(router Router) {
	router.Get("/hello", handlers.Hello)

	router.Post("/goodbye", handlers.Goodbye)

	router.Get("/health", handlers.Health(s.db))

	router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
