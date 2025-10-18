package server

import (
	"time"

	"github.com/labstack/echo/v4"
)

type ServerBuilder struct {
	echo            *echo.Echo
	addr            string
	dbService       DBService
	middlewares     []echo.MiddlewareFunc
	timeoutDuration time.Duration
	jwtSecret       string
}

func NewBuilder(addr string, dbService DBService, jwtSecret string) *ServerBuilder {
	return &ServerBuilder{
		echo:            echo.New(),
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
	s := &Server{
		echo:      b.echo,
		dbService: b.dbService,
		addr:      b.addr,
		jwtSecret: b.jwtSecret,
	}

	RegisterMiddleware(b.echo, b.timeoutDuration)

	for _, mw := range b.middlewares {
		b.echo.Use(mw)
	}

	s.RegisterRoutes()

	return s, nil
}
