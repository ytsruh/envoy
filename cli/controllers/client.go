package controllers

import (
	"ytsruh.com/envoy/cli/config"
	shared "ytsruh.com/envoy/shared"
)

type Client struct {
	*AuthController
	*ProjectsController
	*EnvironmentsController
	*VariablesController
}

func NewClient() (*Client, error) {
	serverURL, err := config.GetServerURL()
	if err != nil {
		return nil, err
	}

	token, err := config.GetToken()
	if err != nil {
		return nil, err
	}

	base := NewBaseClient(serverURL, token)

	return &Client{
		AuthController:         NewAuthController(base),
		ProjectsController:     NewProjectsController(base),
		EnvironmentsController: NewEnvironmentsController(base),
		VariablesController:    NewVariablesController(base),
	}, nil
}

func RequireToken() (*Client, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	if client.AuthController.token == "" {
		return nil, shared.ErrNoToken
	}

	return client, nil
}
