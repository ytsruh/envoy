package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
)

type HealthHandlerImpl struct {
	queries database.Querier
}

func NewHealthHandler(queries database.Querier) HealthHandler {
	return &HealthHandlerImpl{queries: queries}
}

func (h *HealthHandlerImpl) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"health": "ok"})
}
