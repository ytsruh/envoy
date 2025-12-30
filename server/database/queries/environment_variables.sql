-- name: CreateEnvironmentVariable :one
INSERT INTO environment_variables (id, environment_id, key, value, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING id, environment_id, key, value, description, created_at, updated_at;

-- name: GetEnvironmentVariable :one
SELECT id, environment_id, key, value, description, created_at, updated_at
FROM environment_variables
WHERE id = ?;

-- name: ListEnvironmentVariablesByEnvironment :many
SELECT id, environment_id, key, value, description, created_at, updated_at
FROM environment_variables
WHERE environment_id = ?
ORDER BY created_at DESC;

-- name: UpdateEnvironmentVariable :one
UPDATE environment_variables
SET key = ?, value = ?, description = ?, updated_at = ?
WHERE id = ?
RETURNING id, environment_id, key, value, description, created_at, updated_at;

-- name: DeleteEnvironmentVariable :exec
DELETE FROM environment_variables
WHERE id = ?;

-- name: GetAccessibleEnvironmentVariable :one
SELECT ev.id, ev.environment_id, ev.key, ev.value, ev.description, ev.created_at, ev.updated_at
FROM environment_variables ev
INNER JOIN environments e ON ev.environment_id = e.id
WHERE ev.id = ? AND e.deleted_at IS NULL
AND (EXISTS (
    SELECT 1 FROM projects p 
    WHERE p.id = e.project_id AND p.deleted_at IS NULL
    AND (p.owner_id = ? OR EXISTS (
        SELECT 1 FROM project_users pu 
        WHERE pu.project_id = p.id AND pu.user_id = ?
    ))
));

-- name: CanUserModifyEnvironmentVariable :one
SELECT COUNT(*) as count
FROM environment_variables ev
INNER JOIN environments e ON ev.environment_id = e.id
INNER JOIN projects p ON e.project_id = p.id
LEFT JOIN project_users pu ON p.id = pu.project_id
WHERE ev.id = ? AND e.deleted_at IS NULL AND p.deleted_at IS NULL
AND (
    p.owner_id = ? OR 
    (pu.user_id = ? AND pu.role = 'editor')
);
