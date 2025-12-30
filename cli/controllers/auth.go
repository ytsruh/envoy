package controllers

import (
	"ytsruh.com/envoy/cli/config"
	shared "ytsruh.com/envoy/shared"
)

type AuthController struct {
	*BaseClient
}

func NewAuthController(base *BaseClient) *AuthController {
	return &AuthController{BaseClient: base}
}

type AuthResponse struct {
	Token string `json:"token"`
	User  shared.RegisterResponse
}

type ProfileResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Iat    int64  `json:"issued_at"`
	Exp    int64  `json:"expires_at"`
}

func (a *AuthController) Register(name, email, password string) (*AuthResponse, error) {
	reqBody := shared.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}

	resp, err := a.doRequest("POST", "/auth/register", reqBody, false)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := a.decodeResponse(resp, &authResp); err != nil {
		return nil, err
	}

	if err := config.SetToken(authResp.Token); err != nil {
		return nil, err
	}

	a.SetToken(authResp.Token)

	return &authResp, nil
}

func (a *AuthController) Login(email, password string) (*AuthResponse, error) {
	reqBody := shared.LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := a.doRequest("POST", "/auth/login", reqBody, false)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := a.decodeResponse(resp, &authResp); err != nil {
		return nil, err
	}

	if err := config.SetToken(authResp.Token); err != nil {
		return nil, err
	}

	a.SetToken(authResp.Token)

	return &authResp, nil
}

func (a *AuthController) GetProfile() (*ProfileResponse, error) {
	resp, err := a.doRequest("GET", "/auth/profile", nil, true)
	if err != nil {
		return nil, err
	}

	var profileResp ProfileResponse
	if err := a.decodeResponse(resp, &profileResp); err != nil {
		return nil, err
	}

	return &profileResp, nil
}
