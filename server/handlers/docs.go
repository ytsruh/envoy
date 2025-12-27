package handlers

import (
	_ "embed"
	"fmt"
	"net/http"

	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/labstack/echo/v4"
)

//go:embed docs.json
var openapiSpec []byte

func Docs(c echo.Context, _ *HandlerContext) error {
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
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to generate documentation: %w", err))
	}
	return c.HTML(http.StatusOK, html)
}

func OpenAPI(c echo.Context, _ *HandlerContext) error {
	return c.Blob(http.StatusOK, "application/json", openapiSpec)
}
