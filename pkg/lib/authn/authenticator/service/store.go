package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
}

func (s *Store) Count(ctx context.Context, userID string) (uint64, error) {
	builder := s.SQLBuilder.
		Select("count(*)").
		Where("user_id = ?", userID).
		From(s.SQLBuilder.TableName("_auth_authenticator"))
	scanner, err := s.SQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return 0, err
	}

	var count uint64
	if err = scanner.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) ListRefsByUsers(ctx context.Context, userIDs []string, authenticatorType *model.AuthenticatorType, authenticatorKind *authenticator.Kind) ([]*authenticator.Ref, error) {
	builder := s.SQLBuilder.
		Select("id", "type", "user_id", "created_at", "updated_at").
		Where("user_id = ANY (?)", pq.Array(userIDs))

	if authenticatorType != nil && *authenticatorType != "" {
		builder = builder.Where("type = ?", authenticatorType)
	}

	if authenticatorKind != nil && *authenticatorKind != "" {
		builder = builder.Where("kind = ?", authenticatorKind)
	}

	builder = builder.From(s.SQLBuilder.TableName("_auth_authenticator"))

	return s.listRefs(ctx, builder)
}

func (s *Store) ListRefsByIDs(ctx context.Context, ids []string) ([]*authenticator.Ref, error) {
	builder := s.SQLBuilder.
		Select("id", "type", "user_id", "created_at", "updated_at").
		Where("id = ANY (?)", pq.Array(ids)).
		From(s.SQLBuilder.TableName("_auth_authenticator"))

	return s.listRefs(ctx, builder)
}

func (s *Store) GetRefByID(ctx context.Context, id string) (*authenticator.Ref, error) {
	builder := s.SQLBuilder.
		Select("id", "type", "user_id", "created_at", "updated_at").
		Where("id = ?", id).
		From(s.SQLBuilder.TableName("_auth_authenticator"))

	row, err := s.SQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return nil, err
	}

	ref, err := s.scan(row)
	if err != nil {
		return nil, err
	}

	return ref, nil
}

func (s *Store) listRefs(ctx context.Context, builder db.SelectBuilder) ([]*authenticator.Ref, error) {
	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*authenticator.Ref
	for rows.Next() {
		var ref *authenticator.Ref
		ref, err = s.scan(rows)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}

	return refs, nil
}

func (s *Store) scan(scanner db.Scanner) (*authenticator.Ref, error) {
	ref := &authenticator.Ref{}
	err := scanner.Scan(
		&ref.ID,
		&ref.Type,
		&ref.UserID,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	return ref, nil
}
