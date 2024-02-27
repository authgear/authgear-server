package rolesgroups

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/lib/pq"
)

func (s *Store) scanUser(scanner db.Scanner) (string, error) {
	userId := ""
	err := scanner.Scan(&userId)
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (s *Store) selectUserQuery() db.SelectBuilder {
	return s.SQLBuilder.Select("id").From(s.SQLBuilder.TableName("_auth_user"))
}

func (s *Store) GetManyUsersByIds(ids []string) ([]string, error) {
	q := s.selectUserQuery().Where("id = ANY (?)", pq.Array(ids))
	return s.queryUsers(q)
}

func (s *Store) queryUsers(q db.SelectBuilder) ([]string, error) {
	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		r, err := s.scanUser(rows)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, r)
	}

	return userIDs, nil
}
