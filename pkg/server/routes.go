package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterRoutes() {
	s.RegisterHealthHandler()
	s.RegisterHomeHandler()
	s.RegisterAuthHandlers()
	s.RegisterProjectHandlers()
	s.RegisterProjectSharingHandlers()
	s.RegisterEnvironmentHandlers()
	s.RegisterEnvironmentVariableHandlers()
	s.RegisterDocsHandlers()
	s.RegisterFaviconHandler()
}

func (s *Server) RegisterHealthHandler() {
	s.router.GET("/health", s.Health)
}

func (s *Server) RegisterHomeHandler() {
	s.router.GET("/", s.Home)
}

func (s *Server) RegisterDocsHandlers() {
	s.router.GET("/openapi.json", s.OpenAPI)
	s.router.GET("/docs", s.Docs)
}

func (s *Server) RegisterFaviconHandler() {
	s.router.GET("/favicon.ico", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
}

func (s *Server) RegisterAuthHandlers() {
	auth := JWTAuthMiddleware(s.jwtSecret)
	s.router.POST("/auth/register", s.Register)
	s.router.POST("/auth/login", s.Login)
	s.router.GET("/auth/profile", auth(s.GetProfile))
}

func (s *Server) RegisterProjectHandlers() {
	auth := JWTAuthMiddleware(s.jwtSecret)
	s.router.POST("/projects", auth(s.CreateProject))
	s.router.GET("/projects/:id", auth(s.GetProject))
	s.router.GET("/projects", auth(s.ListProjects))
	s.router.PUT("/projects/:id", auth(s.UpdateProject))
	s.router.DELETE("/projects/:id", auth(s.DeleteProject))
}

func (s *Server) RegisterProjectSharingHandlers() {
	auth := JWTAuthMiddleware(s.jwtSecret)
	s.router.POST("/projects/:id/members", auth(s.AddUserToProject))
	s.router.DELETE("/projects/:id/members/:user_id", auth(s.RemoveUserFromProject))
	s.router.PUT("/projects/:id/members/:user_id", auth(s.UpdateUserRole))
	s.router.GET("/projects/:id/members", auth(s.GetProjectUsers))
	s.router.GET("/user/projects", auth(s.ListUserProjects))
}

func (s *Server) RegisterEnvironmentHandlers() {
	auth := JWTAuthMiddleware(s.jwtSecret)
	s.router.POST("/projects/:project_id/environments", auth(s.CreateEnvironment))
	s.router.GET("/projects/:project_id/environments/:id", auth(s.GetEnvironment))
	s.router.GET("/projects/:project_id/environments", auth(s.ListEnvironments))
	s.router.PUT("/projects/:project_id/environments/:id", auth(s.UpdateEnvironment))
	s.router.DELETE("/projects/:project_id/environments/:id", auth(s.DeleteEnvironment))
}

func (s *Server) RegisterEnvironmentVariableHandlers() {
	auth := JWTAuthMiddleware(s.jwtSecret)
	s.router.POST("/projects/:project_id/environments/:environment_id/variables", auth(s.CreateEnvironmentVariable))
	s.router.GET("/projects/:project_id/environments/:environment_id/variables/:id", auth(s.GetEnvironmentVariable))
	s.router.GET("/projects/:project_id/environments/:environment_id/variables", auth(s.ListEnvironmentVariables))
	s.router.PUT("/projects/:project_id/environments/:environment_id/variables/:id", auth(s.UpdateEnvironmentVariable))
	s.router.DELETE("/projects/:project_id/environments/:environment_id/variables/:id", auth(s.DeleteEnvironmentVariable))
}
