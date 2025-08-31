package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"ytsruh.com/envoy/pkg/database"
)

// Server holds the dependencies for a HTTP server.
type Server struct {
	srv *http.Server
	db  database.Service // TODO: Need to replace with interface to avoid the dependency on the database package
}

// New creates and configures a new Server instance.
func New(addr string, dbService database.Service) *Server {
	router := NewRouter()
	s := &Server{
		db: dbService,
	}
	// Register routes as method on Server so it can access dependencies directly
	s.RegisterRoutes(router)
	mux := registerMiddleware(router)
	return &Server{
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		db: dbService,
	}
}

func gracefulShutdown(server *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

// Start runs the server and waits for a graceful shutdown.
func (s *Server) Start() {
	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(s.srv, done)

	log.Printf("Server starting on %s", s.srv.Addr)
	err := s.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
