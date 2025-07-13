package server

import (
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	addr := ":8080"
	handler := http.NewServeMux() // Use a simple http.Handler for testing

	s := New(addr, handler)

	if s == nil {
		t.Fatal("New returned nil")
	}

	if s.srv.Addr != addr {
		t.Errorf("Expected server address to be %s, got %s", addr, s.srv.Addr)
	}
}
