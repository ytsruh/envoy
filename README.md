# Envoy
A secure environment variable management tool built using Go. Project consists of a REST API and a CLI tool.

## Core Features:

- Variable Creation: Securely store environment variables using encryption.
- Project and Environment Management: Categorize environment variables into projects and environments (e.g., development, staging, production).
- Secure Sharing: Share environment variables securely with team members through access controls.

## Environment Variables

- `JWT_SECRET` - Secret key for JWT token signing (use a strong random string in production)
- `ENVOY_SERVER_URL` - URL of the Envoy server for the CLI client
- `DB_URL` - URL for the database
- `DB_TOKEN` - Token for the database
