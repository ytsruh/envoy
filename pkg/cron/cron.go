package cron

import (
	"database/sql"
	"fmt"

	"github.com/robfig/cron"
)

// Represents a cron service
type Service interface {
	// Starts the cron job
	Start()
	// Terminates the cron jobs
	Stop()
	// Inspect returns a string representation of the service.
	Inspect() string
}

type service struct {
	scheduler *cron.Cron
	db        *sql.DB
}

var cronService *service

// New creates a new cron service with jobs
func New(db *sql.DB) Service {
	// Reuse existing cron scheduler
	if cronService != nil {
		return cronService
	}

	c := cron.New()

	cronService = &service{
		scheduler: c,
		db:        db,
	}

	// Add jobs to cron service
	c.AddFunc("*/30 * * * * *", func() {
		err := cronService.db.Ping()
		if err != nil {
			fmt.Println("Error pinging database:", err)
		}
		fmt.Println("Example Cron Job with access to DB")
	})

	return cronService
}

func (s *service) Start() {
	s.scheduler.Start()
}

func (s *service) Stop() {
	s.scheduler.Stop()
}

func (s *service) Inspect() string {
	entries := s.scheduler.Entries()
	for i, entry := range entries {
		fmt.Printf("Cron Job: %d, Schedule: %T, Next: %s\n", i, entry.Schedule, entry.Next)
	}
	return fmt.Sprintf("cron scheduler: %+v", s.scheduler)
}
