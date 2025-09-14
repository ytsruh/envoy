package server

import (
	"testing"
)

func TestNew(t *testing.T) {
	addr := ":8080"
	dbService := &mockDBService{}

	s := New(addr, dbService)

	if s == nil {
		t.Fatal("New returned nil")
	}

	if s.server.Addr != addr {
		t.Errorf("Expected server address to be %s, got %s", addr, s.server.Addr)
	}
}
