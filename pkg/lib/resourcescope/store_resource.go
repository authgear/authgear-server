package resourcescope

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/google/uuid"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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
		Where("uri = ?", options.ResourceURI)

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

func (s *Store) DeleteResourceByURI(ctx context.Context, uri string) error {
	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_resource")).
		Where("uri = ?", uri)

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

func (s *Store) GetResourceByURI(ctx context.Context, uri string) (*Resource, error) {
	q := s.selectResourceQuery().Where("uri = ?", uri)

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

func (s *Store) GetManyResources(ctx context.Context, ids []string) ([]*Resource, error) {
	q := s.selectResourceQuery().Where("id = ANY (?)", pq.Array(ids))
	return s.queryResources(ctx, q)
}

type storeListResourceResult struct {
	Items      []*Resource
	Offset     uint64
	TotalCount uint64
}

func (s *Store) ListResources(ctx context.Context, options *ListResourcesOptions, pageArgs graphqlutil.PageArgs) (*storeListResourceResult, error) {
	q := s.selectResourceQuery().
		OrderBy("uri ASC")
	q = s.applyListResourcesOptions(q, options)

	q, offset, err := db.ApplyPageArgs(q, pageArgs)
	if err != nil {
		return nil, err
	}

	resources, err := s.queryResources(ctx, q)
	if err != nil {
		return nil, err
	}

	totalCount, err := s.countResources(ctx, options)
	if err != nil {
		return nil, err
	}

	return &storeListResourceResult{
		Items:      resources,
		Offset:     offset,
		TotalCount: totalCount,
	}, nil
}

func (s *Store) countResources(ctx context.Context, options *ListResourcesOptions) (uint64, error) {
	q := s.SQLBuilder.Select("COUNT(*)").From(s.SQLBuilder.TableName("_auth_resource"))
	q = s.applyListResourcesOptions(q, options)

	var count uint64
	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return 0, err
	}
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) applyListResourcesOptions(q db.SelectBuilder, options *ListResourcesOptions) db.SelectBuilder {
	if options != nil && options.SearchKeyword != "" {
		q = q.Where("(uri ILIKE ('%' || ? || '%') OR name ILIKE ('%' || ? || '%'))", options.SearchKeyword, options.SearchKeyword)
	}
	// TODO(tung): Implement ClientID filtering if/when resources are associated with clients
	return q
}

func (s *Store) queryResources(ctx context.Context, q db.SelectBuilder) ([]*Resource, error) {
	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*Resource
	for rows.Next() {
		r, err := s.scanResource(rows)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
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
