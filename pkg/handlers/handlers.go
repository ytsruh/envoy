package handlers

import (
	"net/http"
)

type HealthHandler interface {
	Health(w http.ResponseWriter, r *http.Request) error
}

type GreetingHandler interface {
	Hello(w http.ResponseWriter, r *http.Request) error
	Goodbye(w http.ResponseWriter, r *http.Request) error
}
