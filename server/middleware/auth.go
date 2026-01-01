package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"ytsruh.com/envoy/server/utils"
	shared "ytsruh.com/envoy/shared"
)

const UserContextKey = "user"

func JWTAuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Missing authorization header",
				})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid authorization header format. Expected: Bearer <token>",
				})
			}

			token := parts[1]

			claims, err := utils.ValidateJWT(token, jwtSecret)
			if err != nil {
				if err == shared.ErrExpiredToken {
					return c.JSON(http.StatusUnauthorized, map[string]string{
						"error": "Token has expired",
					})
				}
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

			c.Set(UserContextKey, claims)

			return next(c)
		}
	}
}

func GetUserFromContext(c echo.Context) (*utils.JWTClaims, bool) {
	user := c.Get(UserContextKey)
	if user == nil {
		return nil, false
	}

	claims, ok := user.(*utils.JWTClaims)
	return claims, ok
}
