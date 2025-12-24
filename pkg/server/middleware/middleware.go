package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RegisterMiddleware(e *echo.Echo, timeout time.Duration) {
	e.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently, // 301 - remove trailing slashes
	}))
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${remote_addr}] ${method} ${uri} ${status} ${latency_human}\n",
	}))
	e.Use(middleware.Secure())
	e.Use(middleware.Gzip())
	e.Use(middleware.BodyLimit("2M"))
	e.Use(middleware.Decompress())
	e.Use(TimeoutMiddleware(timeout))
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(10), // 10 requests per second
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))
}

func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
		Timeout: timeout,
	})
}
