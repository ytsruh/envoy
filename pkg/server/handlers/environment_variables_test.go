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
	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/utils"
)

func TestCreateEnvironmentVariable(t *testing.T) {
	t.Parallel()

	user := CreateTestUser()

	mock := &MockQueries{
		projects: []database.Project{
			{
				ID:        1,
				Name:      "Test Project",
				OwnerID:   user.UserID,
				CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt: time.Now(),
			},
		},
		users: make(map[string]database.User),
	}

	accessControl := utils.NewAccessControlService(mock)
	ctx := NewHandlerContext(mock, "test-secret", accessControl)

	reqBody := CreateEnvironmentVariableRequest{
		Key:         "API_KEY",
		Value:       "secret-value",
		Description: "API key for external service",
	}
	body, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/projects/1/environments/1/variables", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("project_id", "environment_id")
	c.SetParamValues("1", "1")
	c.Set("user", user)

	err := CreateEnvironmentVariable(c, ctx)
	if err != nil {
		t.Fatalf("CreateEnvironmentVariable returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestListEnvironmentVariables(t *testing.T) {
	t.Parallel()

	user := CreateTestUser()

	mock := &MockQueries{
		projects: []database.Project{
			{
				ID:        1,
				Name:      "Test Project",
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
	req := httptest.NewRequest(http.MethodGet, "/projects/1/environments/1/variables", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("project_id", "environment_id")
	c.SetParamValues("1", "1")
	c.Set("user", user)

	err := ListEnvironmentVariables(c, ctx)
	if err != nil {
		t.Fatalf("ListEnvironmentVariables returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []EnvironmentVariableResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
}

func TestUpdateEnvironmentVariable(t *testing.T) {
	t.Parallel()

	user := CreateTestUser()

	mock := &MockQueries{
		projects: []database.Project{
			{
				ID:        1,
				Name:      "Test Project",
				OwnerID:   user.UserID,
				CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt: time.Now(),
			},
		},
		users: make(map[string]database.User),
	}

	accessControl := utils.NewAccessControlService(mock)
	ctx := NewHandlerContext(mock, "test-secret", accessControl)

	reqBody := UpdateEnvironmentVariableRequest{
		Key:         "UPDATED_KEY",
		Value:       "new-value",
		Description: "Updated description",
	}
	body, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/projects/1/environments/1/variables/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("project_id", "id")
	c.SetParamValues("1", "1")
	c.Set("user", user)

	err := UpdateEnvironmentVariable(c, ctx)
	if err != nil {
		t.Fatalf("UpdateEnvironmentVariable returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestDeleteEnvironmentVariable(t *testing.T) {
	t.Parallel()

	user := CreateTestUser()

	mock := &MockQueries{
		projects: []database.Project{
			{
				ID:        1,
				Name:      "Test Project",
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
	req := httptest.NewRequest(http.MethodDelete, "/projects/1/environments/1/variables/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("project_id", "id")
	c.SetParamValues("1", "1")
	c.Set("user", user)

	err := DeleteEnvironmentVariable(c, ctx)
	if err != nil {
		t.Fatalf("DeleteEnvironmentVariable returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["message"] != "Environment variable deleted successfully" {
		t.Errorf("Expected success message, got '%s'", resp["message"])
	}
}
