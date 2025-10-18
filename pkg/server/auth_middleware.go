package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/pkg/utils"
)

// UserContextKey is the key used to store user claims in the context
const UserContextKey = "user"

// JWTAuthMiddleware creates middleware that validates JWT tokens and adds user info to context
func JWTAuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Missing authorization header",
				})
			}

			// Check if it's a Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid authorization header format. Expected: Bearer <token>",
				})
			}

			token := parts[1]

			// Validate the JWT
			claims, err := utils.ValidateJWT(token, jwtSecret)
			if err != nil {
				if err == utils.ErrExpiredToken {
					return c.JSON(http.StatusUnauthorized, map[string]string{
						"error": "Token has expired",
					})
				}
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

			// Add claims to context
			c.Set(UserContextKey, claims)

			// Continue to next handler
			return next(c)
		}
	}
}

// GetUserFromContext retrieves the JWT claims from the context
func GetUserFromContext(c echo.Context) (*utils.JWTClaims, bool) {
	user := c.Get(UserContextKey)
	if user == nil {
		return nil, false
	}

	claims, ok := user.(*utils.JWTClaims)
	return claims, ok
}
