package server

import (
	"testing"

	"ytsruh.com/envoy/pkg/database"
)

func TestNew(t *testing.T) {
	addr := ":8080"
	dbService := database.NewService("test.db")

	s := New(addr, dbService)

	if s == nil {
		t.Fatal("New returned nil")
	}

	if s.srv.Addr != addr {
		t.Errorf("Expected server address to be %s, got %s", addr, s.srv.Addr)
	}
}
