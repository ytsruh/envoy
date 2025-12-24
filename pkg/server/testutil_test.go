package server

import (
	"database/sql"

	dbpkg "ytsruh.com/envoy/pkg/database"
	gen "ytsruh.com/envoy/pkg/database/generated"
)

type mockDBService struct {
	queries gen.Querier
}

func (m *mockDBService) Health() (*dbpkg.HealthStatus, error) {
	return &dbpkg.HealthStatus{Status: "ok", Message: "mock health check"}, nil
}
func (m *mockDBService) GetDB() *sql.DB          { return nil }
func (m *mockDBService) GetQueries() gen.Querier { return m.queries }
func (m *mockDBService) Close() error            { return nil }
