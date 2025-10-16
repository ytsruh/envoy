package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/pkg/database/generated"
)

type GreetingHandlerImpl struct {
	queries database.Querier
}

func NewGreetingHandler(queries database.Querier) GreetingHandler {
	return &GreetingHandlerImpl{queries: queries}
}

func (h *GreetingHandlerImpl) Hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func (h *GreetingHandlerImpl) Goodbye(c echo.Context) error {
	return c.String(http.StatusOK, "Goodbye, World!")
}
