package server

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	dbpkg "ytsruh.com/envoy/pkg/database"
	gen "ytsruh.com/envoy/pkg/database/generated"
)

type mockDBService struct{}

func (m *mockDBService) Health() *dbpkg.HealthStatus {
	return &dbpkg.HealthStatus{Status: "ok", Message: "mock health check"}
}
func (m *mockDBService) GetDB() *sql.DB          { return nil }
func (m *mockDBService) GetQueries() gen.Querier { return nil }
func (m *mockDBService) Close() error            { return nil }

type mockHandler struct{}

func (h mockHandler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func (h mockHandler) Hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func (h mockHandler) Goodbye(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
