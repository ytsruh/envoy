package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/server/database/generated"
	"ytsruh.com/envoy/server/utils"
)

func TestListProjects_Simple(t *testing.T) {
	t.Parallel()

	user := CreateTestUser()

	mock := &MockQueries{
		projects: []database.Project{
			{
				ID:        "550e8400-e29b-41d4-a716-446655440001",
				Name:      "Project 1",
				OwnerID:   user.UserID,
				CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt: time.Now(),
			},
			{
				ID:        "550e8400-e29b-41d4-a716-446655440002",
				Name:      "Project 2",
				OwnerID:   user.UserID,
				CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt: time.Now(),
			},
		},
		users: make(map[string]database.User),
	}

	accessControl := utils.NewAccessControlService(mock)
	ctx := NewHandlerContext(mock, "test-secret", accessControl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", user)

	err := ListProjects(c, ctx)
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []ProjectResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(resp))
	}
}

func TestUpdateProject_Simple(t *testing.T) {
	t.Parallel()

	user := CreateTestUser()

	mock := &MockQueries{
		projects: []database.Project{
			{
				ID:        "550e8400-e29b-41d4-a716-446655440001",
				Name:      "Original Name",
				OwnerID:   user.UserID,
				CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt: time.Now(),
			},
		},
		users: make(map[string]database.User),
	}

	accessControl := utils.NewAccessControlService(mock)
	ctx := NewHandlerContext(mock, "test-secret", accessControl)

	reqBody := UpdateProjectRequest{
		Name:        "Updated Name",
		Description: "Updated description",
	}
	body, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/projects/550e8400-e29b-41d4-a716-446655440001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440001")
	c.Set("user", user)

	err := UpdateProject(c, ctx)
	if err != nil {
		t.Fatalf("UpdateProject returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp ProjectResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", resp.Name)
	}
}

func TestDeleteProject_Simple(t *testing.T) {
	t.Parallel()

	user := CreateTestUser()

	mock := &MockQueries{
		projects: []database.Project{
			{
				ID:        "550e8400-e29b-41d4-a716-446655440001",
				Name:      "To Delete",
				OwnerID:   user.UserID,
				CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt: time.Now(),
			},
		},
		users: make(map[string]database.User),
	}

	accessControl := utils.NewAccessControlService(mock)
	ctx := NewHandlerContext(mock, "test-secret", accessControl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/projects/550e8400-e29b-41d4-a716-446655440001", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("550e8400-e29b-41d4-a716-446655440001")
	c.Set("user", user)

	err := DeleteProject(c, ctx)
	if err != nil {
		t.Fatalf("DeleteProject returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp["message"] != "Project deleted successfully" {
		t.Errorf("Expected success message, got '%s'", resp["message"])
	}
}
