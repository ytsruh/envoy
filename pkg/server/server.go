package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	database "ytsruh.com/envoy/pkg/database/generated"
	"ytsruh.com/envoy/pkg/handlers"
)

// Represents a database service
type DBService interface {
	// Return the database connection
	GetDB() *sql.DB
	// GetQueries returns the database queries object.
	GetQueries() database.Querier
	// Health returns a map of health status information. The keys and values in the map are service-specific.
	Health() map[string]string
	// Close terminates the database connection.It returns an error if the connection cannot be closed.
	Close() error
}

// Server holds the dependencies for a HTTP server.
type Server struct {
	server    *http.Server
	dbService DBService
	router    *Router
}

// New creates and configures a new Server instance.
func New(addr string, dbService DBService) *Server {
	router := NewRouter()
	s := &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: router.mux,
		},
		dbService: dbService,
		router:    &router,
	}
	// Register routes & middleware
	// Create handlers with injected queries
	h := handlers.New(dbService.GetQueries())
	s.RegisterRoutes(h)
	s.RegisterMiddleware()

	return s
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
	go gracefulShutdown(s.server, done)

	log.Printf("Server starting on %s", s.server.Addr)
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
