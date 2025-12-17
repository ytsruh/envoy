package handlers

import (
	"encoding/json"
	"net/http"

	database "ytsruh.com/envoy/pkg/database/generated"
)

type HealthHandlerImpl struct {
	queries database.Querier
}

func NewHealthHandler(queries database.Querier) HealthHandler {
	return &HealthHandlerImpl{queries: queries}
}

func (h *HealthHandlerImpl) Health(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"health": "ok"})
}
