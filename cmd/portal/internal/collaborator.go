package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type AddCollaboratorOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	AppID          string
	UserID         string
	Role           model.CollaboratorRole
}

type AddCollaboratorResult string

const (
	AddCollaboratorResultNoop     AddCollaboratorResult = "noop"
	AddCollaboratorResultInserted AddCollaboratorResult = "inserted"
	AddCollaboratorResultUpdated  AddCollaboratorResult = "updated"
)

func AddCollaborator(ctx context.Context, options AddCollaboratorOptions) (result AddCollaboratorResult, err error) {
	db := openDB(options.DatabaseURL, options.DatabaseSchema)

	err = WithTx(ctx, db, func(tx *sql.Tx) error {
		err := checkAppExist(ctx, tx, options.AppID)
		if err != nil {
			return err
		}

		err = checkUserExist(ctx, tx, options.UserID)
		if err != nil {
			return err
		}

		row, err := getExistingCollaborator(ctx, tx, options.AppID, options.UserID)
		if err != nil {
			return err
		}

		if row == nil {
			err = insertCollaborator(ctx, tx, options.AppID, options.UserID, options.Role)
			if err != nil {
				return err
			}

			result = AddCollaboratorResultInserted
			return nil
		}

		if row.Role == options.Role {
			result = AddCollaboratorResultNoop
			return nil
		}

		err = updateCollaboratorRole(ctx, tx, row.ID, options.Role)
		if err != nil {
			return err
		}

		result = AddCollaboratorResultUpdated
		return nil
	})
	if err != nil {
		return
	}
	return
}

func checkAppExist(ctx context.Context, tx *sql.Tx, appID string) error {
	builder := newSQLBuilder().Select("id").
		From(pq.QuoteIdentifier("_portal_config_source")).
		Where("app_id = ?", appID)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	row := tx.QueryRowContext(ctx, query, args...)

	var outAppID string
	err = row.Scan(&outAppID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("app not found: %v", appID)
		}
		return err
	}

	return nil
}

func checkUserExist(ctx context.Context, tx *sql.Tx, userID string) error {
	builder := newSQLBuilder().Select("id").
		From(pq.QuoteIdentifier("_auth_user")).
		Where("id = ?", userID)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	row := tx.QueryRowContext(ctx, query, args...)

	var outAppID string
	err = row.Scan(&outAppID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user not found: %v", userID)
		}
		return err
	}

	return nil
}

type collaboratorRow struct {
	ID        string
	AppID     string
	UserID    string
	CreatedAt time.Time
	Role      model.CollaboratorRole
}

func getExistingCollaborator(ctx context.Context, tx *sql.Tx, appID string, userID string) (*collaboratorRow, error) {
	builder := newSQLBuilder().Select(
		"id",
		"app_id",
		"user_id",
		"created_at",
		"role",
	).
		From(pq.QuoteIdentifier("_portal_app_collaborator")).
		Where("app_id = ? AND user_id = ?", appID, userID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, query, args...)
	var out collaboratorRow

	err = row.Scan(
		&out.ID,
		&out.AppID,
		&out.UserID,
		&out.CreatedAt,
		&out.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func insertCollaborator(ctx context.Context, tx *sql.Tx, appID string, userID string, role model.CollaboratorRole) error {
	id := uuid.New()
	now := time.Now().UTC()

	builder := newSQLBuilder().Insert(pq.QuoteIdentifier("_portal_app_collaborator")).Columns(
		"id",
		"app_id",
		"user_id",
		"created_at",
		"role",
	).Values(
		id,
		appID,
		userID,
		now,
		role,
	)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func updateCollaboratorRole(ctx context.Context, tx *sql.Tx, id string, role model.CollaboratorRole) error {
	builder := newSQLBuilder().Update(pq.QuoteIdentifier("_portal_app_collaborator")).
		Set("role", role).
		Where("id = ?", id)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
