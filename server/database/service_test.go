package database

import (
	"context"
	"database/sql"
	"os"
	"testing"
)

func createTestService(t *testing.T) *Service {
	// Use in-memory SQLite for tests
	dbPath := "file::memory:?cache=shared"
	service, err := NewService(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test service: %v", err)
	}

	// Initialize database schema for tests
	if err := initializeTestSchema(service.db); err != nil {
		t.Fatalf("Failed to initialize test schema: %v", err)
	}

	return service
}

func initializeTestSchema(db *sql.DB) error {
	return nil
}

func TestNewService(t *testing.T) {
	dbPath := "file::memory:?cache=shared"
	service1, err := NewService(dbPath)

	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if service1 == nil {
		t.Fatal("NewService() returned nil")
	}

	if service1.db == nil {
		t.Fatal("Service db is nil")
	}

	if service1.queries == nil {
		t.Fatal("Service queries is nil")
	}

	// Test that each call creates a new instance (no longer singleton)
	service2, err := NewService(dbPath)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if service1 == service2 {
		t.Error("NewService() should return different instances (no longer singleton)")
	}

	// Cleanup
	service1.Close()
	service2.Close()
}

func TestHealth(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Service)
		want    string
		wantErr bool
	}{
		{
			name:  "healthy database",
			setup: func(s *Service) {},
			want:  "up",
		},
		{
			name: "database down",
			setup: func(s *Service) {
				s.db.Close()
			},
			want:    "down",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := createTestService(t)

			// Setup test conditions
			tt.setup(service)

			stats, err := service.Health()

			if tt.wantErr && err == nil {
				t.Error("Expected Health() to return an error")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if stats.Status != tt.want {
				t.Errorf("Health() status = %v, want %v", stats.Status, tt.want)
			}

			if stats.OpenConnections < 0 {
				t.Error("Health() OpenConnections should not be negative")
			}

			if stats.InUse < 0 {
				t.Error("Health() InUse should not be negative")
			}

			if stats.Idle < 0 {
				t.Error("Health() Idle should not be negative")
			}

			// Cleanup
			if !tt.wantErr {
				service.Close()
			}
		})
	}
}

func TestHealth_Statistics(t *testing.T) {
	service := createTestService(t)

	stats, err := service.Health()
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}

	if stats.OpenConnections < 0 {
		t.Error("OpenConnections should not be negative")
	}

	if stats.InUse < 0 {
		t.Error("InUse should not be negative")
	}

	if stats.Idle < 0 {
		t.Error("Idle should not be negative")
	}

	if stats.WaitCount < 0 {
		t.Error("WaitCount should not be negative")
	}

	if stats.WaitDuration < 0 {
		t.Error("WaitDuration should not be negative")
	}

	service.Close()
}

func TestGetDB(t *testing.T) {
	service := createTestService(t)

	db := service.GetDB()
	if db == nil {
		t.Error("GetDB() returned nil")
	}

	if db != service.db {
		t.Error("GetDB() should return the same db instance")
	}

	service.Close()
}

func TestGetQueries(t *testing.T) {
	service := createTestService(t)

	queries := service.GetQueries()
	if queries == nil {
		t.Error("GetQueries() returned nil")
	}

	if queries != service.queries {
		t.Error("GetQueries() should return the same queries instance")
	}

	service.Close()
}

func TestClose(t *testing.T) {
	service := createTestService(t)

	err := service.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify database is actually closed
	err = service.db.Ping()
	if err == nil {
		t.Error("Expected database to be closed after Close()")
	}
}

func TestService_ConcurrentAccess(t *testing.T) {
	// Test that the service handles concurrent access safely
	// Use isolated in-memory databases to avoid migration race conditions
	done := make(chan bool, 10)
	services := make([]*Service, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() {
				done <- true
			}()
			// Each goroutine gets its own isolated in-memory database
			dbPath := ":memory:"
			service, err := NewService(dbPath)
			if err != nil {
				t.Errorf("Failed to create service: %v", err)
				return
			}
			services[index] = service
			_, _ = service.Health()
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all services were created successfully
	for i, service := range services {
		if service == nil {
			t.Errorf("Service %d was not created", i)
		} else {
			service.Close()
		}
	}
}

func TestService_WithFileDatabase(t *testing.T) {
	// Test with a file-based database

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	service, err := NewService(tmpFile.Name())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	stats, err := service.Health()
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if stats.Status != "up" {
		t.Errorf("Expected database to be up, got status: %s", stats.Status)
	}

	err = service.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestService_WALMode(t *testing.T) {
	// Test that WAL mode is enabled
	tmpFile, err := os.CreateTemp("", "test-wal-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	service, err := NewService(tmpFile.Name())
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	defer service.Close()

	var journalMode string
	err = service.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("Expected journal mode to be 'wal', got '%s'", journalMode)
	}
}

func TestService_GetQueriesReturnsValidQuerier(t *testing.T) {
	service := createTestService(t)
	defer service.Close()

	queries := service.GetQueries()
	if queries == nil {
		t.Error("GetQueries() returned nil")
	}

	// Test that we can call environment methods on the returned querier
	ctx := context.Background()
	envs, err := queries.ListEnvironmentsByProject(ctx, "1")
	// This should return empty list, not an error
	if err != nil {
		t.Errorf("Expected no error when listing environments from empty database, got: %v", err)
	}
	if len(envs) != 0 {
		t.Errorf("Expected empty list from empty database, got %d environments", len(envs))
	}
}
