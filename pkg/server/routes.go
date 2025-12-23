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

	homeHandler := handlers.NewHomeHandler()
	s.RegisterHomeHandler(homeHandler)

	authHandler := handlers.NewAuthHandler(s.dbService.GetQueries(), s.jwtSecret)
	s.RegisterAuthHandlers(authHandler)

	// Protected project routes
	projectHandler := handlers.NewProjectHandler(s.dbService.GetQueries())
	s.RegisterProjectHandlers(projectHandler, s.jwtSecret)

	// Protected project sharing routes
	projectSharingHandler := handlers.NewProjectSharingHandler(s.dbService.GetQueries())
	s.RegisterProjectSharingHandlers(projectSharingHandler, s.jwtSecret)

	// Documentation routes
	docsHandler := handlers.NewDocsHandler()
	s.RegisterDocsHandlers(docsHandler)

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

func (s *Server) RegisterProjectHandlers(h handlers.ProjectHandler, jwtSecret string) {
	authMiddleware := JWTAuthMiddleware(jwtSecret)

	// Create project
	s.router.POST("/projects", authMiddleware(func(c echo.Context) error {
		return h.CreateProject(c.Response(), c.Request())
	}))

	// Get single project
	s.router.GET("/projects/:id", authMiddleware(func(c echo.Context) error {
		return h.GetProject(c.Response(), c.Request())
	}))

	// List projects
	s.router.GET("/projects", authMiddleware(func(c echo.Context) error {
		return h.ListProjects(c.Response(), c.Request())
	}))

	// Update project
	s.router.PUT("/projects/:id", authMiddleware(func(c echo.Context) error {
		return h.UpdateProject(c.Response(), c.Request())
	}))

	// Delete project
	s.router.DELETE("/projects/:id", authMiddleware(func(c echo.Context) error {
		return h.DeleteProject(c.Response(), c.Request())
	}))
}

func (s *Server) RegisterProjectSharingHandlers(h handlers.ProjectSharingHandler, jwtSecret string) {
	authMiddleware := JWTAuthMiddleware(jwtSecret)

	// Add user to project
	s.router.POST("/projects/:id/members", authMiddleware(func(c echo.Context) error {
		return h.AddUserToProject(c.Response(), c.Request())
	}))

	// Remove user from project
	s.router.DELETE("/projects/:id/members/:user_id", authMiddleware(func(c echo.Context) error {
		return h.RemoveUserFromProject(c.Response(), c.Request())
	}))

	// Update user role in project
	s.router.PUT("/projects/:id/members/:user_id", authMiddleware(func(c echo.Context) error {
		return h.UpdateUserRole(c.Response(), c.Request())
	}))

	// Get project members
	s.router.GET("/projects/:id/members", authMiddleware(func(c echo.Context) error {
		return h.GetProjectUsers(c.Response(), c.Request())
	}))

	// List all projects accessible to user (owned + shared)
	s.router.GET("/user/projects", authMiddleware(func(c echo.Context) error {
		return h.ListUserProjects(c.Response(), c.Request())
	}))
}

func (s *Server) RegisterHomeHandler(h *handlers.HomeHandler) {
	s.router.GET("/", func(c echo.Context) error {
		return h.Home(c.Response(), c.Request())
	})
}

func (s *Server) RegisterDocsHandlers(h handlers.DocsHandler) {
	// OpenAPI specification JSON
	s.router.GET("/openapi.json", func(c echo.Context) error {
		return h.OpenAPI(c.Response(), c.Request())
	})
	// API documentation interface
	s.router.GET("/docs", func(c echo.Context) error {
		return h.Docs(c.Response(), c.Request())
	})
}

func (s *Server) RegisterFaviconHandler() {
	s.router.GET("/favicon.ico", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
}
