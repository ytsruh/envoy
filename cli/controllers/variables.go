package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	shared "ytsruh.com/envoy/shared"
)

type VariablesController struct {
	*BaseClient
}

func NewVariablesController(base *BaseClient) *VariablesController {
	return &VariablesController{BaseClient: base}
}

type EnvironmentVariableResponse struct {
	ID            shared.EnvironmentVariableID `json:"id"`
	EnvironmentID shared.EnvironmentID         `json:"environment_id"`
	Key           string                       `json:"key"`
	Value         string                       `json:"value"`
	Description   *string                      `json:"description"`
	CreatedAt     shared.Timestamp             `json:"created_at"`
	UpdatedAt     shared.Timestamp             `json:"updated_at"`
}

func (v *VariablesController) CreateEnvironmentVariable(projectID, environmentID int64, key, value string) (*EnvironmentVariableResponse, error) {
	reqBody := map[string]any{
		"key":   key,
		"value": value,
	}

	resp, err := v.doRequest("POST", fmt.Sprintf("/projects/%d/environments/%d/variables", projectID, environmentID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var varResp EnvironmentVariableResponse
	if err := v.decodeResponse(resp, &varResp); err != nil {
		return nil, err
	}

	return &varResp, nil
}

func (v *VariablesController) ListEnvironmentVariables(projectID, environmentID int64) ([]EnvironmentVariableResponse, error) {
	resp, err := v.doRequest("GET", fmt.Sprintf("/projects/%d/environments/%d/variables", projectID, environmentID), nil, true)
	if err != nil {
		return nil, err
	}

	var variables []EnvironmentVariableResponse
	if err := v.decodeResponse(resp, &variables); err != nil {
		return nil, err
	}

	return variables, nil
}

func (v *VariablesController) GetEnvironmentVariable(projectID, environmentID, variableID int64) (*EnvironmentVariableResponse, error) {
	resp, err := v.doRequest("GET", fmt.Sprintf("/projects/%d/environments/%d/variables/%d", projectID, environmentID, variableID), nil, true)
	if err != nil {
		return nil, err
	}

	var varResp EnvironmentVariableResponse
	if err := v.decodeResponse(resp, &varResp); err != nil {
		return nil, err
	}

	return &varResp, nil
}

func (v *VariablesController) UpdateEnvironmentVariable(projectID, environmentID, variableID int64, key, value string) (*EnvironmentVariableResponse, error) {
	reqBody := map[string]any{
		"key":   key,
		"value": value,
	}

	resp, err := v.doRequest("PUT", fmt.Sprintf("/projects/%d/environments/%d/variables/%d", projectID, environmentID, variableID), reqBody, true)
	if err != nil {
		return nil, err
	}

	var varResp EnvironmentVariableResponse
	if err := v.decodeResponse(resp, &varResp); err != nil {
		return nil, err
	}

	return &varResp, nil
}

func (v *VariablesController) DeleteEnvironmentVariable(projectID, environmentID, variableID int64) error {
	resp, err := v.doRequest("DELETE", fmt.Sprintf("/projects/%d/environments/%d/variables/%d", projectID, environmentID, variableID), nil, true)
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
