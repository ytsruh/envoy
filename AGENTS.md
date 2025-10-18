# AGENTS.md

## Overview
A Go server built using the standard library, SQLite, SQLC library, Goose for database migrations, and internal packages.
Frontend is a React application built using TypeScript and React within the /client directory. Frontend toolchain uses Bun and does not have access to Node, NPM or NodeJS.

## Code style & conventions
- Formatting: gofmt / goimports for Go; Prettier for frontend. Always run formatters before commits.
- Imports: group standard library, external, internal; use goimports to auto-fix.
- Dependencies: use go mod tidy to manage dependencies but avoid using 3rd party packages wherever possible. Always ask for approval before adding new dependencies.
- Types: use interfaces to allow for mocking and testing.
- Naming: use camelCase for JS/TS, MixedCaps (CamelCase) for Go exported identifiers, short receiver names (r, s), avoid underscores.
- Error handling: check and return errors explicitly; wrap with fmt.Errorf("context: %w", err) when adding context.
- Logging: use structured logs in server packages; avoid fmt.Println in production code.
- Tests: When writing new code always write tests & keep tests deterministic, use table-driven tests for Go. Run `go test -race` when adding concurrency.
- Commits: small atomic commits; clear messages describing why.

## Agent rules
- Follow repository toolchain (use go tools for Go, Bun tooling in client/).
- If adding env variables, create .env.example with placeholders.

## Quick checklist for agents
- Run unit tests (single test when debugging)
- Add/modify .env.example if new secrets needed
- Update AGENTS.md when new tooling is added
