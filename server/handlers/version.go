package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/shared"
)

func Version(c echo.Context, _ *HandlerContext) error {
	return c.JSON(http.StatusOK, map[string]string{
		"version": shared.Version,
	})
}
