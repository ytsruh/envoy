package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/server/database"
	queries "ytsruh.com/envoy/server/database/generated"
	"ytsruh.com/envoy/server/middleware"
	"ytsruh.com/envoy/server/utils"
)

type DBService interface {
	GetDB() *sql.DB
	GetQueries() queries.Querier
	Health() (*database.HealthStatus, error)
	Close() error
}

type Server struct {
	router        *echo.Echo
	dbService     DBService
	accessControl utils.AccessControlService
	addr          string
	jwtSecret     string
}

func New(addr string, env *utils.EnvVar) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Create a DBService instance
	dbService, err := database.NewService(env.DB_URL, env.DB_TOKEN)
	if err != nil {
		panic(err)
	}
	// Create an AccessControlService instance
	accessControl := utils.NewAccessControlService(dbService.GetQueries())

	// Create a Server instance
	server := &Server{
		router:        e,
		dbService:     dbService,
		accessControl: accessControl,
		addr:          addr,
		jwtSecret:     env.JWT_SECRET,
	}

	// Register middleware & routers
	middleware.RegisterMiddleware(e, 30*time.Second)
	server.RegisterRoutes()

	return server
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

	go gracefulShutdown(s.router, done)

	// Clickable link to open the server URL
	url := fmt.Sprintf("http://localhost%s", s.addr)
	fmt.Println("----------------------------------")
	fmt.Println("Server started")
	fmt.Printf("\033]8;;%s\033\\Click to open: %s\033]8;;\033\\\n", url, url)
	fmt.Println("----------------------------------")

	err := s.router.Start(s.addr)
	if err != nil && err.Error() != "http: Server closed" {
		panic(err)
	}

	<-done
	log.Println("Graceful shutdown complete.")
}
