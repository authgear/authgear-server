package rolesgroups

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func (s *Store) NewRole(options *NewRoleOptions) *Role {
	now := s.Clock.NowUTC()
	return &Role{
		ID:          uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Key:         options.Key,
		Name:        options.Name,
		Description: options.Description,
	}
}

func (s *Store) CreateRole(r *Role) error {
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_role")).
		Columns(
			"id",
			"created_at",
			"updated_at",
			"key",
			"name",
			"description",
		).
		Values(
			r.ID,
			r.CreatedAt,
			r.UpdatedAt,
			r.Key,
			r.Name,
			r.Description,
		)

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) {
			// https://www.postgresql.org/docs/13/errcodes-appendix.html
			// 23505 is unique_violation
			if pqError.Code == "23505" {
				err = ErrRoleDuplicateKey
			}
		}
		return err
	}

	return nil
}

func (s *Store) UpdateRole(options *UpdateRoleOptions) error {
	now := s.Clock.NowUTC()

	q := s.SQLBuilder.Update(s.SQLBuilder.TableName("_auth_role")).
		Set("updated_at", now).
		Where("id = ?", options.ID)

	if options.NewKey != nil {
		q = q.Set("key", *options.NewKey)
	}

	if options.NewName != nil {
		if *options.NewName == "" {
			q = q.Set("name", nil)
		} else {
			q = q.Set("name", *options.NewName)
		}
	}

	if options.NewDescription != nil {
		if *options.NewDescription == "" {
			q = q.Set("description", nil)
		} else {
			q = q.Set("description", *options.NewDescription)
		}
	}

	result, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) {
			// https://www.postgresql.org/docs/13/errcodes-appendix.html
			// 23505 is unique_violation
			if pqError.Code == "23505" {
				err = ErrRoleDuplicateKey
			}
		}
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrRoleNotFound
	}

	return nil
}

func (s *Store) DeleteRole(id string) error {
	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_group_role")).
		Where("role_id = ?", id)

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_user_role")).
		Where("role_id = ?", id)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_role")).
		Where("id = ?", id)

	result, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrRoleNotFound
	}

	return nil
}

func (s *Store) GetRoleByID(id string) (*Role, error) {
	q := s.selectRoleQuery().Where("id = ?", id)

	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	r, err := s.scanRole(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	return r, nil
}

func (s *Store) GetRoleByKey(key string) (*Role, error) {
	q := s.selectRoleQuery().Where("key = ?", key)

	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	r, err := s.scanRole(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	return r, nil
}

func (s *Store) ListRoles(options *ListRolesOptions, pageArgs graphqlutil.PageArgs) ([]*Role, uint64, error) {
	q := s.selectRoleQuery().
		// Sort by key to ensure we have a stable order.
		OrderBy("key ASC")

	if options.KeyPrefix != "" {
		q = q.Where("key ILIKE (? || '%')", options.KeyPrefix)
	}

	q, offset, err := db.ApplyPageArgs(q, pageArgs)
	if err != nil {
		return nil, 0, err
	}

	roles, err := s.queryRoles(q)
	if err != nil {
		return nil, 0, err
	}

	return roles, offset, nil
}

func (s *Store) selectRoleQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"id",
			"created_at",
			"updated_at",
			"key",
			"name",
			"description",
		).
		From(s.SQLBuilder.TableName("_auth_role"))
}

func (s *Store) scanRole(scanner db.Scanner) (*Role, error) {
	r := &Role{}

	err := scanner.Scan(
		&r.ID,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.Key,
		&r.Name,
		&r.Description,
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *Store) GetManyRoles(ids []string) ([]*Role, error) {
	q := s.selectRoleQuery().Where("id = ANY (?)", pq.Array(ids))
	return s.queryRoles(q)
}

func (s *Store) GetManyRolesByKeys(keys []string) ([]*Role, error) {
	q := s.selectRoleQuery().Where("key = ANY (?)", pq.Array(keys))
	return s.queryRoles(q)
}

func (s *Store) queryRoles(q db.SelectBuilder) ([]*Role, error) {
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*Role
	for rows.Next() {
		r, err := s.scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}

	return roles, nil
}
