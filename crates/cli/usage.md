# Envoy CLI Usage Guide

## Command groups (6 top-level)

```
envoy manage ...   # server, auth, projects, envs, invites, members, links, status
envoy vars ...     # daily variables: pull, push (files only)
envoy link         # link this directory (defaults env to `dev`)
envoy use <env>    # switch default environment (switch-only)
envoy unlink       # remove link for cwd
envoy list         # read-only inventory: envs + keys + file paths
```

Everything is stored in one global file, `~/.envoy/config.json`
(`server_url`, `token`, plus a `links` map). Nothing is written into your
project directories.

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
envoy manage connect --url http://localhost:8080
envoy manage init        # first user becomes instance admin (fresh server only)
envoy manage login
envoy manage invites create someone@example.com [--name Someone]   # admin creates a one-time invite
envoy manage accept --token <token>                # invitee joins with their own password
envoy manage status
```

## Context (per directory)

```bash
envoy link [--project demo]   # defaults to the `dev` environment (created with every project)
envoy use prod                # only to switch away from `dev`
envoy manage links
envoy unlink
```

With a link in place, daily commands just work against `dev` — no `use` needed.

## Daily variables (`envoy vars ...`)

```bash
envoy vars push [--env dev] [-f server/.env]   # scans linked dir for env files by default
envoy vars pull [--env dev] [-f single.env]    # writes each stored file_path by default
```

Files are the interface: `push` scans the linked directory subtree (no depth
cap, skipping `.git/`, `node_modules/`, `target/`, etc.) for `.env`,
`.env.*`, and `*.env` (excluding `*.example`), uploads each with its
repo-relative path (`server/.env`), and fails if the same key appears in two
files of one environment. `pull` groups by stored `file_path` and writes each
file back under the link dir. `-f` is a one-off single-file override in both
directions and does not change tracking.

Mutating commands print the resolved `project/env` target
(e.g. `Pushed to 'demo/dev'`).

## Read-only inventory (`envoy list`)

```bash
envoy list [--project demo] [--env prod] [--show-values]
```

Lists every environment in the project with keys grouped by stored
`file_path`. Values are hidden by default (safe to share); pass
`--show-values` to print `KEY=VALUE`. Read-only: only `GET
/projects/{id}/environments` and `GET /environments/{id}/variables`
(viewer role suffices).

### `.env` file format

Both `push` and `pull` follow standard dotenv rules (same as docker
compose / python-dotenv / node dotenv): quotes, backslash escapes, the
`export` prefix, inline ` #` comments, BOM, and multiline double-quoted
values are all honoured. Malformed lines fail loudly naming the file and
line instead of uploading junk, and empty values are rejected (the server
requires non-empty values).

One rule deserves emphasis: unquoted and double-quoted `$VAR`/`${VAR}`
are expanded against your **local** environment at upload time. To store
a literal dollar sign, single-quote the value:

```bash
PASSWORD='s3cr$et'   # stored literally as s3cr$et
PASSWORD=s3cr$et     # $et is expanded (usually to empty) — not what you want
```

Duplicate keys resolve last-wins with a warning naming the key. Files
written by `pull` quote values as needed (spaces, `#`, `$`, quotes,
newlines), so a pulled file pushes back byte-identical.

## Management (`envoy manage ...`)

```bash
envoy manage projects create   # auto-creates a `dev` environment
envoy manage projects list
envoy manage envs create [--project demo]   # the ONLY way to add envs
envoy manage envs list [--project demo]
envoy manage invites create someone@example.com
envoy manage invites list
envoy manage invites revoke <invite-id>
envoy manage members search <email>
envoy manage members list <project>
envoy manage members add <project> <email>       # prompts for role
envoy manage members remove <project> <email>
```

Adding a new environment (explicit, never implicit — a misspelled `--env`
fails instead of creating a stray env):

```bash
envoy manage envs create --project demo   # prompts for the name, e.g. prod
envoy use prod
envoy vars push --env prod -f .env
```

Misspelled names fail closed with a suggestion:

```bash
$ envoy vars pull --env prdo
Error: environment 'prdo' not found. Did you mean 'prod'? — run 'envoy manage envs list' to see available environments
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

1. `envoy manage status` first when something looks wrong (server/link/login for cwd).
2. Use `--project/--env` in CI; use `link`/`use` locally.
