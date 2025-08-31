-- name: GetUser :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (
  id, name, email, password, created_at, updated_at, deleted_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET name = ?, email = ?, password = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = ?
WHERE id = ?;

-- name: HardDeleteUser :exec
DELETE FROM users
WHERE id = ?;
