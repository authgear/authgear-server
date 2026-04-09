package service

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type CollaboratorServiceStore interface {
	ListCollaborators(ctx context.Context, appID string) ([]*model.Collaborator, error)
	GetCollaborator(ctx context.Context, id string) (*model.Collaborator, error)
	GetCollaboratorByAppAndUser(ctx context.Context, appID string, userID string) (*model.Collaborator, error)
	NewCollaborator(appID string, userID string, role model.CollaboratorRole) *model.Collaborator
	CreateCollaborator(ctx context.Context, c *model.Collaborator) error
	DeleteCollaborator(ctx context.Context, c *model.Collaborator) error
}

type CollaboratorStore struct {
	Clock       clock.Clock
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *CollaboratorStore) selectCollaborator() sq.SelectBuilder {
	return s.SQLBuilder.Select(
		"id",
		"app_id",
		"user_id",
		"created_at",
		"role",
	).From(s.SQLBuilder.TableName("_portal_app_collaborator"))
}

func (s *CollaboratorStore) ListCollaborators(ctx context.Context, appID string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("app_id = ?", appID)
	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborators []*model.Collaborator
	for rows.Next() {
		c, err := scanCollaborator(rows)
		if err != nil {
			return nil, err
		}
		collaborators = append(collaborators, c)
	}

	return collaborators, nil
}

func (s *CollaboratorStore) GetCollaborator(ctx context.Context, id string) (*model.Collaborator, error) {
	row, err := s.SQLExecutor.QueryRowWith(ctx, s.selectCollaborator().Where("id = ?", id))
	if err != nil {
		return nil, err
	}
	return scanCollaborator(row)
}

func (s *CollaboratorStore) GetCollaboratorByAppAndUser(ctx context.Context, appID string, userID string) (*model.Collaborator, error) {
	row, err := s.SQLExecutor.QueryRowWith(ctx, s.selectCollaborator().Where("app_id = ? AND user_id = ?", appID, userID))
	if err != nil {
		return nil, err
	}
	return scanCollaborator(row)
}

func (s *CollaboratorStore) NewCollaborator(appID string, userID string, role model.CollaboratorRole) *model.Collaborator {
	return &model.Collaborator{
		ID:        uuid.New(),
		AppID:     appID,
		UserID:    userID,
		CreatedAt: s.Clock.NowUTC(),
		Role:      role,
	}
}

func (s *CollaboratorStore) CreateCollaborator(ctx context.Context, c *model.Collaborator) error {
	_, err := s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Columns("id", "app_id", "user_id", "created_at", "role").
		Values(c.ID, c.AppID, c.UserID, c.CreatedAt, c.Role),
	)
	if isUniqueViolation(err) {
		return portalservice.ErrCollaboratorDuplicate
	}
	return err
}

func (s *CollaboratorStore) DeleteCollaborator(ctx context.Context, c *model.Collaborator) error {
	result, err := s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("id = ?", c.ID),
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return portalservice.ErrCollaboratorNotFound
	}
	return nil
}

type CollaboratorService struct {
	GlobalDatabase AppServiceDatabase
	Store          CollaboratorServiceStore
	AdminAPI       *AdminAPIService
}

func (s *CollaboratorService) ListCollaborators(ctx context.Context, appID string) ([]siteadmin.Collaborator, error) {
	var collaborators []*model.Collaborator
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		var err error
		collaborators, err = s.Store.ListCollaborators(ctx, appID)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(collaborators) == 0 {
		return []siteadmin.Collaborator{}, nil
	}

	userIDs := make([]string, len(collaborators))
	for i, c := range collaborators {
		userIDs[i] = c.UserID
	}

	emailMap, err := s.AdminAPI.ResolveUserEmails(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]siteadmin.Collaborator, len(collaborators))
	for i, c := range collaborators {
		result[i] = siteadmin.Collaborator{
			Id:        c.ID,
			AppId:     c.AppID,
			UserId:    c.UserID,
			UserEmail: emailMap[c.UserID],
			Role:      siteadmin.CollaboratorRole(c.Role),
			CreatedAt: c.CreatedAt,
		}
	}
	return result, nil
}

func (s *CollaboratorService) AddCollaborator(ctx context.Context, appID string, userEmail string) (*siteadmin.Collaborator, error) {
	userIDs, err := s.AdminAPI.FindUserIDsByEmail(ctx, userEmail)
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		return nil, portalservice.ErrCollaboratorNotFound
	}

	targetUserID := userIDs[0]
	var newCollaborator *model.Collaborator
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		_, err := s.Store.GetCollaboratorByAppAndUser(ctx, appID, targetUserID)
		if err == nil {
			return portalservice.ErrCollaboratorDuplicate
		}
		if !errors.Is(err, portalservice.ErrCollaboratorNotFound) {
			return err
		}

		newCollaborator = s.Store.NewCollaborator(appID, targetUserID, model.CollaboratorRoleEditor)
		return s.Store.CreateCollaborator(ctx, newCollaborator)
	})
	if err != nil {
		return nil, err
	}

	return &siteadmin.Collaborator{
		Id:        newCollaborator.ID,
		AppId:     newCollaborator.AppID,
		UserId:    newCollaborator.UserID,
		UserEmail: userEmail,
		Role:      siteadmin.CollaboratorRole(newCollaborator.Role),
		CreatedAt: newCollaborator.CreatedAt,
	}, nil
}

func (s *CollaboratorService) RemoveCollaborator(ctx context.Context, appID string, collaboratorID string) error {
	return s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		c, err := s.Store.GetCollaborator(ctx, collaboratorID)
		if err != nil {
			return err
		}
		if c.AppID != appID {
			return portalservice.ErrCollaboratorNotFound
		}
		return s.Store.DeleteCollaborator(ctx, c)
	})
}

func scanCollaborator(scanner interface{ Scan(dest ...any) error }) (*model.Collaborator, error) {
	c := &model.Collaborator{}
	err := scanner.Scan(
		&c.ID,
		&c.AppID,
		&c.UserID,
		&c.CreatedAt,
		&c.Role,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, portalservice.ErrCollaboratorNotFound
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
