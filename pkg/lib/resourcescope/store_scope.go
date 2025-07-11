package resourcescope

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/google/uuid"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func (s *Store) NewScope(options *NewScopeOptions) *Scope {
	now := s.Clock.NowUTC()
	return &Scope{
		ID:          uuid.NewString(),
		CreatedAt:   now,
		UpdatedAt:   now,
		ResourceID:  options.ResourceID,
		Scope:       options.Scope,
		Description: options.Description,
	}
}

func (s *Store) CreateScope(ctx context.Context, scope *Scope) error {
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_resource_scope")).
		Columns(
			"id",
			"created_at",
			"updated_at",
			"resource_id",
			"scope",
			"description",
		).
		Values(
			scope.ID,
			scope.CreatedAt,
			scope.UpdatedAt,
			scope.ResourceID,
			scope.Scope,
			scope.Description,
		)

	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) {
			if pqError.Code == "23505" {
				return ErrScopeDuplicate
			}
		}
		return err
	}

	return nil
}

func (s *Store) UpdateScope(ctx context.Context, options *UpdateScopeOptions) error {
	now := s.Clock.NowUTC()

	q := s.SQLBuilder.Update(s.SQLBuilder.TableName("_auth_resource_scope")).
		Set("updated_at", now).
		Where("id = ?", options.ID)

	if options.NewDesc != nil {
		if *options.NewDesc == "" {
			q = q.Set("description", nil)
		} else {
			q = q.Set("description", *options.NewDesc)
		}
	}

	result, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) {
			if pqError.Code == "23505" {
				return ErrScopeDuplicate
			}
		}
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrScopeNotFound
	}

	return nil
}

func (s *Store) GetScopeByID(ctx context.Context, id string) (*Scope, error) {
	q := s.selectScopeQuery().Where("id = ?", id)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrScopeNotFound
		}
		return nil, err
	}

	sc, err := s.scanScope(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrScopeNotFound
		}
		return nil, err
	}

	return sc, nil
}

func (s *Store) DeleteScope(ctx context.Context, id string) error {
	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_resource_scope")).
		Where("id = ?", id)

	result, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrScopeNotFound
	}

	return nil
}

func (s *Store) GetManyScopes(ctx context.Context, ids []string) ([]*Scope, error) {
	q := s.selectScopeQuery().Where("id = ANY (?)", pq.Array(ids))
	return s.queryScopes(ctx, q)
}

func (s *Store) ListScopes(ctx context.Context, resourceID string) ([]*Scope, error) {
	q := s.selectScopeQuery().Where("resource_id = ?", resourceID)
	return s.queryScopes(ctx, q)
}

func (s *Store) queryScopes(ctx context.Context, q db.SelectBuilder) ([]*Scope, error) {
	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scopes []*Scope
	for rows.Next() {
		sc, err := s.scanScope(rows)
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, sc)
	}

	return scopes, nil
}

func (s *Store) selectScopeQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"id",
			"created_at",
			"updated_at",
			"resource_id",
			"scope",
			"description",
		).
		From(s.SQLBuilder.TableName("_auth_resource_scope"))
}

func (s *Store) scanScope(scanner db.Scanner) (*Scope, error) {
	sc := &Scope{}

	err := scanner.Scan(
		&sc.ID,
		&sc.CreatedAt,
		&sc.UpdatedAt,
		&sc.ResourceID,
		&sc.Scope,
		&sc.Description,
	)
	if err != nil {
		return nil, err
	}

	return sc, nil
}
