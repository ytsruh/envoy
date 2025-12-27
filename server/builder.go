package server

import (
	"time"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/server/middleware"
	"ytsruh.com/envoy/server/utils"
)

type ServerBuilder struct {
	router          *echo.Echo
	addr            string
	dbService       DBService
	middlewares     []echo.MiddlewareFunc
	timeoutDuration time.Duration
	jwtSecret       string
}

func NewBuilder(addr string, dbService DBService, jwtSecret string) *ServerBuilder {
	return &ServerBuilder{
		router:          echo.New(),
		addr:            addr,
		dbService:       dbService,
		timeoutDuration: 30 * time.Second,
		jwtSecret:       jwtSecret,
	}
}

func (b *ServerBuilder) WithMiddleware(mw echo.MiddlewareFunc) *ServerBuilder {
	b.middlewares = append(b.middlewares, mw)
	return b
}

func (b *ServerBuilder) WithTimeout(duration time.Duration) *ServerBuilder {
	b.timeoutDuration = duration
	return b
}

func (b *ServerBuilder) Build() (*Server, error) {
	accessControl := utils.NewAccessControlService(b.dbService.GetQueries())
	s := &Server{
		router:        b.router,
		dbService:     b.dbService,
		accessControl: accessControl,
		addr:          b.addr,
		jwtSecret:     b.jwtSecret,
	}

	middleware.RegisterMiddleware(b.router, b.timeoutDuration)

	for _, mw := range b.middlewares {
		b.router.Use(mw)
	}

	s.RegisterRoutes()

	return s, nil
}
