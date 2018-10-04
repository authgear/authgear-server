package pq

import (
	"fmt"

	sq "github.com/lann/squirrel"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type RoleStore struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func NewRoleStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *RoleStore {
	return &RoleStore{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func (s RoleStore) CreateRoles(roles []string) error {
	s.logger.Debugf("createRole %v", roles)
	for _, role := range roles {
		builder := s.sqlBuilder.Insert(s.sqlBuilder.TableName("role")).Columns(
			"id",
		).Values(
			role,
		)
		_, err := s.sqlExecutor.ExecWith(builder)
		if db.IsUniqueViolated(err) {
			return skyerr.NewError(skyerr.Duplicated,
				fmt.Sprintf("Duplicated roles %v", role))
		}
	}
	return nil
}

func (s RoleStore) QueryRoles(roles []string) ([]string, error) {
	if roles == nil {
		return nil, nil
	}

	roleArgs := make([]interface{}, len(roles))
	for i, v := range roles {
		roleArgs[i] = interface{}(v)
	}
	builder := s.sqlBuilder.Select("id").
		From(s.sqlBuilder.TableName("role")).
		Where("id IN ("+sq.Placeholders(len(roles))+")", roleArgs...)
	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	existedRoles := []string{}
	for rows.Next() {
		var roleStr string
		if err := rows.Scan(&roleStr); err != nil {
			panic(err)
		}
		existedRoles = append(existedRoles, roleStr)
	}
	return existedRoles, nil
}
