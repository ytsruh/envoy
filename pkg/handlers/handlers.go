package handlers

import (
	"github.com/labstack/echo/v4"
)

type HealthHandler interface {
	Health(c echo.Context) error
}

type GreetingHandler interface {
	Hello(c echo.Context) error
	Goodbye(c echo.Context) error
}
