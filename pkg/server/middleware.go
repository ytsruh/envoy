package server

import (
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter = rate.NewLimiter(rate.Limit(10), 10)
	rl.limiters[ip] = limiter

	return limiter
}

func RegisterMiddleware(e *echo.Echo, timeout time.Duration) {
	e.Use(RecoveryMiddleware())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${remote_addr}] ${method} ${uri} ${status} ${latency_human}\n",
	}))
	e.Use(middleware.Secure())
	e.Use(middleware.Gzip())
	e.Use(middleware.BodyLimit("2M"))
	e.Use(middleware.Decompress())
	e.Use(TimeoutMiddleware(timeout))
	e.Use(RateLimiterMiddleware(NewRateLimiter()))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))
}

func RecoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("PANIC: %v\n%s", err, debug.Stack())
					c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
				}
			}()
			return next(c)
		}
	}
}

func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
		Timeout: timeout,
	})
}

func RateLimiterMiddleware(rl *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := rl.GetLimiter(ip)

			if !limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "Rate limit exceeded",
				})
			}

			return next(c)
		}
	}
}
