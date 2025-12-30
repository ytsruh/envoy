package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	shared "ytsruh.com/envoy/shared"
)

type EnvironmentsController struct {
	*BaseClient
}

func NewEnvironmentsController(base *BaseClient) *EnvironmentsController {
	return &EnvironmentsController{BaseClient: base}
}

type EnvironmentResponse struct {
	ID          shared.ProjectID `json:"id"`
	ProjectID   shared.ProjectID `json:"project_id"`
	Name        string           `json:"name"`
	Description *string          `json:"description"`
	CreatedAt   shared.Timestamp `json:"created_at"`
	UpdatedAt   shared.Timestamp `json:"updated_at"`
}

func (e *EnvironmentsController) CreateEnvironment(projectID int64, name, description string) (*EnvironmentResponse, error) {
	reqBody := map[string]any{
		"name": name,
	}
	if description != "" {
		reqBody["description"] = description
	} else {
		reqBody["description"] = nil
	}

	resp, err := e.doRequest("POST", fmt.Sprintf("/projects/%d/environments", projectID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var envResp EnvironmentResponse
	if err := e.decodeResponse(resp, &envResp); err != nil {
		return nil, err
	}

	return &envResp, nil
}

func (e *EnvironmentsController) ListEnvironments(projectID int64) ([]EnvironmentResponse, error) {
	resp, err := e.doRequest("GET", fmt.Sprintf("/projects/%d/environments", projectID), nil, true)
	if err != nil {
		return nil, err
	}

	var environments []EnvironmentResponse
	if err := e.decodeResponse(resp, &environments); err != nil {
		return nil, err
	}

	return environments, nil
}

func (e *EnvironmentsController) GetEnvironment(projectID, environmentID int64) (*EnvironmentResponse, error) {
	resp, err := e.doRequest("GET", fmt.Sprintf("/projects/%d/environments/%d", projectID, environmentID), nil, true)
	if err != nil {
		return nil, err
	}

	var envResp EnvironmentResponse
	if err := e.decodeResponse(resp, &envResp); err != nil {
		return nil, err
	}

	return &envResp, nil
}

func (e *EnvironmentsController) UpdateEnvironment(projectID, environmentID int64, name, description string) (*EnvironmentResponse, error) {
	reqBody := map[string]any{
		"name": name,
	}
	if description != "" {
		reqBody["description"] = description
	} else {
		reqBody["description"] = nil
	}

	resp, err := e.doRequest("PUT", fmt.Sprintf("/projects/%d/environments/%d", projectID, environmentID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var envResp EnvironmentResponse
	if err := e.decodeResponse(resp, &envResp); err != nil {
		return nil, err
	}

	return &envResp, nil
}

func (e *EnvironmentsController) DeleteEnvironment(projectID, environmentID int64) error {
	resp, err := e.doRequest("DELETE", fmt.Sprintf("/projects/%d/environments/%d", projectID, environmentID), nil, true)
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
