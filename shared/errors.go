package shared

import "errors"

// Common API errors used throughout the system.
var (
	// ErrUnauthorized indicates the request requires authentication.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidToken indicates the provided JWT token is invalid.
	ErrInvalidToken = errors.New("invalid token")

	// ErrExpiredToken indicates the provided JWT token has expired.
	ErrExpiredToken = errors.New("expired token")

	// ErrMalformedToken indicates the JWT token is malformed.
	ErrMalformedToken = errors.New("malformed token")

	// ErrInvalidSignature indicates the JWT token has an invalid signature.
	ErrInvalidSignature = errors.New("invalid token signature")

	// ErrAccessDenied indicates the user does not have permission for the requested operation.
	ErrAccessDenied = errors.New("access denied")

	// ErrProjectAccessDenied indicates the user does not have access to the specified project.
	ErrProjectAccessDenied = errors.New("project access denied")

	// ErrNotMember indicates the user is not a member of the project.
	ErrNotMember = errors.New("user is not a project member")

	// ErrInvalidRole indicates an invalid role was specified.
	ErrInvalidRole = errors.New("invalid role")

	// ErrBadRequest indicates the request is malformed or invalid.
	ErrBadRequest = errors.New("bad request")

	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrInternalError indicates an internal server error occurred.
	ErrInternalError = errors.New("internal error")

	// ErrConflict indicates the request conflicts with existing data.
	ErrConflict = errors.New("conflict")

	// ErrDuplicateEmail indicates an email address is already registered.
	ErrDuplicateEmail = errors.New("email already registered")

	// ErrInvalidCredentials indicates the provided username or password is incorrect.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrNoToken indicates no authentication token is available.
	ErrNoToken = errors.New("no token available")
)
