package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Health(c echo.Context, _ *HandlerContext) error {
	return c.JSON(http.StatusOK, map[string]string{"health": "ok"})
}
