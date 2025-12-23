package server

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	addr := ":8080"
	dbService := &mockDBService{}

	s := New(addr, dbService, "test-secret")

	if s == nil {
		t.Fatal("New returned nil")
	}

	if s.addr != addr {
		t.Errorf("Expected server address to be %s, got %s", addr, s.addr)
	}

	if s.router == nil {
		t.Fatal("Expected router instance to be initialized")
	}
}
