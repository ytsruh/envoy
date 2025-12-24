-- name: AddUserToProject :one
INSERT INTO project_users (project_id, user_id, role, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
RETURNING id, project_id, user_id, role, created_at, updated_at;

-- name: RemoveUserFromProject :exec
DELETE FROM project_users
WHERE project_id = ? AND user_id = ?;

-- name: UpdateUserRole :exec
UPDATE project_users
SET role = ?, updated_at = ?
WHERE project_id = ? AND user_id = ?;

-- name: GetProjectUsers :many
SELECT pu.id, pu.project_id, pu.user_id, pu.role, pu.created_at, pu.updated_at
FROM project_users pu
INNER JOIN users u ON pu.user_id = u.id
WHERE pu.project_id = ?
ORDER BY pu.created_at ASC;

-- name: GetUserProjects :many
SELECT DISTINCT p.id, p.name, p.description, p.git_repo, p.owner_id, p.created_at, p.updated_at, p.deleted_at
FROM projects p
LEFT JOIN project_users pu ON p.id = pu.project_id
WHERE (p.owner_id = ? OR pu.user_id = ?) AND p.deleted_at IS NULL
ORDER BY p.created_at DESC;

-- name: GetProjectMembership :one
SELECT pu.id, pu.project_id, pu.user_id, pu.role, pu.created_at, pu.updated_at
FROM project_users pu
WHERE pu.project_id = ? AND pu.user_id = ?;

-- name: IsProjectOwner :one
SELECT COUNT(*) as count
FROM projects
WHERE id = ? AND owner_id = ? AND deleted_at IS NULL;

-- name: GetAccessibleProject :one
SELECT p.id, p.name, p.description, p.git_repo, p.owner_id, p.created_at, p.updated_at, p.deleted_at
FROM projects p
WHERE p.id = ? AND p.deleted_at IS NULL
AND (p.owner_id = ? OR EXISTS (
    SELECT 1 FROM project_users pu 
    WHERE pu.project_id = p.id AND pu.user_id = ?
));

-- name: CanUserModifyProject :one
SELECT COUNT(*) as count
FROM projects p
LEFT JOIN project_users pu ON p.id = pu.project_id
WHERE p.id = ? AND p.deleted_at IS NULL
AND (
    p.owner_id = ? OR 
    (pu.user_id = ? AND pu.role = 'editor')
);

-- name: GetProjectMemberRole :one
SELECT pu.role
FROM project_users pu
WHERE pu.project_id = ? AND pu.user_id = ?;