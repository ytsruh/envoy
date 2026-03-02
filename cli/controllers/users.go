package controllers

import (
	"fmt"
	"net/url"

	shared "ytsruh.com/envoy/shared"
)

type UsersController struct {
	*BaseClient
}

func NewUsersController(base *BaseClient) *UsersController {
	return &UsersController{BaseClient: base}
}

func (u *UsersController) SearchByEmail(email string) ([]shared.UserSearchResponse, error) {
	queryParams := url.Values{}
	queryParams.Add("email", email)

	resp, err := u.doRequest("GET", "/users/search?"+queryParams.Encode(), nil, true)
	if err != nil {
		return nil, err
	}

	var users []shared.UserSearchResponse
	if err := u.decodeResponse(resp, &users); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return users, nil
}
