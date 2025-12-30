package utils

import (
	"context"
	"database/sql"
	"fmt"

	database "ytsruh.com/envoy/server/database/generated"
	shared "ytsruh.com/envoy/shared"
)

type AccessControlService interface {
	RequireOwner(ctx context.Context, projectID int64, userID string) error
	RequireEditor(ctx context.Context, projectID int64, userID string) error
	RequireViewer(ctx context.Context, projectID int64, userID string) error
	GetRole(ctx context.Context, projectID int64, userID string) (string, error)
}

type AccessControlServiceImpl struct {
	queries database.Querier
}

func NewAccessControlService(queries database.Querier) AccessControlService {
	return &AccessControlServiceImpl{
		queries: queries,
	}
}

func (s *AccessControlServiceImpl) RequireOwner(ctx context.Context, projectID int64, userID string) error {
	count, err := s.queries.IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to check ownership: %w", err)
	}
	if count == 0 {
		return shared.ErrAccessDenied
	}
	return nil
}

func (s *AccessControlServiceImpl) RequireEditor(ctx context.Context, projectID int64, userID string) error {
	count, err := s.queries.CanUserModifyProject(ctx, database.CanUserModifyProjectParams{
		ID:      projectID,
		OwnerID: userID,
		UserID:  userID,
	})
	if err != nil {
		return fmt.Errorf("failed to check editor permissions: %w", err)
	}
	if count == 0 {
		return shared.ErrAccessDenied
	}
	return nil
}

func (s *AccessControlServiceImpl) RequireViewer(ctx context.Context, projectID int64, userID string) error {
	_, err := s.queries.GetAccessibleProject(ctx, database.GetAccessibleProjectParams{
		ID:      projectID,
		OwnerID: userID,
		UserID:  userID,
	})
	if err == sql.ErrNoRows {
		return shared.ErrAccessDenied
	}
	if err != nil {
		return fmt.Errorf("failed to check viewer permissions: %w", err)
	}
	return nil
}

func (s *AccessControlServiceImpl) GetRole(ctx context.Context, projectID int64, userID string) (string, error) {
	// Check if user is owner first
	ownerCount, err := s.queries.IsProjectOwner(ctx, database.IsProjectOwnerParams{
		ID:      projectID,
		OwnerID: userID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to check ownership: %w", err)
	}
	if ownerCount > 0 {
		return "owner", nil
	}

	// Check project membership
	membership, err := s.queries.GetProjectMembership(ctx, database.GetProjectMembershipParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err == sql.ErrNoRows {
		return "", shared.ErrNotMember
	}
	if err != nil {
		return "", fmt.Errorf("failed to get project membership: %w", err)
	}

	return membership.Role, nil
}
