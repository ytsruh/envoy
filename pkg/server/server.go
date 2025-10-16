package server

import (
	"context"
	"database/sql"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	dbpkg "ytsruh.com/envoy/pkg/database"
	database "ytsruh.com/envoy/pkg/database/generated"
)

type DBService interface {
	GetDB() *sql.DB
	GetQueries() database.Querier
	Health() *dbpkg.HealthStatus
	Close() error
}

type Server struct {
	echo      *echo.Echo
	dbService DBService
	addr      string
}

func New(addr string, dbService DBService) *Server {
	e := echo.New()
	s := &Server{
		echo:      e,
		dbService: dbService,
		addr:      addr,
	}
	RegisterMiddleware(e, DefaultSecurityHeadersConfig(), 30*time.Second)
	s.RegisterRoutes()

	return s
}

func gracefulShutdown(e *echo.Echo, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")
	done <- true
}

func (s *Server) Start() {
	done := make(chan bool, 1)

	go gracefulShutdown(s.echo, done)

	log.Printf("Server starting on %s", s.addr)
	err := s.echo.Start(s.addr)
	if err != nil && err.Error() != "http: Server closed" {
		panic(err)
	}

	<-done
	log.Println("Graceful shutdown complete.")
}
