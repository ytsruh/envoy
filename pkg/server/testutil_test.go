package server

import (
	"database/sql"
	"fmt"
	"net/http"

	gen "ytsruh.com/envoy/pkg/database/generated"
)

// shared mock for server package tests
type mockDBService struct{}

func (m *mockDBService) Health() map[string]string {
	return map[string]string{"status": "ok", "message": "mock health check"}
}
func (m *mockDBService) GetDB() *sql.DB          { return nil }
func (m *mockDBService) GetQueries() gen.Querier { return nil }
func (m *mockDBService) Close() error            { return nil }

type mockHandler struct{}

func (h mockHandler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	}
}

func (h mockHandler) Hello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	}
}

func (h mockHandler) Goodbye() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	}
}
