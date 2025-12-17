package handlers

import (
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
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
	return nil
}

func (h *GreetingHandlerImpl) Goodbye(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Goodbye, World!"))
	return nil
}
