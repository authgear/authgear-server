package rolesgroups

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func (s *Store) scanUserID(scanner db.Scanner) (string, error) {
	userId := ""
	err := scanner.Scan(&userId)
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (s *Store) selectUserIDQuery() db.SelectBuilder {
	return s.SQLBuilder.Select("id").From(s.SQLBuilder.TableName("_auth_user"))
}

func (s *Store) GetUserByID(id string) (string, error) {
	q := s.selectUserIDQuery().Where("id = ?", id)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return "", err
	}

	r, err := s.scanUserID(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", api.ErrUserNotFound
		}
		return "", err
	}

	return r, nil
}

func (s *Store) GetManyUsersByIds(ids []string) ([]string, error) {
	q := s.selectUserIDQuery().Where("id = ANY (?)", pq.Array(ids))
	return s.queryUserIDs(q)
}

func (s *Store) queryUserIDs(q db.SelectBuilder) ([]string, error) {
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		r, err := s.scanUserID(rows)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, r)
	}

	return userIDs, nil
}
