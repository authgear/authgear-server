package service

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/db"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var ErrCollaboratorNotFound = apierrors.NotFound.WithReason("CollaboratorNotFound").New("collaborator not found")
var ErrCollaboratorUnauthorized = apierrors.Unauthorized.WithReason("CollaboratorUnauthorized").New("collaborator unauthorized")
var ErrCollaboratorSelfDeletion = apierrors.Forbidden.WithReason("CollaboratorSelfDeletion").New("cannot remove self from collaborator")

type CollaboratorService struct {
	Context     context.Context
	Clock       clock.Clock
	SQLBuilder  *db.SQLBuilder
	SQLExecutor *db.SQLExecutor
}

func (s *CollaboratorService) selectCollaborator() sq.SelectBuilder {
	return s.SQLBuilder.Select(
		"id",
		"app_id",
		"user_id",
		"created_at",
	).From(s.SQLBuilder.FullTableName("app_collaborator"))
}

func (s *CollaboratorService) ListCollaborators(appID string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("app_id = ?", appID)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cs []*model.Collaborator
	for rows.Next() {
		c, err := scanCollaborator(rows)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}

	return cs, nil
}

func (s *CollaboratorService) ListCollaboratorsByUser(userID string) ([]*model.Collaborator, error) {
	q := s.selectCollaborator().Where("user_id = ?", userID)
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cs []*model.Collaborator
	for rows.Next() {
		c, err := scanCollaborator(rows)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}

	return cs, nil
}

func (s *CollaboratorService) NewCollaborator(appID string, userID string) *model.Collaborator {
	now := s.Clock.NowUTC()
	c := &model.Collaborator{
		ID:        uuid.New(),
		AppID:     appID,
		UserID:    userID,
		CreatedAt: now,
	}
	return c
}

func (s *CollaboratorService) CreateCollaborator(c *model.Collaborator) error {
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Insert(s.SQLBuilder.FullTableName("app_collaborator")).
		Columns(
			"id",
			"app_id",
			"user_id",
			"created_at",
		).
		Values(
			c.ID,
			c.AppID,
			c.UserID,
			c.CreatedAt,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *CollaboratorService) GetCollaborator(id string) (*model.Collaborator, error) {
	q := s.selectCollaborator().Where("id = ?", id)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return scanCollaborator(row)
}

func (s *CollaboratorService) DeleteCollaborator(c *model.Collaborator) error {
	sessionInfo := session.GetValidSessionInfo(s.Context)
	if sessionInfo == nil {
		return ErrCollaboratorUnauthorized
	}
	if c.UserID == sessionInfo.UserID {
		return ErrCollaboratorSelfDeletion
	}

	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Delete(s.SQLBuilder.FullTableName("app_collaborator")).
		Where("id = ?", c.ID),
	)
	if err != nil {
		return err
	}

	return nil
}

func scanCollaborator(scan db.Scanner) (*model.Collaborator, error) {
	c := &model.Collaborator{}

	err := scan.Scan(
		&c.ID,
		&c.AppID,
		&c.UserID,
		&c.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCollaboratorNotFound
	} else if err != nil {
		return nil, err
	}

	return c, nil
}
