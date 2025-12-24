# AGENTS.md

## Overview
A Go server built using the Echo framework, SQLite, SQLC library, Goose for database migrations, and internal packages.

## Architecture
- **Server package** (`pkg/server/`): Server core, initialization, routing
  - Server struct in `server.go`, Builder pattern in `builder.go`
  - Routes registered in `routes.go` with explicit closures
- **Handlers package** (`pkg/server/handlers/`): All HTTP handlers
  - Handlers use `HandlerContext` struct for dependencies
  - Request/response utilities in `pkg/response`
- **Middleware package** (`pkg/server/middleware/`): All Echo middleware
  - Auth middleware: JWT validation, user context
  - Access control middleware: Owner/Editor/Viewer checks
  - General middleware: logging, CORS, rate limiting
- **Database package** (`pkg/database/`): Database service, migrations, and SQLC-generated code
- **Utils package** (`pkg/utils/`): Shared utilities (JWT, password hashing, validation, access control)

## Code style & conventions
- Formatting: gofmt / goimports for Go; Always run formatters before commits.
- Imports: group standard library, external, internal; use goimports to auto-fix.
- Dependencies: use go mod tidy to manage dependencies but avoid using 3rd party packages wherever possible. Always ask for approval before adding new dependencies.
- Types: use interfaces to allow for mocking and testing.
- Error handling: check and return errors explicitly; wrap with fmt.Errorf("context: %w", err) when adding context.
- Logging: use structured logs in server packages; avoid fmt.Println in production code.
- Tests: When writing new code always write tests & keep tests deterministic, use table-driven tests for Go. Run `go test -race` when adding concurrency.

## Agent rules
- Follow repository toolchain (use go tools for Go).
- If adding env variables, create .env.example with placeholders.
- Handlers are functions in `pkg/server/handlers/` with signature:
  `func Handler(c echo.Context, ctx *handlers.HandlerContext) error`
- HandlerContext contains Queries, JWTSecret, AccessControl
- Tests use shared utilities in `handlers/testutil.go`
- Tests use standard `httptest` with Echo context setup.

## Quick checklist for agents
- Run unit tests (single test when debugging) by running 'make test'
- Add/modify .env.example if new secrets needed
- Update AGENTS.md when new tooling is added
