package server

import (
	_ "embed"
	"fmt"
	"net/http"

	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/labstack/echo/v4"
)

//go:embed docs.json
var openapiSpec []byte

type ErrorResponse struct {
	Error string `json:"error"`
}

func sendErrorResponse(c echo.Context, code int, err error) error {
	return c.JSON(code, ErrorResponse{Error: err.Error()})
}

func nullStringToStringPtr(ns any) *string {
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

func (s *Server) Docs(c echo.Context) error {
	html, err := scalargo.NewV2(
		scalargo.WithSpecBytes(openapiSpec),
		scalargo.WithTheme(scalargo.ThemeMoon),
		scalargo.WithLayout(scalargo.LayoutModern),
		scalargo.WithMetaDataOpts(
			scalargo.WithTitle("Envoy API Documentation"),
			scalargo.WithKeyValue("description", "Interactive API documentation for Envoy platform"),
		),
	)

	if err != nil {
		return sendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to generate documentation: %w", err))
	}
	return c.HTML(http.StatusOK, html)
}

func (s *Server) OpenAPI(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/json", openapiSpec)
}
