# Envoy CLI Usage Guide

## Overview

Daily variables live under `vars`; context and management are separate
groups, so `envoy --help` reads clean. Everything is stored in one global
file, `~/.envoy/config.json` (`server_url`, `token`, plus a `links` map).
Nothing is written into your project directories.

Each link maps one directory to one project/env:

```json
"links": {
  "/Users/you/api": {"project_id": "uuid", "project_name": "demo", "env_name": "dev", "git_repo": "acme/demo"}
}
```

Precedence for project/env resolution everywhere:
`--project/--env` flag > `ENVOY_PROJECT`/`ENVOY_ENV` > global `links` entry for
cwd (with parent-dir inheritance) > `dev` fallback > single-env auto-pick >
git-remote auto-match > interactive prompt.

## Setup (once)

```bash
envoy setup connect --url http://localhost:8080
envoy setup init        # first user becomes instance admin (fresh server only)
envoy setup login
envoy admin invite someone@example.com   # admin creates a one-time invite
envoy setup accept --token <token>       # invitee joins with their own password
envoy status            # server + link + login for cwd
```

## Context (per directory)

```bash
# Link this directory (defaults to the `dev` environment, created with every project)
envoy link [--project demo]
envoy use prod                     # only to switch away from `dev`
envoy links                        # list all directory links
envoy unlink                       # remove link for cwd
```

With a link in place, daily commands just work against `dev` — no `use` needed.

## Daily variables (`envoy vars ...`)

```bash
envoy vars push [--env dev] [-f .env]
envoy vars pull [--env dev]            # -> .env.<env> by default, -f to override
envoy vars set API_KEY secret [--env dev]
envoy vars get API_KEY [--env dev]     # prints value only (script-friendly)
envoy vars unset API_KEY [--env dev] [--yes]
envoy vars list [--env dev]            # KEY=VALUE lines
```

```bash
envoy link --project demo
envoy vars push            # demo/dev from global links map
envoy vars pull
envoy vars list
```

Mutating commands always print the resolved `project/env` target
(e.g. `Pushed to 'demo/dev'`), so a stale `ENVOY_PROJECT` or wrong directory
is visible before damage spreads. Each directory resolves independently.

## Rare management

```bash
envoy projects create   # auto-creates a `dev` environment
envoy projects list
envoy envs create [--project demo]   # name only; the ONLY way to add envs
envoy envs list [--project demo]
```

Adding a new environment (explicit, never implicit — a misspelled `--env`
fails instead of creating a stray env):

```bash
envoy envs create --project demo   # prompts for the name, e.g. prod
envoy use prod                     # switch this directory to it (updates global links map)
envoy vars push --env prod -f .env # populate it
```

Misspelled names fail closed with a suggestion:

```bash
$ envoy vars list --env prdo
Error: environment 'prdo' not found. Did you mean 'prod'? — run 'envoy envs list' to see available environments
```

The server API is flat:

```
POST/GET /projects
GET/PUT/DELETE /projects/{id}
POST/GET /projects/{id}/environments
GET/PUT/DELETE /environments/{id}
POST/GET /environments/{id}/variables
GET/PUT/DELETE /variables/{id}
```

Name-or-ID resolution (`demo`, `prod`, `API_KEY`) is done client-side via
list + match.

## Tips

1. `envoy status` first when something looks wrong (server/link/login for cwd).
2. Use `--project/--env` in CI; use `link`/`use` locally.
3. `get` prints only the value, so `export KEY=$(envoy vars get KEY --env prod)` works.
4. `list` prints `KEY=VALUE` lines, safe to pipe to files.
