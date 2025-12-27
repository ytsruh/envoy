package handlers

import (
	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/server/database/generated"
	"ytsruh.com/envoy/server/utils"
)

type HandlerContext struct {
	Queries       database.Querier
	JWTSecret     string
	AccessControl utils.AccessControlService
}

func NewHandlerContext(queries database.Querier, jwtSecret string, accessControl utils.AccessControlService) *HandlerContext {
	return &HandlerContext{
		Queries:       queries,
		JWTSecret:     jwtSecret,
		AccessControl: accessControl,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func SendErrorResponse(c echo.Context, code int, err error) error {
	return c.JSON(code, ErrorResponse{Error: err.Error()})
}

func NullStringToStringPtr(ns any) *string {
	if ns == nil {
		return nil
	}
	switch v := ns.(type) {
	case string:
		return &v
	default:
		return nil
	}
}
