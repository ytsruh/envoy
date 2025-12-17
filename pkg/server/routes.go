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

	s.RegisterFaviconHandler()
}

func (s *Server) RegisterHealthHandler(h handlers.HealthHandler) {
	s.router.GET("/health", func(c echo.Context) error {
		return h.Health(c.Response(), c.Request())
	})
}

func (s *Server) RegisterAuthHandlers(h handlers.AuthHandler) {
	s.router.POST("/auth/register", func(c echo.Context) error {
		return h.Register(c.Response(), c.Request())
	})
	s.router.POST("/auth/login", func(c echo.Context) error {
		return h.Login(c.Response(), c.Request())
	})
	authMiddleware := JWTAuthMiddleware(s.jwtSecret)
	s.router.GET("/auth/profile", authMiddleware(func(c echo.Context) error {
		return h.GetProfile(c.Response(), c.Request())
	}))
}

func (s *Server) RegisterFaviconHandler() {
	s.router.GET("/favicon.ico", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
}
