package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ytsruh.com/envoy/cli/utils"
	"ytsruh.com/envoy/shared"
)

type BaseClient struct {
	serverURL string
	token     string
	client    *http.Client
}

func NewBaseClient(serverURL, token string) *BaseClient {
	return &BaseClient{
		serverURL: serverURL,
		token:     token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (b *BaseClient) SetToken(token string) {
	b.token = token
}

func (b *BaseClient) buildURL(path string) string {
	return b.serverURL + path
}

func (b *BaseClient) doRequest(method, path string, body interface{}, authRequired bool) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, b.buildURL(path), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if authRequired && b.token != "" {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized && errResp.Error == "Token has expired" {
				if err := utils.ClearToken(); err == nil {
					b.token = ""
				}
				return resp, shared.ErrExpiredToken
			}

			return resp, fmt.Errorf("server error: %s", errResp.Error)
		}
	}

	return resp, nil
}

func (b *BaseClient) decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

type ErrorResponse struct {
	Error string `json:"error"`
}
