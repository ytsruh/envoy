package database

import (
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
	// Create basic schema for testing
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
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

			// Capture log output to prevent os.Exit from terminating tests
			// Note: In a real scenario, you might want to mock log.Fatal
			if tt.wantErr {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected Health() to panic on database down")
					}
				}()
			}

			stats := service.Health()

			if stats["status"] != tt.want {
				t.Errorf("Health() status = %v, want %v", stats["status"], tt.want)
			}

			// Verify required fields are present
			requiredFields := []string{"status", "open_connections", "in_use", "idle"}
			for _, field := range requiredFields {
				if _, exists := stats[field]; !exists {
					t.Errorf("Health() missing required field: %s", field)
				}
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

	stats := service.Health()

	// Verify statistics are properly formatted
	if stats["open_connections"] == "" {
		t.Error("open_connections should not be empty")
	}

	if stats["in_use"] == "" {
		t.Error("in_use should not be empty")
	}

	if stats["idle"] == "" {
		t.Error("idle should not be empty")
	}

	if stats["wait_count"] == "" {
		t.Error("wait_count should not be empty")
	}

	if stats["wait_duration"] == "" {
		t.Error("wait_duration should not be empty")
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
	dbPath := "file::memory:?cache=shared"

	// Create multiple services concurrently
	done := make(chan bool, 10)
	services := make([]*Service, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			service, err := NewService(dbPath)
			if err != nil {
				t.Errorf("Failed to create service: %v", err)
			}
			services[index] = service
			_ = service.Health()
			done <- true
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

	// Verify the service works
	stats := service.Health()
	if stats["status"] != "up" {
		t.Errorf("Expected database to be up, got status: %s", stats["status"])
	}

	err = service.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
