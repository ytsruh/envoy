package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

type EnvironmentVariableHandler interface {
	CreateEnvironmentVariable(w http.ResponseWriter, r *http.Request) error
	GetEnvironmentVariable(w http.ResponseWriter, r *http.Request) error
	ListEnvironmentVariables(w http.ResponseWriter, r *http.Request) error
	UpdateEnvironmentVariable(w http.ResponseWriter, r *http.Request) error
	DeleteEnvironmentVariable(w http.ResponseWriter, r *http.Request) error
}

type EnvironmentVariableHandlerImpl struct {
	queries       database.Querier
	accessControl utils.AccessControlService
}

func NewEnvironmentVariableHandler(queries database.Querier, accessControl utils.AccessControlService) EnvironmentVariableHandler {
	return &EnvironmentVariableHandlerImpl{
		queries:       queries,
		accessControl: accessControl,
	}
}

type CreateEnvironmentVariableRequest struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value" validate:"required,max=255"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateEnvironmentVariableRequest struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value" validate:"required,max=255"`
	Description string `json:"description" validate:"max=500"`
}

type EnvironmentVariableResponse struct {
	ID            int64     `json:"id"`
	EnvironmentID int64     `json:"environment_id"`
	Key           string    `json:"key"`
	Value         string    `json:"value"`
	Description   *string   `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (h *EnvironmentVariableHandlerImpl) CreateEnvironmentVariable(w http.ResponseWriter, r *http.Request) error {
	claims, err := getUserClaims(r)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, err)
		return nil
	}

	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	environmentID, err := getEnvironmentIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	if err := h.accessControl.RequireEditor(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	var req CreateEnvironmentVariableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	if err := utils.Validate(req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	envVar, err := h.queries.CreateEnvironmentVariable(ctx, database.CreateEnvironmentVariableParams{
		EnvironmentID: environmentID,
		Key:           strings.ToUpper(req.Key),
		Value:         req.Value,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		CreatedAt:     sql.NullTime{Time: now, Valid: true},
		UpdatedAt:     sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	response := EnvironmentVariableResponse{
		ID:            envVar.ID,
		EnvironmentID: envVar.EnvironmentID,
		Key:           envVar.Key,
		Value:         envVar.Value,
		Description:   nullStringToStringPtr(envVar.Description),
		CreatedAt:     envVar.CreatedAt.Time,
		UpdatedAt:     envVar.UpdatedAt.Time,
	}

	return sendJSONResponse(w, http.StatusCreated, response)
}

func (h *EnvironmentVariableHandlerImpl) GetEnvironmentVariable(w http.ResponseWriter, r *http.Request) error {
	claims, err := getUserClaims(r)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, err)
		return nil
	}

	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	envVarID, err := getEnvironmentVariableIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	if err := h.accessControl.RequireViewer(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	envVar, err := h.queries.GetAccessibleEnvironmentVariable(ctx, database.GetAccessibleEnvironmentVariableParams{
		ID:      envVarID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		sendErrorResponse(w, http.StatusNotFound, fmt.Errorf("environment variable not found"))
		return nil
	} else if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	response := EnvironmentVariableResponse{
		ID:            envVar.ID,
		EnvironmentID: envVar.EnvironmentID,
		Key:           envVar.Key,
		Value:         envVar.Value,
		Description:   nullStringToStringPtr(envVar.Description),
		CreatedAt:     envVar.CreatedAt.Time,
		UpdatedAt:     envVar.UpdatedAt.Time,
	}

	return sendJSONResponse(w, http.StatusOK, response)
}

func (h *EnvironmentVariableHandlerImpl) ListEnvironmentVariables(w http.ResponseWriter, r *http.Request) error {
	claims, err := getUserClaims(r)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, err)
		return nil
	}

	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	environmentID, err := getEnvironmentIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	if err := h.accessControl.RequireViewer(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	envVars, err := h.queries.ListEnvironmentVariablesByEnvironment(ctx, environmentID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	var response []EnvironmentVariableResponse
	for _, envVar := range envVars {
		response = append(response, EnvironmentVariableResponse{
			ID:            envVar.ID,
			EnvironmentID: envVar.EnvironmentID,
			Key:           envVar.Key,
			Value:         envVar.Value,
			Description:   nullStringToStringPtr(envVar.Description),
			CreatedAt:     envVar.CreatedAt.Time,
			UpdatedAt:     envVar.UpdatedAt.Time,
		})
	}

	return sendJSONResponse(w, http.StatusOK, response)
}

func (h *EnvironmentVariableHandlerImpl) UpdateEnvironmentVariable(w http.ResponseWriter, r *http.Request) error {
	claims, err := getUserClaims(r)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, err)
		return nil
	}

	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	envVarID, err := getEnvironmentVariableIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	if err := h.accessControl.RequireEditor(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	var req UpdateEnvironmentVariableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	if err := utils.Validate(req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	envVar, err := h.queries.UpdateEnvironmentVariable(ctx, database.UpdateEnvironmentVariableParams{
		Key:         strings.ToUpper(req.Key),
		Value:       req.Value,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		ID:          envVarID,
	})
	if err == sql.ErrNoRows {
		sendErrorResponse(w, http.StatusNotFound, fmt.Errorf("environment variable not found"))
		return nil
	} else if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	response := EnvironmentVariableResponse{
		ID:            envVar.ID,
		EnvironmentID: envVar.EnvironmentID,
		Key:           envVar.Key,
		Value:         envVar.Value,
		Description:   nullStringToStringPtr(envVar.Description),
		CreatedAt:     envVar.CreatedAt.Time,
		UpdatedAt:     envVar.UpdatedAt.Time,
	}

	return sendJSONResponse(w, http.StatusOK, response)
}

func (h *EnvironmentVariableHandlerImpl) DeleteEnvironmentVariable(w http.ResponseWriter, r *http.Request) error {
	claims, err := getUserClaims(r)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, err)
		return nil
	}

	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	envVarID, err := getEnvironmentVariableIDFromPath(r)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err)
		return nil
	}

	if err := h.accessControl.RequireEditor(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.queries.DeleteEnvironmentVariable(ctx, envVarID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	return sendJSONResponse(w, http.StatusOK, map[string]string{"message": "Environment variable deleted successfully"})
}

func getEnvironmentVariableIDFromPath(r *http.Request) (int64, error) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 6 || pathParts[5] == "" {
		return 0, fmt.Errorf("environment variable ID is required")
	}
	return strconv.ParseInt(pathParts[5], 10, 64)
}
