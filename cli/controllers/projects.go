package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	shared "ytsruh.com/envoy/shared"
)

type ProjectsController struct {
	*BaseClient
}

func NewProjectsController(base *BaseClient) *ProjectsController {
	return &ProjectsController{BaseClient: base}
}

type ProjectResponse struct {
	ID          shared.ProjectID `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	GitRepo     *string          `json:"git_repo"`
	OwnerID     shared.UserID    `json:"owner_id"`
	CreatedAt   shared.Timestamp `json:"created_at"`
	UpdatedAt   shared.Timestamp `json:"updated_at"`
}

func (p *ProjectsController) CreateProject(name, description, gitRepo string) (*ProjectResponse, error) {
	reqBody := map[string]any{
		"name": name,
	}
	if description != "" {
		reqBody["description"] = description
	} else {
		reqBody["description"] = nil
	}
	if gitRepo != "" {
		reqBody["git_repo"] = gitRepo
	} else {
		reqBody["git_repo"] = nil
	}

	resp, err := p.doRequest("POST", "/projects", reqBody, true)
	if err != nil {
		return nil, err
	}

	var projectResp ProjectResponse
	if err := p.decodeResponse(resp, &projectResp); err != nil {
		return nil, err
	}

	return &projectResp, nil
}

func (p *ProjectsController) ListProjects() ([]ProjectResponse, error) {
	resp, err := p.doRequest("GET", "/projects", nil, true)
	if err != nil {
		return nil, err
	}

	var projects []ProjectResponse
	if err := p.decodeResponse(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (p *ProjectsController) GetProject(projectID int64) (*ProjectResponse, error) {
	resp, err := p.doRequest("GET", fmt.Sprintf("/projects/%d", projectID), nil, true)
	if err != nil {
		return nil, err
	}

	var projectResp ProjectResponse
	if err := p.decodeResponse(resp, &projectResp); err != nil {
		return nil, err
	}

	return &projectResp, nil
}

func (p *ProjectsController) UpdateProject(projectID int64, name, description, gitRepo string) (*ProjectResponse, error) {
	reqBody := map[string]any{
		"name": name,
	}
	if description != "" {
		reqBody["description"] = description
	} else {
		reqBody["description"] = nil
	}
	if gitRepo != "" {
		reqBody["git_repo"] = gitRepo
	} else {
		reqBody["git_repo"] = nil
	}

	resp, err := p.doRequest("PUT", fmt.Sprintf("/projects/%d", projectID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var projectResp ProjectResponse
	if err := p.decodeResponse(resp, &projectResp); err != nil {
		return nil, err
	}

	return &projectResp, nil
}

func (p *ProjectsController) DeleteProject(projectID int64) error {
	resp, err := p.doRequest("DELETE", fmt.Sprintf("/api/projects/%d", projectID), nil, true)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
