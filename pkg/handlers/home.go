package handlers

import (
	"net/http"

	"ytsruh.com/envoy/pkg/templates"
)

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

func (h *HomeHandler) Home(w http.ResponseWriter, r *http.Request) error {
	component := templates.Home()
	return component.Render(r.Context(), w)
}
