package handlers

import (
	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/server/views"
)

func Home(c echo.Context, _ *HandlerContext) error {
	component := views.Home()
	return component.Render(c.Request().Context(), c.Response())
}
