package pq

import (
	"fmt"

	sq "github.com/lann/squirrel"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type roleType string

const (
	roleTypeDefault = roleType("default")
)

type roleStore struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	logger      *logrus.Entry
}

func newRoleStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) *roleStore {
	return &roleStore{
		sqlBuilder:  builder,
		sqlExecutor: executor,
		logger:      logger,
	}
}

func NewRoleStore(builder db.SQLBuilder, executor db.SQLExecutor, logger *logrus.Entry) role.Store {
	return newRoleStore(builder, executor, logger)
}

func (s roleStore) CreateRoles(roles []string) error {
	s.logger.Debugf("createRole %v", roles)
	for _, role := range roles {
		builder := s.sqlBuilder.Insert(s.sqlBuilder.FullTableName("role")).Columns(
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

func (s roleStore) QueryRoles(roles []string) ([]role.Role, error) {
	if roles == nil {
		return nil, nil
	}

	if len(roles) == 0 {
		return []role.Role{}, nil
	}

	roleArgs := make([]interface{}, len(roles))
	for i, v := range roles {
		roleArgs[i] = interface{}(v)
	}
	builder := s.sqlBuilder.Select("id", "is_admin", "by_default").
		From(s.sqlBuilder.FullTableName("role")).
		Where("id IN ("+sq.Placeholders(len(roles))+")", roleArgs...)
	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	existedRoles := []role.Role{}
	for rows.Next() {
		role := role.Role{}
		if err := rows.Scan(&role.Name, &role.IsAdmin, &role.IsDefault); err != nil {
			panic(err)
		}
		existedRoles = append(existedRoles, role)
	}
	return existedRoles, nil
}

func (s roleStore) GetDefaultRoles() ([]string, error) {
	return s.getRolesByType(roleTypeDefault)
}

func (s roleStore) getRolesByType(rtype roleType) ([]string, error) {
	var col string
	switch rtype {
	case roleTypeDefault:
		col = "by_default"
	default:
		panic("Unknow role type")
	}
	builder := s.sqlBuilder.Select("id").
		From(s.sqlBuilder.FullTableName("role")).
		Where(col + " = true")
	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := []string{}
	for rows.Next() {
		var roleStr string
		if err := rows.Scan(&roleStr); err != nil {
			panic(err)
		}
		roles = append(roles, roleStr)
	}
	return roles, nil
}

func (s roleStore) SetAdminRoles(roles []string) error {
	return s.setRoleType(roles, "is_admin")
}

func (s roleStore) SetDefaultRoles(roles []string) error {
	return s.setRoleType(roles, "by_default")
}

func (s roleStore) setRoleType(roles []string, col string) error {
	resetSQL := s.sqlBuilder.
		Update(s.sqlBuilder.FullTableName("role")).
		Where(col+" = ?", true).Set(col, false)
	_, err := s.sqlExecutor.ExecWith(resetSQL)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		return nil
	}
	roleArgs := make([]interface{}, len(roles))
	for i, v := range roles {
		roleArgs[i] = interface{}(v)
	}

	updateSQL := s.sqlBuilder.
		Update(s.sqlBuilder.FullTableName("role")).
		Where("id IN ("+sq.Placeholders(len(roles))+")", roleArgs...).
		Set(col, true)
	_, err = s.sqlExecutor.ExecWith(updateSQL)
	if err != nil {
		return err
	}
	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ role.Store = &roleStore{}
)
