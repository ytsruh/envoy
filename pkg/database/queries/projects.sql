-- name: CreateProject :one
INSERT INTO projects (name, description, git_repo, owner_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at;

-- name: GetProject :one
SELECT id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at
FROM projects
WHERE id = ? AND deleted_at IS NULL;

-- name: ListProjectsByOwner :many
SELECT id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at
FROM projects
WHERE owner_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET name = ?, description = ?, git_repo = ?, updated_at = ?
WHERE id = ? AND owner_id = ? AND deleted_at IS NULL
RETURNING id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at;

-- name: DeleteProject :exec
UPDATE projects
SET deleted_at = ?
WHERE id = ? AND owner_id = ? AND deleted_at IS NULL;

-- name: GetProjectByGitRepo :one
SELECT id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at
FROM projects
WHERE owner_id = ? AND git_repo = ? AND deleted_at IS NULL;