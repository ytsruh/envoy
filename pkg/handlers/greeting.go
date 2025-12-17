package handlers

import (
	"encoding/json"
	"net/http"

	database "ytsruh.com/envoy/pkg/database/generated"
)

type GreetingHandlerImpl struct {
	queries database.Querier
}

func NewGreetingHandler(queries database.Querier) GreetingHandler {
	return &GreetingHandlerImpl{queries: queries}
}

func (h *GreetingHandlerImpl) Hello(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
}

func (h *GreetingHandlerImpl) Goodbye(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"message": "Goodbye, World!"})
}
