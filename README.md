# Envoy
A secure environment variable management tool built using Go. Project consists of a REST API and a CLI.

## Core Features:

- Variable Creation: Securely store environment variables using encryption.
- Project and Environment Management: Categorize environment variables into projects and environments (e.g., development, staging, production).
- Secure Sharing: Share environment variables securely with team members through access controls.

## Installation

Install/Update the latest version:
```bash
go install ytsruh.com/envoy@latest
```

Install a specific version:
```bash
go install ytsruh.com/envoy@v0.0.1
```

## Version Management

Check your installed version:
```bash
envoy version
```

Envoy uses semantic versioning with git tags. When you install using `@latest`, Go automatically installs the highest version tag.
