package routes

import (
	"fmt"
	"net/http"

	"ytsruh.com/envman/pkg/server"
)

func RegisterRoutes() http.Handler {
	r := server.NewRouter()

	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	r.Post("/goodbye", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Goodbye, World!")
	})

	return r
}
