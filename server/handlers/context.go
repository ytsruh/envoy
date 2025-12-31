package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	database "ytsruh.com/envoy/server/database/generated"
	"ytsruh.com/envoy/server/middleware"
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

func BindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}
	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}
	return nil
}

func GetUserOrUnauthorized(c echo.Context) (*utils.JWTClaims, error) {
	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return nil, SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}
	return claims, nil
}

func GetDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
