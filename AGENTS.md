# AGENTS.md

## Overview
A Go server built using the Echo framework, SQLite, SQLC library, Goose for database migrations, and internal packages.

## Architecture
- **Server package** (`pkg/server/`): Contains all HTTP handlers, middleware, and routing logic
  - Handlers are methods on the `Server` struct and accept Echo's `echo.Context`
  - Access control is handled via Echo middleware (`RequireProjectOwner`, `RequireProjectEditor`, `RequireProjectViewer`)
  - Routes are registered in `routes.go` with simplified handler registration
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
- Handlers are methods on the Server struct in `pkg/server/` package, accepting `echo.Context`.
- Tests use standard `httptest` with Echo context setup for handler testing.

## Quick checklist for agents
- Run unit tests (single test when debugging) by running 'make test'
- Add/modify .env.example if new secrets needed
- Update AGENTS.md when new tooling is added
