package server

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/pkg/utils"
)

func RequireProjectOwner(accessControl utils.AccessControlService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := GetUserFromContext(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			}

			projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
			}

			if err := accessControl.RequireOwner(c.Request().Context(), projectID, claims.UserID); err != nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
			}

			return next(c)
		}
	}
}

func RequireProjectEditor(accessControl utils.AccessControlService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := GetUserFromContext(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			}

			projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
			}

			if err := accessControl.RequireEditor(c.Request().Context(), projectID, claims.UserID); err != nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
			}

			return next(c)
		}
	}
}

func RequireProjectViewer(accessControl utils.AccessControlService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := GetUserFromContext(c)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			}

			projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid project ID"})
			}

			if err := accessControl.RequireViewer(c.Request().Context(), projectID, claims.UserID); err != nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
			}

			return next(c)
		}
	}
}
