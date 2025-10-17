package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "modernc.org/sqlite"
	database "ytsruh.com/envoy/pkg/database/generated"
)

type HealthStatus struct {
	Status            string
	Message           string
	OpenConnections   int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxLifetimeClosed int64
}

type Service struct {
	db      *sql.DB
	queries *database.Queries
}

func NewService(dbPath string) (*Service, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error creating connector: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	queries := database.New(db)

	return &Service{
		db:      db,
		queries: queries,
	}, nil
}

func (s *Service) GetDB() *sql.DB {
	return s.db
}

func (s *Service) GetQueries() database.Querier {
	return s.queries
}

func (s *Service) Health() (*HealthStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	status := &HealthStatus{}

	err := s.db.PingContext(ctx)
	if err != nil {
		status.Status = "down"
		status.Message = fmt.Sprintf("db down: %v", err)
		return status, fmt.Errorf("database health check failed: %w", err)
	}

	status.Status = "up"
	status.Message = "It's healthy"

	dbStats := s.db.Stats()
	status.OpenConnections = dbStats.OpenConnections
	status.InUse = dbStats.InUse
	status.Idle = dbStats.Idle
	status.WaitCount = dbStats.WaitCount
	status.WaitDuration = dbStats.WaitDuration
	status.MaxIdleClosed = dbStats.MaxIdleClosed
	status.MaxLifetimeClosed = dbStats.MaxLifetimeClosed

	if dbStats.OpenConnections > 40 {
		status.Message = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		status.Message = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		status.Message = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		status.Message = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return status, nil
}

// Close closes the database connection. It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil. If an error occurs while closing the connection, it returns the error.
func (s *Service) Close() error {
	log.Println("Disconnected from database")
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}
