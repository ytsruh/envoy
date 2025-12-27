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

func TestAddUserToProject(t *testing.T) {
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
		users: map[string]database.User{
			"user-to-add": {
				ID:    "user-to-add",
				Name:  "Other User",
				Email: "other@example.com",
			},
		},
	}

	accessControl := utils.NewAccessControlService(mock)
	ctx := NewHandlerContext(mock, "test-secret", accessControl)

	reqBody := AddUserRequest{
		UserID: "user-to-add",
		Role:   "viewer",
	}
	body, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/projects/1/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user", user)

	err := AddUserToProject(c, ctx)
	if err != nil {
		t.Fatalf("AddUserToProject returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestRemoveUserFromProject(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodDelete, "/projects/1/members/user-to-remove", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "user_id")
	c.SetParamValues("1", "user-to-remove")
	c.Set("user", user)

	err := RemoveUserFromProject(c, ctx)
	if err != nil {
		t.Fatalf("RemoveUserFromProject returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["message"] != "User removed successfully" {
		t.Errorf("Expected success message, got '%s'", resp["message"])
	}
}

func TestUpdateUserRole(t *testing.T) {
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

	reqBody := UpdateRoleRequest{
		Role: "editor",
	}
	body, _ := json.Marshal(reqBody)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/projects/1/members/user-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "user_id")
	c.SetParamValues("1", "user-id")
	c.Set("user", user)

	err := UpdateUserRole(c, ctx)
	if err != nil {
		t.Fatalf("UpdateUserRole returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetProjectUsers(t *testing.T) {
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
	req := httptest.NewRequest(http.MethodGet, "/projects/1/members", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user", user)

	err := GetProjectUsers(c, ctx)
	if err != nil {
		t.Fatalf("GetProjectUsers returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []ProjectUserResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
}

func TestListUserProjects(t *testing.T) {
	t.Parallel()

	user := CreateTestUser()

	mock := &MockQueries{
		projects: []database.Project{
			{
				ID:        1,
				Name:      "Project 1",
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
	req := httptest.NewRequest(http.MethodGet, "/user/projects", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", user)

	err := ListUserProjects(c, ctx)
	if err != nil {
		t.Fatalf("ListUserProjects returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp []UserProjectResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Errorf("Expected 1 project, got %d", len(resp))
	}
}
