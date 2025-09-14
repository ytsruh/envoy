package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	database "ytsruh.com/envoy/pkg/database/generated"
)

type Handler struct {
	queries database.Querier
}

func New(queries database.Querier) Handler {
	return Handler{queries: queries}
}

// Hello handler
func (h Handler) Hello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	}
}

// Goodbye handler
func (h Handler) Goodbye() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Goodbye, World!")
	}
}

// Health handler uses HealthChecker
func (h Handler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"health": "ok"})
	}
}
