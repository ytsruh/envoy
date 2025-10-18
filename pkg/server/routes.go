package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/pkg/handlers"
)

func (s *Server) RegisterRoutes() {
	// Public routes
	healthHandler := handlers.NewHealthHandler(s.dbService.GetQueries())
	s.RegisterHealthHandler(healthHandler)

	authHandler := handlers.NewAuthHandler(s.dbService.GetQueries(), s.jwtSecret)
	s.RegisterAuthHandlers(authHandler)

	// Protected routes (require JWT authentication)
	greetingHandler := handlers.NewGreetingHandler(s.dbService.GetQueries())
	s.RegisterGreetingHandlers(greetingHandler)

	s.RegisterFaviconHandler()
}

func (s *Server) RegisterHealthHandler(h handlers.HealthHandler) {
	s.echo.GET("/health", h.Health)
}

func (s *Server) RegisterGreetingHandlers(h handlers.GreetingHandler) {
	// Create JWT middleware
	authMiddleware := JWTAuthMiddleware(s.jwtSecret)
	s.echo.Use(authMiddleware)

	// Protected greeting routes
	s.echo.GET("/hello", h.Hello)
	s.echo.POST("/goodbye", h.Goodbye)
}

func (s *Server) RegisterAuthHandlers(h handlers.AuthHandler) {
	s.echo.POST("/auth/register", h.Register)
	s.echo.POST("/auth/login", h.Login)
	authMiddleware := JWTAuthMiddleware(s.jwtSecret)
	s.echo.GET("/auth/profile", h.GetProfile, authMiddleware)
}

func (s *Server) RegisterFaviconHandler() {
	s.echo.GET("/favicon.ico", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
}
