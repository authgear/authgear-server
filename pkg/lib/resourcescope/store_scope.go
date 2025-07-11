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

func (s *Store) NewScope(resource *Resource, options *NewScopeOptions) *Scope {
	now := s.Clock.NowUTC()
	return &Scope{
		ID:          uuid.NewString(),
		CreatedAt:   now,
		UpdatedAt:   now,
		ResourceID:  resource.URI,
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
		if databaseutil.IsDuplicateKeyError(err) {
			return ErrScopeDuplicate
		}
		return err
	}

	return nil
}

func (s *Store) UpdateScope(ctx context.Context, options *UpdateScopeOptions) error {
	now := s.Clock.NowUTC()

	resource, err := s.GetResourceByURI(ctx, options.ResourceURI)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.Update(s.SQLBuilder.TableName("_auth_resource_scope")).
		Set("updated_at", now).
		Where("resource_id = ? AND scope = ?", resource.ID, options.Scope)

	if options.NewDesc != nil {
		if *options.NewDesc == "" {
			q = q.Set("description", nil)
		} else {
			q = q.Set("description", *options.NewDesc)
		}
	}

	result, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		if databaseutil.IsDuplicateKeyError(err) {
			return ErrScopeDuplicate
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
	q := s.selectScopeQuery("s").Where("s.id = ?", id)

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

func (s *Store) DeleteScope(ctx context.Context, resourceURI string, scope string) error {
	resource, err := s.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_resource_scope")).
		Where("resource_id = ? AND scope = ?", resource.ID, scope)

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
	q := s.selectScopeQuery("s").Where("s.id = ANY (?)", pq.Array(ids))
	return s.queryScopes(ctx, q)
}

type storeListScopeResult struct {
	Items      []*Scope
	Offset     uint64
	TotalCount uint64
}

func (s *Store) ListScopes(ctx context.Context, resourceID string, options *ListScopeOptions, pageArgs graphqlutil.PageArgs) (*storeListScopeResult, error) {
	q := s.selectScopeQuery("s").Where("s.resource_id = ?", resourceID)
	q = s.applyListScopesOptions(q, "s", options)
	q = q.OrderBy("s.scope ASC")

	q, offset, err := db.ApplyPageArgs(q, pageArgs)
	if err != nil {
		return nil, err
	}

	scopes, err := s.queryScopes(ctx, q)
	if err != nil {
		return nil, err
	}

	totalCount, err := s.countScopes(ctx, resourceID, options)
	if err != nil {
		return nil, err
	}

	return &storeListScopeResult{
		Items:      scopes,
		Offset:     offset,
		TotalCount: totalCount,
	}, nil
}

func (s *Store) countScopes(ctx context.Context, resourceID string, options *ListScopeOptions) (uint64, error) {
	q := s.SQLBuilder.Select("COUNT(*)").From(s.SQLBuilder.TableName("_auth_resource_scope"), "s")
	q = q.Where("s.resource_id = ?", resourceID)
	q = s.applyListScopesOptions(q, "s", options)

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

func (s *Store) applyListScopesOptions(q db.SelectBuilder, alias string, options *ListScopeOptions) db.SelectBuilder {
	if options == nil {
		return q
	}
	if options.SearchKeyword != "" {
		q = q.Where(fmt.Sprintf("(%s.scope ILIKE ('%%' || ? || '%%') OR %s.description ILIKE ('%%' || ? || '%%'))", alias, alias),
			options.SearchKeyword,
			options.SearchKeyword,
		)
	}
	if options.ClientID != "" {
		q = q.Join(
			s.SQLBuilder.TableName("_auth_client_resource_scope"),
			"acrs",
			fmt.Sprintf("acrs.scope_id = %s.id", alias),
		)
		q = q.Where("acrs.client_id = ?", options.ClientID)
	}
	return q
}

func (s *Store) GetScope(ctx context.Context, resourceURI string, scope string) (*Scope, error) {
	resource, err := s.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return nil, err
	}

	q := s.selectScopeQuery("s").Where("s.resource_id = ? AND s.scope = ?", resource.ID, scope)
	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}
	sc, err := s.scanScope(row)
	if err != nil {
		return nil, err
	}
	return sc, nil
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

func (s *Store) selectScopeQuery(alias string) db.SelectBuilder {
	aliasedColumn := func(col string) string {
		return alias + "." + col
	}
	return s.SQLBuilder.
		Select(
			aliasedColumn("id"),
			aliasedColumn("created_at"),
			aliasedColumn("updated_at"),
			aliasedColumn("resource_id"),
			aliasedColumn("scope"),
			aliasedColumn("description"),
		).
		From(s.SQLBuilder.TableName("_auth_resource_scope"), alias)
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

func (s *Store) GetScopeByResourceIDAndScope(ctx context.Context, resourceID, scope string) (*Scope, error) {
	q := s.selectScopeQuery("s").Where("s.resource_id = ? AND s.scope = ?", resourceID, scope)
	scopes, err := s.queryScopes(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(scopes) == 0 {
		return nil, ErrScopeNotFound
	}
	return scopes[0], nil
}

func (s *Store) AddScopeToClientID(ctx context.Context, resourceID, scopeID, clientID string) error {
	now := s.Clock.NowUTC()
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_client_resource_scope")).
		Columns("id", "created_at", "updated_at", "client_id", "resource_id", "scope_id").
		Values(uuid.NewString(), now, now, clientID, resourceID, scopeID).
		Suffix("ON CONFLICT DO NOTHING")
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	return err
}

func (s *Store) RemoveScopeFromClientID(ctx context.Context, scopeID, clientID string) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_client_resource_scope")).
		Where("client_id = ? AND scope_id = ?", clientID, scopeID)
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	return err
}

func (s *Store) ListClientScopesByResourceID(ctx context.Context, resourceID, clientID string) ([]*Scope, error) {
	q := s.selectScopeQuery("s").
		Join(s.SQLBuilder.TableName("_auth_client_resource_scope"), "acrs", "acrs.scope_id = s.id").
		Where("s.resource_id = ? AND acrs.client_id = ?", resourceID, clientID)
	return s.queryScopes(ctx, q)
}
