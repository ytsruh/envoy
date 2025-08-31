package server

import (
	"database/sql"
	gen "ytsruh.com/envoy/pkg/database/generated"
)

// shared mock for server package tests
type mockDBService struct{}

func (m *mockDBService) Health() map[string]string {
	return map[string]string{"status": "ok", "message": "mock health check"}
}
func (m *mockDBService) GetDB() *sql.DB           { return nil }
func (m *mockDBService) GetQueries() *gen.Queries { return nil }
func (m *mockDBService) Close() error             { return nil }
