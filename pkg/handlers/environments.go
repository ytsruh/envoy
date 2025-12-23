package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

type EnvironmentHandler interface {
	CreateEnvironment(w http.ResponseWriter, r *http.Request) error
	GetEnvironment(w http.ResponseWriter, r *http.Request) error
	ListEnvironments(w http.ResponseWriter, r *http.Request) error
	UpdateEnvironment(w http.ResponseWriter, r *http.Request) error
	DeleteEnvironment(w http.ResponseWriter, r *http.Request) error
}

type EnvironmentHandlerImpl struct {
	queries       database.Querier
	accessControl utils.AccessControlService
}

func NewEnvironmentHandler(queries database.Querier, accessControl utils.AccessControlService) EnvironmentHandler {
	return &EnvironmentHandlerImpl{
		queries:       queries,
		accessControl: accessControl,
	}
}

type CreateEnvironmentRequest struct {
	Name        string `json:"name" validate:"required,environment_name"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateEnvironmentRequest struct {
	Name        string `json:"name" validate:"required,environment_name"`
	Description string `json:"description" validate:"max=500"`
}

type EnvironmentResponse struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *EnvironmentHandlerImpl) CreateEnvironment(w http.ResponseWriter, r *http.Request) error {
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

	// Check if user has editor access to the project
	if err := h.accessControl.RequireEditor(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	var req CreateEnvironmentRequest
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
	environment, err := h.queries.CreateEnvironment(ctx, database.CreateEnvironmentParams{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		CreatedAt:   sql.NullTime{Time: now, Valid: true},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	response := EnvironmentResponse{
		ID:          environment.ID,
		ProjectID:   environment.ProjectID,
		Name:        environment.Name,
		Description: nullStringToStringPtr(environment.Description),
		CreatedAt:   environment.CreatedAt.Time,
		UpdatedAt:   environment.UpdatedAt.Time,
	}

	return sendJSONResponse(w, http.StatusCreated, response)
}

func (h *EnvironmentHandlerImpl) GetEnvironment(w http.ResponseWriter, r *http.Request) error {
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

	// Check if user has viewer access to the project
	if err := h.accessControl.RequireViewer(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	environment, err := h.queries.GetAccessibleEnvironment(ctx, database.GetAccessibleEnvironmentParams{
		ID:      environmentID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		sendErrorResponse(w, http.StatusNotFound, err)
		return nil
	} else if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	response := EnvironmentResponse{
		ID:          environment.ID,
		ProjectID:   environment.ProjectID,
		Name:        environment.Name,
		Description: nullStringToStringPtr(environment.Description),
		CreatedAt:   environment.CreatedAt.Time,
		UpdatedAt:   environment.UpdatedAt.Time,
	}

	return sendJSONResponse(w, http.StatusOK, response)
}

func (h *EnvironmentHandlerImpl) ListEnvironments(w http.ResponseWriter, r *http.Request) error {
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

	// Check if user has viewer access to the project
	if err := h.accessControl.RequireViewer(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	environments, err := h.queries.ListEnvironmentsByProject(ctx, projectID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	var response []EnvironmentResponse
	for _, environment := range environments {
		response = append(response, EnvironmentResponse{
			ID:          environment.ID,
			ProjectID:   environment.ProjectID,
			Name:        environment.Name,
			Description: nullStringToStringPtr(environment.Description),
			CreatedAt:   environment.CreatedAt.Time,
			UpdatedAt:   environment.UpdatedAt.Time,
		})
	}

	return sendJSONResponse(w, http.StatusOK, response)
}

func (h *EnvironmentHandlerImpl) UpdateEnvironment(w http.ResponseWriter, r *http.Request) error {
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

	// Check if user has editor access to the project
	if err := h.accessControl.RequireEditor(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	var req UpdateEnvironmentRequest
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
	environment, err := h.queries.UpdateEnvironment(ctx, database.UpdateEnvironmentParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		ID:          environmentID,
	})
	if err == sql.ErrNoRows {
		sendErrorResponse(w, http.StatusNotFound, err)
		return nil
	} else if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	response := EnvironmentResponse{
		ID:          environment.ID,
		ProjectID:   environment.ProjectID,
		Name:        environment.Name,
		Description: nullStringToStringPtr(environment.Description),
		CreatedAt:   environment.CreatedAt.Time,
		UpdatedAt:   environment.UpdatedAt.Time,
	}

	return sendJSONResponse(w, http.StatusOK, response)
}

func (h *EnvironmentHandlerImpl) DeleteEnvironment(w http.ResponseWriter, r *http.Request) error {
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

	// Check if user has editor access to the project
	if err := h.accessControl.RequireEditor(r.Context(), projectID, claims.UserID); err != nil {
		sendErrorResponse(w, http.StatusForbidden, err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.queries.DeleteEnvironment(ctx, database.DeleteEnvironmentParams{
		DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        environmentID,
	})
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err)
		return nil
	}

	return sendJSONResponse(w, http.StatusOK, map[string]string{"message": "Environment deleted successfully"})
}
