package resourcescope

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/google/uuid"
)

func (s *Store) NewResource(options *NewResourceOptions) *Resource {
	now := s.Clock.NowUTC()
	return &Resource{
		ID:        uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
		URI:       options.URI,
		Name:      options.Name,
	}
}

func (s *Store) CreateResource(ctx context.Context, r *Resource) error {
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_resource")).
		Columns(
			"id",
			"created_at",
			"updated_at",
			"uri",
			"name",
		).
		Values(
			r.ID,
			r.CreatedAt,
			r.UpdatedAt,
			r.URI,
			r.Name,
		)

	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) {
			if pqError.Code == "23505" {
				return ErrResourceDuplicateURI
			}
		}
		return err
	}

	return nil
}

func (s *Store) UpdateResource(ctx context.Context, options *UpdateResourceOptions) error {
	now := s.Clock.NowUTC()

	q := s.SQLBuilder.Update(s.SQLBuilder.TableName("_auth_resource")).
		Set("updated_at", now).
		Where("id = ?", options.ID)

	if options.NewName != nil {
		if *options.NewName == "" {
			q = q.Set("name", nil)
		} else {
			q = q.Set("name", *options.NewName)
		}
	}

	result, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) {
			if pqError.Code == "23505" {
				return ErrResourceDuplicateURI
			}
		}
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *Store) DeleteResource(ctx context.Context, id string) error {
	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_resource")).
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
		return ErrResourceNotFound
	}

	return nil
}

func (s *Store) GetResourceByID(ctx context.Context, id string) (*Resource, error) {
	q := s.selectResourceQuery().Where("id = ?", id)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	r, err := s.scanResource(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	return r, nil
}

func (s *Store) selectResourceQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"id",
			"created_at",
			"updated_at",
			"uri",
			"name",
		).
		From(s.SQLBuilder.TableName("_auth_resource"))
}

func (s *Store) scanResource(scanner db.Scanner) (*Resource, error) {
	r := &Resource{}

	err := scanner.Scan(
		&r.ID,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.URI,
		&r.Name,
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}
