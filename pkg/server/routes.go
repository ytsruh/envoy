package server

import (
	"net/http"
)

// Represents a set of handlers for the server
type Handlers interface {
	Health() http.HandlerFunc
	Hello() http.HandlerFunc
	Goodbye() http.HandlerFunc
}

// RegisterRoutes sets up all the application routes using Server's dependencies.
func (s *Server) RegisterRoutes(handlers Handlers) {
	s.router.Get("/hello", handlers.Hello())

	s.router.Post("/goodbye", handlers.Goodbye())

	s.router.Get("/health", handlers.Health())

	s.router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
