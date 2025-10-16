package server

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	addr := ":8080"
	dbService := &mockDBService{}

	s := New(addr, dbService)

	if s == nil {
		t.Fatal("New returned nil")
	}

	if s.addr != addr {
		t.Errorf("Expected server address to be %s, got %s", addr, s.addr)
	}

	if s.echo == nil {
		t.Fatal("Expected echo instance to be initialized")
	}
}
