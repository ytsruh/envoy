# Phase 1: Server Changes (Backend)

## Overview
Update the database schema, SQL queries, and server API to support git repositories on projects.

---

## Step 1: Override Existing Migration

**File: `pkg/database/migrations/00001_initial_schema.sql`**

Replace the projects table definition:

```sql
CREATE TABLE projects (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  git_repo TEXT,
  owner_id text NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP DEFAULT NULL,
  FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE(owner_id, git_repo)
);
```

Changes:
- Removed global `UNIQUE` on `name`
- Added `git_repo` column
- Added `UNIQUE(owner_id, git_repo)` for per-user uniqueness

---

## Step 2: Update SQL Queries

**File: `pkg/database/queries/projects.sql`**

```sql
-- name: CreateProject :one
INSERT INTO projects (name, description, git_repo, owner_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at;

-- name: GetProject :one
SELECT id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at
FROM projects
WHERE id = ? AND deleted_at IS NULL;

-- name: UpdateProject :one
UPDATE projects
SET name = ?, description = ?, git_repo = ?, updated_at = ?
WHERE id = ? AND owner_id = ? AND deleted_at IS NULL
RETURNING id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at;

-- name: ListProjectsByOwner :many
SELECT id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at
FROM projects
WHERE owner_id = ? AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: DeleteProject :exec
UPDATE projects
SET deleted_at = ?
WHERE id = ? AND owner_id = ? AND deleted_at IS NULL;

-- name: GetProjectByGitRepo :one
SELECT id, name, description, git_repo, owner_id, created_at, updated_at, deleted_at
FROM projects
WHERE git_repo = ? AND owner_id = ? AND deleted_at IS NULL;
```

**File: `pkg/database/queries/project_users.sql`**

Update `GetUserProjects`:

```sql
-- name: GetUserProjects :many
SELECT DISTINCT p.id, p.name, p.description, p.git_repo, p.owner_id, p.created_at, p.updated_at, p.deleted_at
FROM projects p
LEFT JOIN project_users pu ON p.id = pu.project_id
WHERE (p.owner_id = ? OR pu.user_id = ?) AND p.deleted_at IS NULL
ORDER BY p.created_at DESC;
```

Update `GetAccessibleProject`:

```sql
-- name: GetAccessibleProject :one
SELECT p.id, p.name, p.description, p.git_repo, p.owner_id, p.created_at, p.updated_at, p.deleted_at
FROM projects p
WHERE p.id = ? AND p.deleted_at IS NULL
AND (p.owner_id = ? OR EXISTS (
    SELECT 1 FROM project_users pu
    WHERE pu.project_id = p.id AND pu.user_id = ?
));
```

Regenerate SQLC:
```bash
make generate
```

---

## Step 3: Update Project Handlers

**File: `pkg/server/handlers/projects.go`**

Update structs:

```go
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,project_name"`
	Description string `json:"description" validate:"max=500"`
	GitRepo     string `json:"git_repo"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" validate:"required,project_name"`
	Description string `json:"description" validate:"max=500"`
	GitRepo     string `json:"git_repo"`
}

type ProjectResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	GitRepo     string    `json:"git_repo"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

Update `CreateProject` to validate git_repo uniqueness:

```go
func CreateProject(c echo.Context, ctx *HandlerContext) error {
	var req CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// If git_repo is provided, check for uniqueness per user
	if req.GitRepo != "" {
		_, err := ctx.Queries.GetProjectByGitRepo(dbCtx, database.GetProjectByGitRepoParams{
			GitRepo: req.GitRepo,
			OwnerID: claims.UserID,
		})
		if err == nil {
			return SendErrorResponse(c, http.StatusConflict, fmt.Errorf("project with this git repository already exists"))
		} else if err != sql.ErrNoRows {
			return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check git repository"))
		}
	}

	now := time.Now()
	project, err := ctx.Queries.CreateProject(dbCtx, database.CreateProjectParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		GitRepo:     sql.NullString{String: req.GitRepo, Valid: req.GitRepo != ""},
		OwnerID:     claims.UserID,
		CreatedAt:   sql.NullTime{Time: now, Valid: true},
		UpdatedAt:   now,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to create project"))
	}

	resp := ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: NullStringToStringPtr(project.Description),
		GitRepo:     NullStringToString(project.GitRepo),
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.(time.Time),
	}

	return c.JSON(http.StatusCreated, resp)
}
```

Update `UpdateProject`:

```go
func UpdateProject(c echo.Context, ctx *HandlerContext) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid project ID"))
	}

	var req UpdateProjectRequest
	if err := c.Bind(&req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := utils.Validate(req); err != nil {
		return SendErrorResponse(c, http.StatusBadRequest, err)
	}

	claims, ok := middleware.GetUserFromContext(c)
	if !ok {
		return SendErrorResponse(c, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	canModify, err := ctx.Queries.CanUserModifyProject(dbCtx, database.CanUserModifyProjectParams{
		ID:      projectID,
		OwnerID: claims.UserID,
		UserID:  claims.UserID,
	})
	if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check permissions"))
	}
	if canModify == 0 {
		return SendErrorResponse(c, http.StatusForbidden, fmt.Errorf("access denied"))
	}

	originalProject, err := ctx.Queries.GetProject(dbCtx, projectID)
	if err == sql.ErrNoRows {
		return SendErrorResponse(c, http.StatusNotFound, fmt.Errorf("project not found"))
	} else if err != nil {
		return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch project"))
	}

	// If git_repo is being changed, check for uniqueness
	if req.GitRepo != "" && NullStringToString(originalProject.GitRepo) != req.GitRepo {
		_, err := ctx.Queries.GetProjectByGitRepo(dbCtx, database.GetProjectByGitRepoParams{
			GitRepo: req.GitRepo,
			OwnerID: claims.UserID,
		})
		if err == nil {
			return SendErrorResponse(c, http.StatusConflict, fmt.Errorf("project with this git repository already exists"))
		} else if err != sql.ErrNoRows {
			return SendErrorResponse(c, http.StatusInternalServerError, fmt.Errorf("failed to check git repository"))
		}
	}

	now := time.Now()
	updatedProject, err := ctx.Queries.UpdateProject(dbCtx, database.UpdateProjectParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		GitRepo:     sql.NullString{String: req.GitRepo, Valid: req.GitRepo != ""},
		UpdatedAt:   now,
		ID:          projectID,
		OwnerID:     originalProject.OwnerID,
	})

	resp := ProjectResponse{
		ID:          updatedProject.ID,
		Name:        updatedProject.Name,
		Description: NullStringToStringPtr(updatedProject.Description),
		GitRepo:     NullStringToString(updatedProject.GitRepo),
		OwnerID:     updatedProject.OwnerID,
		CreatedAt:   updatedProject.CreatedAt.Time,
		UpdatedAt:   updatedProject.UpdatedAt.(time.Time),
	}

	return c.JSON(http.StatusOK, resp)
}
```

Add helper function:
```go
func NullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
```

---

## Step 4: Update Environment Variable

**File: `.env.example`**

```env
# Database configuration
DB_PATH=./data/envoy.db

# JWT Secret for authentication (use a strong random string in production)
JWT_SECRET=your-secret-key-change-this-in-production

# Server URL for CLI client
ENVOY_SERVER_URL=http://localhost:8080
```

---

## Step 5: Test Server Changes

1. Reset database if needed
2. Run migrations
3. Start server: `make run`
4. Test via existing API or `docs` endpoint
5. Verify projects can be created with/without `git_repo`
6. Verify uniqueness constraint per-user works
7. Run `make test` to ensure existing tests still pass
