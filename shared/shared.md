# Shared Package Overview
Package shared provides common types, errors, and utilities used by both the CLI client and server. It enforces API contracts between the client and server by defining shared types for all HTTP requests and responses.

The shared package includes:
- Type aliases for IDs and enums (UserID, ProjectID, EnvironmentID, Role, etc.)
- Conversion helpers for database types (sql.NullString <-> *string)
- All API error definitions used throughout the system
- Request type definitions with validation tags
- Response type definitions with Timestamp support for JSON serialization
- Version variable for application versioning (injected at build time)

## Usage
```
	import "ytsruh.com/envoy/shared"

	// Use shared types in handlers
	var req shared.CreateProjectRequest
	
	// Use shared errors
	return shared.ErrInvalidToken
	
	// Use conversion helpers
	s := shared.NullStringToStringPtr(ns)

	// Get application version
	v := shared.Version
```

## Version

The `Version` variable contains the application version. It's read from Go's build metadata at runtime:
- Release builds: semver from git tag (e.g., "v1.0.0")
- Development builds: "(devel)" with helpful installation message
