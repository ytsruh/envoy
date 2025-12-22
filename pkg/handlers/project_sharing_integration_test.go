package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

func TestProjectSharingIntegration(t *testing.T) {
	mock := NewMockSharingQueries()

	ownerID := uuid.New().String()
	editorID := uuid.New().String()
	viewerID := uuid.New().String()

	// Create users in mock
	createTestUserInMock(t, mock, ownerID)
	createTestUserInMock(t, mock, editorID)
	createTestUserInMock(t, mock, viewerID)

	// Import the database types
	var _ database.Project
	var _ utils.JWTClaims

	// Create project as owner
	project := createTestProject(t, mock, ownerID)

	// Test sharing functionality
	projectHandler := NewProjectHandler(mock)
	sharingHandler := NewProjectSharingHandler(mock)

	t.Run("owner can add editor to project", func(t *testing.T) {
		reqBody := AddUserRequest{
			UserID: editorID,
			Role:   "editor",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/projects/1/members", bytes.NewReader(body))
		claims := createSharingTestUser(ownerID)
		req = req.WithContext(context.WithValue(req.Context(), "user", claims))
		w := httptest.NewRecorder()

		err := sharingHandler.AddUserToProject(w, req)
		if err != nil {
			t.Fatalf("AddUserToProject returned error: %v", err)
		}

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response ProjectUserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.UserID != editorID {
			t.Errorf("Expected user ID %s, got %s", editorID, response.UserID)
		}
		if response.Role != "editor" {
			t.Errorf("Expected role editor, got %s", response.Role)
		}
	})

	t.Run("editor can update project", func(t *testing.T) {
		reqBody := UpdateProjectRequest{
			Name:        "Updated by Editor",
			Description: "This project was updated by an editor",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/projects/1", bytes.NewReader(body))
		claims := createSharingTestUser(editorID)
		req = req.WithContext(context.WithValue(req.Context(), "user", claims))
		w := httptest.NewRecorder()

		err := projectHandler.UpdateProject(w, req)
		if err != nil {
			t.Fatalf("UpdateProject returned error: %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response ProjectResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Name != "Updated by Editor" {
			t.Errorf("Expected name 'Updated by Editor', got '%s'", response.Name)
		}
		if *response.Description != "This project was updated by an editor" {
			t.Errorf("Expected description 'This project was updated by an editor', got '%s'", *response.Description)
		}
	})

	t.Run("owner can add viewer to project", func(t *testing.T) {
		reqBody := AddUserRequest{
			UserID: viewerID,
			Role:   "viewer",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/projects/1/members", bytes.NewReader(body))
		claims := createSharingTestUser(ownerID)
		req = req.WithContext(context.WithValue(req.Context(), "user", claims))
		w := httptest.NewRecorder()

		err := sharingHandler.AddUserToProject(w, req)
		if err != nil {
			t.Fatalf("AddUserToProject returned error: %v", err)
		}

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("viewer cannot update project", func(t *testing.T) {
		reqBody := UpdateProjectRequest{
			Name:        "Updated by Viewer",
			Description: "This project was updated by a viewer",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/projects/1", bytes.NewReader(body))
		claims := createSharingTestUser(viewerID)
		req = req.WithContext(context.WithValue(req.Context(), "user", claims))
		w := httptest.NewRecorder()

		err := projectHandler.UpdateProject(w, req)
		if err != nil {
			t.Fatalf("UpdateProject returned error: %v", err)
		}

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("editor can list all their accessible projects", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/user/projects", nil)
		claims := createSharingTestUser(editorID)
		req = req.WithContext(context.WithValue(req.Context(), "user", claims))
		w := httptest.NewRecorder()

		err := sharingHandler.ListUserProjects(w, req)
		if err != nil {
			t.Fatalf("ListUserProjects returned error: %v", err)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response []UserProjectResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(response) != 1 {
			t.Errorf("Expected 1 project accessible to editor, got %d", len(response))
		}

		if response[0].ID != project.ID {
			t.Errorf("Expected project ID %d, got %d", project.ID, response[0].ID)
		}
	})
}
