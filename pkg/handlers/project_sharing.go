package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

type ProjectSharingHandler interface {
	AddUserToProject(w http.ResponseWriter, r *http.Request) error
	RemoveUserFromProject(w http.ResponseWriter, r *http.Request) error
	UpdateUserRole(w http.ResponseWriter, r *http.Request) error
	GetProjectUsers(w http.ResponseWriter, r *http.Request) error
	ListUserProjects(w http.ResponseWriter, r *http.Request) error
}

type ProjectSharingHandlerImpl struct {
	queries database.Querier
}

func NewProjectSharingHandler(queries database.Querier) ProjectSharingHandler {
	return &ProjectSharingHandlerImpl{
		queries: queries,
	}
}

type AddUserRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=viewer editor"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=viewer editor"`
}

type ProjectUserResponse struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type UserProjectResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func getUserIDFromPath(r *http.Request) (string, error) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[3] == "" {
		return "", httpError{code: http.StatusBadRequest, message: "user ID is required"}
	}
	return pathParts[3], nil
}

func (h *ProjectSharingHandlerImpl) AddUserToProject(w http.ResponseWriter, r *http.Request) error {
	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		return err
	}

	var req AddUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return nil
	}

	if err := utils.Validate(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return nil
	}

	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user is project owner
	ownerCount, err := h.queries.IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to check ownership"})
		return nil
	}
	if ownerCount == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Only project owners can add users"})
		return nil
	}

	// Check if user exists
	_, err = h.queries.GetUser(ctx, req.UserID)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found"})
		return nil
	} else if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to verify user"})
		return nil
	}

	// Check if user is already a member
	_, err = h.queries.GetProjectMembership(ctx, database.GetProjectMembershipParams{
		ProjectID: projectID,
		UserID:    req.UserID,
	})
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User is already a member of this project"})
		return nil
	} else if err != sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to check membership"})
		return nil
	}

	now := time.Now()
	membership, err := h.queries.AddUserToProject(ctx, database.AddUserToProjectParams{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
		CreatedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: now,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to add user to project"})
		return nil
	}

	response := ProjectUserResponse{
		UserID:    membership.UserID,
		Role:      membership.Role,
		CreatedAt: membership.CreatedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(response)
}

func (h *ProjectSharingHandlerImpl) RemoveUserFromProject(w http.ResponseWriter, r *http.Request) error {
	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		return err
	}

	targetUserID, err := getUserIDFromPath(r)
	if err != nil {
		return err
	}

	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user is project owner
	ownerCount, err := h.queries.IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to check ownership"})
		return nil
	}
	if ownerCount == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Only project owners can remove users"})
		return nil
	}

	// Check if target user is a member
	_, err = h.queries.GetProjectMembership(ctx, database.GetProjectMembershipParams{
		ProjectID: projectID,
		UserID:    targetUserID,
	})
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User is not a member of this project"})
		return nil
	} else if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to check membership"})
		return nil
	}

	err = h.queries.RemoveUserFromProject(ctx, database.RemoveUserFromProjectParams{
		ProjectID: projectID,
		UserID:    targetUserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to remove user from project"})
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"message": "User removed from project successfully"})
}

func (h *ProjectSharingHandlerImpl) UpdateUserRole(w http.ResponseWriter, r *http.Request) error {
	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		return err
	}

	targetUserID, err := getUserIDFromPath(r)
	if err != nil {
		return err
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return nil
	}

	if err := utils.Validate(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return nil
	}

	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user is project owner
	ownerCount, err := h.queries.IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: claims.UserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to check ownership"})
		return nil
	}
	if ownerCount == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Only project owners can update user roles"})
		return nil
	}

	// Check if target user is a member
	membership, err := h.queries.GetProjectMembership(ctx, database.GetProjectMembershipParams{
		ProjectID: projectID,
		UserID:    targetUserID,
	})
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User is not a member of this project"})
		return nil
	} else if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to check membership"})
		return nil
	}

	// Don't allow changing role if it's the same
	if membership.Role == req.Role {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User already has this role"})
		return nil
	}

	err = h.queries.UpdateUserRole(ctx, database.UpdateUserRoleParams{
		Role:      req.Role,
		UpdatedAt: time.Now(),
		ProjectID: projectID,
		UserID:    targetUserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update user role"})
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"message": "User role updated successfully"})
}

func (h *ProjectSharingHandlerImpl) GetProjectUsers(w http.ResponseWriter, r *http.Request) error {
	projectID, err := getProjectIDFromPath(r)
	if err != nil {
		return err
	}

	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user has access to the project
	_, err = h.queries.GetAccessibleProject(ctx, database.GetAccessibleProjectParams{
		ID:      projectID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Project not found or access denied"})
		return nil
	} else if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch project"})
		return nil
	}

	projectUsers, err := h.queries.GetProjectUsers(ctx, projectID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch project users"})
		return nil
	}

	var response []ProjectUserResponse
	for _, user := range projectUsers {
		response = append(response, ProjectUserResponse{
			UserID:    user.UserID,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Time,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func (h *ProjectSharingHandlerImpl) ListUserProjects(w http.ResponseWriter, r *http.Request) error {
	user := r.Context().Value("user")
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return nil
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse user claims"})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	projects, err := h.queries.GetUserProjects(ctx, database.GetUserProjectsParams{
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch projects"})
		return nil
	}

	var response []UserProjectResponse
	for _, project := range projects {
		response = append(response, UserProjectResponse{
			ID:          project.ID,
			Name:        project.Name,
			Description: nullStringToStringPtr(project.Description),
			OwnerID:     project.OwnerID,
			CreatedAt:   project.CreatedAt.Time,
			UpdatedAt:   project.UpdatedAt.(time.Time),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

// Helper type for returning HTTP errors from helper functions
type httpError struct {
	code    int
	message string
}

func (e httpError) Error() string {
	return e.message
}
