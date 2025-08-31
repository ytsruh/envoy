package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HealthChecker minimal interface for health endpoint
type HealthChecker interface {
	Health() map[string]string
}

// Hello handler
func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

// Goodbye handler
func Goodbye(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Goodbye, World!")
}

// Health handler uses HealthChecker
func Health(h HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		healthStatus := h.Health()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(healthStatus)
	}
}
