package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/server/handlers"
	"ytsruh.com/envoy/server/middleware"
)

func (s *Server) RegisterRoutes() {
	s.RegisterHealthHandler()
	s.RegisterHomeHandler()
	s.RegisterVersionHandler()
	s.RegisterAuthHandlers()
	s.RegisterProjectHandlers()
	s.RegisterProjectSharingHandlers()
	s.RegisterEnvironmentHandlers()
	s.RegisterEnvironmentVariableHandlers()
	s.RegisterDocsHandlers()
	s.RegisterFaviconHandler()
}

func (s *Server) RegisterHealthHandler() {
	ctx := handlers.NewHandlerContext(s.dbService.GetQueries(), s.jwtSecret, s.accessControl)
	s.router.GET("/health", func(c echo.Context) error {
		return handlers.Health(c, ctx)
	})
}

func (s *Server) RegisterHomeHandler() {
	ctx := handlers.NewHandlerContext(s.dbService.GetQueries(), s.jwtSecret, s.accessControl)
	s.router.GET("/", func(c echo.Context) error {
		return handlers.Home(c, ctx)
	})
}

func (s *Server) RegisterDocsHandlers() {
	ctx := handlers.NewHandlerContext(s.dbService.GetQueries(), s.jwtSecret, s.accessControl)
	s.router.GET("/openapi.json", func(c echo.Context) error {
		return handlers.OpenAPI(c, ctx)
	})
	s.router.GET("/docs", func(c echo.Context) error {
		return handlers.Docs(c, ctx)
	})
}

func (s *Server) RegisterFaviconHandler() {
	s.router.GET("/favicon.ico", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
}

func (s *Server) RegisterAuthHandlers() {
	auth := middleware.JWTAuthMiddleware(s.jwtSecret)
	ctx := handlers.NewHandlerContext(s.dbService.GetQueries(), s.jwtSecret, s.accessControl)
	s.router.POST("/auth/register", func(c echo.Context) error {
		return handlers.Register(c, ctx)
	})
	s.router.POST("/auth/login", func(c echo.Context) error {
		return handlers.Login(c, ctx)
	})
	s.router.GET("/auth/profile", auth(func(c echo.Context) error {
		return handlers.GetProfile(c, ctx)
	}))
}

func (s *Server) RegisterProjectHandlers() {
	auth := middleware.JWTAuthMiddleware(s.jwtSecret)
	ctx := handlers.NewHandlerContext(s.dbService.GetQueries(), s.jwtSecret, s.accessControl)
	s.router.POST("/projects", auth(func(c echo.Context) error {
		return handlers.CreateProject(c, ctx)
	}))
	s.router.GET("/projects/:id", auth(func(c echo.Context) error {
		return handlers.GetProject(c, ctx)
	}))
	s.router.GET("/projects", auth(func(c echo.Context) error {
		return handlers.ListProjects(c, ctx)
	}))
	s.router.PUT("/projects/:id", auth(func(c echo.Context) error {
		return handlers.UpdateProject(c, ctx)
	}))
	s.router.DELETE("/projects/:id", auth(func(c echo.Context) error {
		return handlers.DeleteProject(c, ctx)
	}))
}

func (s *Server) RegisterProjectSharingHandlers() {
	auth := middleware.JWTAuthMiddleware(s.jwtSecret)
	ctx := handlers.NewHandlerContext(s.dbService.GetQueries(), s.jwtSecret, s.accessControl)
	s.router.POST("/projects/:id/members", auth(func(c echo.Context) error {
		return handlers.AddUserToProject(c, ctx)
	}))
	s.router.DELETE("/projects/:id/members/:user_id", auth(func(c echo.Context) error {
		return handlers.RemoveUserFromProject(c, ctx)
	}))
	s.router.PUT("/projects/:id/members/:user_id", auth(func(c echo.Context) error {
		return handlers.UpdateUserRole(c, ctx)
	}))
	s.router.GET("/projects/:id/members", auth(func(c echo.Context) error {
		return handlers.GetProjectUsers(c, ctx)
	}))
	s.router.GET("/user/projects", auth(func(c echo.Context) error {
		return handlers.ListUserProjects(c, ctx)
	}))
}

func (s *Server) RegisterEnvironmentHandlers() {
	auth := middleware.JWTAuthMiddleware(s.jwtSecret)
	ctx := handlers.NewHandlerContext(s.dbService.GetQueries(), s.jwtSecret, s.accessControl)
	s.router.POST("/projects/:project_id/environments", auth(func(c echo.Context) error {
		return handlers.CreateEnvironment(c, ctx)
	}))
	s.router.GET("/projects/:project_id/environments/:id", auth(func(c echo.Context) error {
		return handlers.GetEnvironment(c, ctx)
	}))
	s.router.GET("/projects/:project_id/environments", auth(func(c echo.Context) error {
		return handlers.ListEnvironments(c, ctx)
	}))
	s.router.PUT("/projects/:project_id/environments/:id", auth(func(c echo.Context) error {
		return handlers.UpdateEnvironment(c, ctx)
	}))
	s.router.DELETE("/projects/:project_id/environments/:id", auth(func(c echo.Context) error {
		return handlers.DeleteEnvironment(c, ctx)
	}))
}

func (s *Server) RegisterEnvironmentVariableHandlers() {
	auth := middleware.JWTAuthMiddleware(s.jwtSecret)
	ctx := handlers.NewHandlerContext(s.dbService.GetQueries(), s.jwtSecret, s.accessControl)
	s.router.POST("/projects/:project_id/environments/:environment_id/variables", auth(func(c echo.Context) error {
		return handlers.CreateEnvironmentVariable(c, ctx)
	}))
	s.router.GET("/projects/:project_id/environments/:environment_id/variables/:id", auth(func(c echo.Context) error {
		return handlers.GetEnvironmentVariable(c, ctx)
	}))
	s.router.GET("/projects/:project_id/environments/:environment_id/variables", auth(func(c echo.Context) error {
		return handlers.ListEnvironmentVariables(c, ctx)
	}))
	s.router.PUT("/projects/:project_id/environments/:environment_id/variables/:id", auth(func(c echo.Context) error {
		return handlers.UpdateEnvironmentVariable(c, ctx)
	}))
	s.router.DELETE("/projects/:project_id/environments/:environment_id/variables/:id", auth(func(c echo.Context) error {
		return handlers.DeleteEnvironmentVariable(c, ctx)
	}))
}

func (s *Server) RegisterVersionHandler() {
	s.router.GET("/version", func(c echo.Context) error {
		return handlers.Version(c, nil)
	})
}
