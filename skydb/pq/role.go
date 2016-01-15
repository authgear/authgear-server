package pq

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	sq "github.com/lann/squirrel"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
	"github.com/oursky/skygear/utils"
)

func (c *conn) UpdateUserRoles(userinfo *skydb.UserInfo) error {
	log.Debugf("UpdateRoles %v", userinfo)
	builder := psql.Delete(c.tableName("_user_role")).Where("user_id = ?", userinfo.ID)
	_, err := c.ExecWith(builder)
	if err != nil {
		return skyerr.NewError(skyerr.ConstraintViolated,
			fmt.Sprintf("Fails to reset user roles %v", userinfo.ID))
	}
	for _, role := range userinfo.Roles {
		builder := psql.Insert(c.tableName("_user_role")).Columns(
			"user_id",
			"role_id",
		).Values(
			userinfo.ID,
			role,
		)
		_, err := c.ExecWith(builder)
		if err != nil {
			return skyerr.NewError(skyerr.ConstraintViolated,
				fmt.Sprintf("Duplicated user roles %v", role))
		}
	}
	return nil
}

func (c *conn) createRoles(roles []string) error {
	log.Debugf("createRole %v", roles)
	for _, role := range roles {
		builder := psql.Insert(c.tableName("_role")).Columns(
			"id",
		).Values(
			role,
		)
		_, err := c.ExecWith(builder)
		if isUniqueViolated(err) {
			return skyerr.NewError(skyerr.Duplicated,
				fmt.Sprintf("Duplicated roles %v", role))
		}
	}
	return nil
}

func (c *conn) queryRoles(roles []string) ([]string, error) {
	if roles == nil {
		return nil, nil
	}

	roleArgs := make([]interface{}, len(roles))
	for i, v := range roles {
		roleArgs[i] = interface{}(v)
	}
	builder := psql.Select("id").
		From(c.tableName("_role")).
		Where("id IN ("+sq.Placeholders(len(roles))+")", roleArgs...)
	rows, err := c.QueryWith(builder)
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

func (c *conn) ensureRole(roles []string) error {
	if roles == nil || len(roles) == 0 {
		return nil
	}
	existedRole, err := c.queryRoles(roles)
	if err != nil {
		return err
	}
	if len(existedRole) == len(roles) {
		return nil
	}
	log.Debugf("Diffing the roles not exist in DB")
	absenceRoles := utils.StrSliceWithout(roles, existedRole)
	return c.createRoles(absenceRoles)
}
