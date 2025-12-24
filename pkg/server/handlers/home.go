package handlers

import (
	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/pkg/templates"
)

func Home(c echo.Context, _ *HandlerContext) error {
	component := templates.Home()
	return component.Render(c.Request().Context(), c.Response())
}
