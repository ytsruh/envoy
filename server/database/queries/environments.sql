-- name: CreateEnvironment :one
INSERT INTO environments (id, project_id, name, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, project_id, name, description, created_at, updated_at, deleted_at;

-- name: GetEnvironment :one
SELECT id, project_id, name, description, created_at, updated_at, deleted_at
FROM environments
WHERE id = ? AND deleted_at IS NULL;

-- name: ListEnvironmentsByProject :many
SELECT id, project_id, name, description, created_at, updated_at, deleted_at
FROM environments
WHERE project_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateEnvironment :one
UPDATE environments
SET name = ?, description = ?, updated_at = ?
WHERE id = ? AND deleted_at IS NULL
RETURNING id, project_id, name, description, created_at, updated_at, deleted_at;

-- name: DeleteEnvironment :exec
UPDATE environments
SET deleted_at = ?
WHERE id = ? AND deleted_at IS NULL;

-- name: GetAccessibleEnvironment :one
SELECT e.id, e.project_id, e.name, e.description, e.created_at, e.updated_at, e.deleted_at
FROM environments e
WHERE e.id = ? AND e.deleted_at IS NULL
AND (EXISTS (
    SELECT 1 FROM projects p 
    WHERE p.id = e.project_id AND p.deleted_at IS NULL
    AND (p.owner_id = ? OR EXISTS (
        SELECT 1 FROM project_users pu 
        WHERE pu.project_id = p.id AND pu.user_id = ?
    ))
));

-- name: CanUserModifyEnvironment :one
SELECT COUNT(*) as count
FROM environments e
INNER JOIN projects p ON e.project_id = p.id
LEFT JOIN project_users pu ON p.id = pu.project_id
WHERE e.id = ? AND e.deleted_at IS NULL AND p.deleted_at IS NULL
AND (
    p.owner_id = ? OR 
    (pu.user_id = ? AND pu.role = 'editor')
);