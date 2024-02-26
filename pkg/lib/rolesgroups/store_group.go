package rolesgroups

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func (s *Store) NewGroup(options *NewGroupOptions) *Group {
	now := s.Clock.NowUTC()
	return &Group{
		ID:          uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Key:         options.Key,
		Name:        options.Name,
		Description: options.Description,
	}
}

func (s *Store) CreateGroup(r *Group) error {
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_group")).
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
				err = ErrGroupDuplicateKey
			}
		}
		return err
	}

	return nil
}

func (s *Store) UpdateGroup(options *UpdateGroupOptions) error {
	now := s.Clock.NowUTC()

	q := s.SQLBuilder.Update(s.SQLBuilder.TableName("_auth_group")).
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
				err = ErrGroupDuplicateKey
			}
		}
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrGroupNotFound
	}

	return nil
}

func (s *Store) DeleteGroup(id string) error {
	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_group_role")).
		Where("group_id = ?", id)

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_user_group")).
		Where("group_id = ?", id)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_group")).
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
		return ErrGroupNotFound
	}

	return nil
}

func (s *Store) GetGroupByID(id string) (*Group, error) {
	q := s.selectGroupQuery().Where("id = ?", id)

	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	r, err := s.scanGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	return r, nil
}

func (s *Store) GetGroupByKey(key string) (*Group, error) {
	q := s.selectGroupQuery().Where("key = ?", key)

	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	r, err := s.scanGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	return r, nil
}

func (s *Store) selectGroupQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"id",
			"created_at",
			"updated_at",
			"key",
			"name",
			"description",
		).
		From(s.SQLBuilder.TableName("_auth_group"))
}

func (s *Store) scanGroup(scanner db.Scanner) (*Group, error) {
	r := &Group{}

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

func (s *Store) GetManyGroups(ids []string) ([]*Group, error) {
	q := s.selectGroupQuery().Where("id = ANY (?)", pq.Array(ids))
	return s.queryGroups(q)
}

func (s *Store) GetManyGroupsByKeys(keys []string) ([]*Group, error) {
	q := s.selectGroupQuery().Where("key = ANY (?)", pq.Array(keys))
	return s.queryGroups(q)
}

func (s *Store) queryGroups(q db.SelectBuilder) ([]*Group, error) {
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		r, err := s.scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, r)
	}

	return groups, nil
}
