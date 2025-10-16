package cron

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/robfig/cron"
)

type Job func() error

type Service interface {
	Start()
	Stop()
	Inspect() string
	AddJob(schedule string, job Job) error
}

type service struct {
	scheduler *cron.Cron
	db        *sql.DB
	logger    *log.Logger
}

func New(db *sql.DB, logger *log.Logger) Service {
	return &service{
		scheduler: cron.New(),
		db:        db,
		logger:    logger,
	}
}

func (s *service) AddJob(schedule string, job Job) error {
	return s.scheduler.AddFunc(schedule, func() {
		if err := job(); err != nil {
			s.logger.Printf("cron job failed: %v", err)
		}
	})
}

func (s *service) Start() {
	s.scheduler.Start()
}

func (s *service) Stop() {
	s.scheduler.Stop()
}

func (s *service) Inspect() string {
	entries := s.scheduler.Entries()
	var result string
	for i, entry := range entries {
		result += fmt.Sprintf("Cron Job: %d, Schedule: %T, Next: %s\n", i, entry.Schedule, entry.Next)
	}
	return result
}

func DatabaseHealthCheck(db *sql.DB, logger *log.Logger) Job {
	return func() error {
		if err := db.Ping(); err != nil {
			return fmt.Errorf("database ping: %w", err)
		}
		logger.Println("database health check passed")
		return nil
	}
}
