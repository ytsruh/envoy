package cron

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	db := &sql.DB{}

	s := New(db, logger)

	if s == nil {
		t.Error("expected service, got nil")
	}

	svc, ok := s.(*service)
	if !ok {
		t.Error("expected *service type")
	}

	if svc.db != db {
		t.Error("expected db to be set")
	}

	if svc.logger != logger {
		t.Error("expected logger to be set")
	}

	if svc.scheduler == nil {
		t.Error("expected scheduler to be initialized")
	}
}

func TestAddJob(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	db := &sql.DB{}
	svc := New(db, logger).(*service)

	testJob := func() error {
		return nil
	}

	err := svc.AddJob("*/1 * * * * *", testJob)
	if err != nil {
		t.Fatalf("AddJob failed: %v", err)
	}

	entries := svc.scheduler.Entries()
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestAddJobInvalidSchedule(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	db := &sql.DB{}
	svc := New(db, logger).(*service)

	testJob := func() error {
		return nil
	}

	err := svc.AddJob("invalid schedule", testJob)
	if err == nil {
		t.Error("expected error for invalid schedule, got nil")
	}
}

func TestStartStop(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	db := &sql.DB{}
	svc := New(db, logger).(*service)

	svc.AddJob("*/1 * * * * *", func() error { return nil })
	svc.Start()
	time.Sleep(100 * time.Millisecond)

	entries := svc.scheduler.Entries()
	if len(entries) == 0 {
		t.Error("scheduler should have entries after Start()")
	}

	svc.Stop()
	t.Log("scheduler stopped successfully")
}

func TestJobExecution(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	db := &sql.DB{}
	svc := New(db, logger).(*service)

	called := make(chan bool, 1)
	testJob := func() error {
		called <- true
		return nil
	}

	svc.AddJob("*/1 * * * * *", testJob)
	svc.Start()
	defer svc.Stop()

	select {
	case <-called:
		t.Log("job executed successfully")
	case <-time.After(2 * time.Second):
		t.Error("job did not execute within timeout")
	}
}

func TestJobExecutionWithError(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	db := &sql.DB{}
	svc := New(db, logger).(*service)

	called := make(chan bool, 1)
	testJob := func() error {
		called <- true
		return errors.New("job error")
	}

	svc.AddJob("*/1 * * * * *", testJob)
	svc.Start()
	defer svc.Stop()

	select {
	case <-called:
		t.Log("job executed and returned error (as expected)")
	case <-time.After(2 * time.Second):
		t.Error("job did not execute within timeout")
	}
}

func TestInspect(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	db := &sql.DB{}
	svc := New(db, logger).(*service)

	svc.AddJob("*/1 * * * * *", func() error { return nil })
	svc.AddJob("0 0 * * *", func() error { return nil })

	result := svc.Inspect()

	if result == "" {
		t.Error("Inspect should return non-empty string")
	}

	if len(svc.scheduler.Entries()) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(svc.scheduler.Entries()))
	}
}

func TestInspectEmpty(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	db := &sql.DB{}
	svc := New(db, logger).(*service)

	result := svc.Inspect()

	if result != "" {
		t.Errorf("expected empty string for no jobs, got %q", result)
	}
}

func TestDatabaseHealthCheck(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)

	t.Run("returns job function", func(t *testing.T) {
		db := &sql.DB{}
		job := DatabaseHealthCheck(db, logger)

		if job == nil {
			t.Error("expected job function, got nil")
		}
	})
}

func TestMultipleJobs(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	db := &sql.DB{}
	svc := New(db, logger).(*service)

	job1Called := make(chan bool, 1)
	job2Called := make(chan bool, 1)

	svc.AddJob("*/1 * * * * *", func() error {
		job1Called <- true
		return nil
	})

	svc.AddJob("*/1 * * * * *", func() error {
		job2Called <- true
		return nil
	})

	svc.Start()
	defer svc.Stop()

	timeout := time.After(2 * time.Second)

	job1Executed := false
	job2Executed := false

	for {
		select {
		case <-job1Called:
			job1Executed = true
		case <-job2Called:
			job2Executed = true
		case <-timeout:
			if !job1Executed || !job2Executed {
				t.Error("not all jobs executed within timeout")
			}
			return
		}
	}
}
