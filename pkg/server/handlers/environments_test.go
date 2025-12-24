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

func TestCreateEnvironment(t *testing.T) {
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

	reqBody := CreateEnvironmentRequest{
		Name:        "Production",
		Description: "Production environment",
	}
	body, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/projects/1/environments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("project_id")
	c.SetParamValues("1")
	c.Set("user", user)

	err := CreateEnvironment(c, ctx)
	if err != nil {
		t.Fatalf("CreateEnvironment returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestListEnvironments(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodGet, "/projects/1/environments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("project_id")
	c.SetParamValues("1")
	c.Set("user", user)

	err := ListEnvironments(c, ctx)
	if err != nil {
		t.Fatalf("ListEnvironments returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []EnvironmentResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
}

func TestUpdateEnvironment(t *testing.T) {
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

	reqBody := UpdateEnvironmentRequest{
		Name:        "Staging",
		Description: "Updated description",
	}
	body, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/projects/1/environments/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user", user)

	err := UpdateEnvironment(c, ctx)
	if err != nil {
		t.Fatalf("UpdateEnvironment returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestDeleteEnvironment(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodDelete, "/projects/1/environments/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user", user)

	err := DeleteEnvironment(c, ctx)
	if err != nil {
		t.Fatalf("DeleteEnvironment returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["message"] != "Environment deleted successfully" {
		t.Errorf("Expected success message, got '%s'", resp["message"])
	}
}
