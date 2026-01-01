package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/server/utils"
)

type ProjectRole string

const (
	RoleOwner  ProjectRole = "owner"
	RoleEditor ProjectRole = "editor"
	RoleViewer ProjectRole = "viewer"
)

func RequireProjectRole(role ProjectRole, accessControl utils.AccessControlService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := GetUserFromContext(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			}

			projectID := c.Param("id")

			var err error
			switch role {
			case RoleOwner:
				err = accessControl.RequireOwner(c.Request().Context(), projectID, claims.UserID)
			case RoleEditor:
				err = accessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID)
			case RoleViewer:
				err = accessControl.RequireViewer(c.Request().Context(), projectID, claims.UserID)
			}

			if err != nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
			}

			return next(c)
		}
	}
}

func RequireProjectOwner(accessControl utils.AccessControlService) echo.MiddlewareFunc {
	return RequireProjectRole(RoleOwner, accessControl)
}

func RequireProjectEditor(accessControl utils.AccessControlService) echo.MiddlewareFunc {
	return RequireProjectRole(RoleEditor, accessControl)
}

func RequireProjectViewer(accessControl utils.AccessControlService) echo.MiddlewareFunc {
	return RequireProjectRole(RoleViewer, accessControl)
}
