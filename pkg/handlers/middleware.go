package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ytsruh.com/envoy/pkg/utils"
)

// Helper function to extract user claims from request context
func getUserClaims(r *http.Request) (*utils.JWTClaims, error) {
	user := r.Context().Value("user")
	if user == nil {
		return nil, fmt.Errorf("user not found in context")
	}

	claims, ok := user.(*utils.JWTClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse user claims")
	}

	return claims, nil
}

// Helper function to send standardized error responses
func sendErrorResponse(w http.ResponseWriter, code int, message error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message.Error()})
}

// Helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, code int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}

// Helper function to extract environment ID from URL path
func getEnvironmentIDFromPath(r *http.Request) (int64, error) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[3] == "" {
		return 0, fmt.Errorf("environment ID is required")
	}
	return strconv.ParseInt(pathParts[3], 10, 64)
}

// Access control middleware factories
func RequireProjectOwner(accessControl utils.AccessControlService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := getUserClaims(r)
			if err != nil {
				sendErrorResponse(w, http.StatusUnauthorized, err)
				return
			}

			projectID, err := getProjectIDFromPath(r)
			if err != nil {
				sendErrorResponse(w, http.StatusBadRequest, err)
				return
			}

			if err := accessControl.RequireOwner(r.Context(), projectID, claims.UserID); err != nil {
				sendErrorResponse(w, http.StatusForbidden, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireProjectEditor(accessControl utils.AccessControlService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := getUserClaims(r)
			if err != nil {
				sendErrorResponse(w, http.StatusUnauthorized, err)
				return
			}

			projectID, err := getProjectIDFromPath(r)
			if err != nil {
				sendErrorResponse(w, http.StatusBadRequest, err)
				return
			}

			if err := accessControl.RequireEditor(r.Context(), projectID, claims.UserID); err != nil {
				sendErrorResponse(w, http.StatusForbidden, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireProjectViewer(accessControl utils.AccessControlService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := getUserClaims(r)
			if err != nil {
				sendErrorResponse(w, http.StatusUnauthorized, err)
				return
			}

			projectID, err := getProjectIDFromPath(r)
			if err != nil {
				sendErrorResponse(w, http.StatusBadRequest, err)
				return
			}

			if err := accessControl.RequireViewer(r.Context(), projectID, claims.UserID); err != nil {
				sendErrorResponse(w, http.StatusForbidden, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
