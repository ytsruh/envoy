package handlers

import (
	_ "embed"
	"fmt"
	"net/http"

	scalargo "github.com/bdpiprava/scalar-go"
)

//go:embed docs.json
var openapiSpec []byte

type DocsHandler interface {
	Docs(w http.ResponseWriter, r *http.Request) error
	OpenAPI(w http.ResponseWriter, r *http.Request) error
}

type DocsHandlerImpl struct{}

func NewDocsHandler() DocsHandler {
	return &DocsHandlerImpl{}
}

func (h *DocsHandlerImpl) Docs(w http.ResponseWriter, r *http.Request) error {
	// Generate beautiful docs with ScalarUI using embedded spec
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "Failed to generate documentation: %v"}`, err)
		return nil
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
	return nil
}

func (h *DocsHandlerImpl) OpenAPI(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	w.Write(openapiSpec)
	return nil
}
