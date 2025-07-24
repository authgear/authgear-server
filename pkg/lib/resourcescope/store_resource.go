package resourcescope

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/google/uuid"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	databaseutil "github.com/authgear/authgear-server/pkg/util/databaseutil"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func (s *Store) NewResource(options *NewResourceOptions) *Resource {
	now := s.Clock.NowUTC()
	return &Resource{
		ID:          uuid.NewString(),
		CreatedAt:   now,
		UpdatedAt:   now,
		ResourceURI: options.URI.Value,
		Name:        options.Name,
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
			r.ResourceURI,
			r.Name,
		)

	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		if databaseutil.IsDuplicateKeyError(err) {
			return ErrResourceDuplicateURI
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
	q := s.selectResourceQuery("r").Where("r.id = ?", id)

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
	q := s.selectResourceQuery("r").Where("r.uri = ?", uri)

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
	q := s.selectResourceQuery("r").Where("r.id = ANY (?)", pq.Array(ids))
	return s.queryResources(ctx, q)
}

func (s *Store) GetClientResource(ctx context.Context, clientID, resourceID string) (*Resource, error) {
	q := s.selectResourceQuery("r").
		Join(s.SQLBuilder.TableName("_auth_client_resource"), "acr", "acr.resource_id = r.id").
		Where("r.id = ? AND acr.client_id = ?", resourceID, clientID)
	resources, err := s.queryResources(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(resources) == 0 {
		return nil, ErrResourceNotAssociatedWithClient
	}
	return resources[0], nil
}

type storeListResourceResult struct {
	Items      []*Resource
	Offset     uint64
	TotalCount uint64
}

func (s *Store) ListResources(ctx context.Context, options *ListResourcesOptions, pageArgs graphqlutil.PageArgs) (*storeListResourceResult, error) {
	q := s.selectResourceQuery("r").
		OrderBy("r.created_at DESC")
	q = s.applyListResourcesOptions(q, "r", options)

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
	q := s.SQLBuilder.Select("COUNT(*)").From(s.SQLBuilder.TableName("_auth_resource"), "r")
	q = s.applyListResourcesOptions(q, "r", options)

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

func (s *Store) applyListResourcesOptions(q db.SelectBuilder, authResourceAlias string, options *ListResourcesOptions) db.SelectBuilder {
	if options.SearchKeyword != "" {
		q = q.Where(fmt.Sprintf("(%s.uri ILIKE ('%%' || ? || '%%') OR %s.name ILIKE ('%%' || ? || '%%'))", authResourceAlias, authResourceAlias),
			options.SearchKeyword,
			options.SearchKeyword,
		)
	}
	if options.ClientID != "" {
		q = q.Join(
			s.SQLBuilder.TableName("_auth_client_resource"),
			"acr",
			fmt.Sprintf("acr.resource_id = %s.id", authResourceAlias),
		)
		q = q.Where("acr.client_id = ?", options.ClientID)
	}
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

func (s *Store) selectResourceQuery(alias string) db.SelectBuilder {
	aliasedColumn := func(col string) string {
		return fmt.Sprintf("%s.%s", alias, col)
	}
	return s.SQLBuilder.
		Select(
			aliasedColumn("id"),
			aliasedColumn("created_at"),
			aliasedColumn("updated_at"),
			aliasedColumn("uri"),
			aliasedColumn("name"),
		).
		From(s.SQLBuilder.TableName("_auth_resource"), alias)
}

func (s *Store) scanResource(scanner db.Scanner) (*Resource, error) {
	r := &Resource{}

	err := scanner.Scan(
		&r.ID,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.ResourceURI,
		&r.Name,
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *Store) AddResourceToClientID(ctx context.Context, resourceID, clientID string) error {
	now := s.Clock.NowUTC()
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_client_resource")).
		Columns("id", "created_at", "updated_at", "client_id", "resource_id").
		Values(uuid.NewString(), now, now, clientID, resourceID).
		Suffix("ON CONFLICT DO NOTHING")

	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) RemoveResourceFromClientID(ctx context.Context, resourceID, clientID string) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_client_resource")).
		Where("client_id = ? AND resource_id = ?", clientID, resourceID)
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}
	return nil
}
