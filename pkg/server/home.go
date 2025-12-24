package server

import (
	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/pkg/templates"
)

func (s *Server) Home(c echo.Context) error {
	component := templates.Home()
	return component.Render(c.Request().Context(), c.Response())
}
