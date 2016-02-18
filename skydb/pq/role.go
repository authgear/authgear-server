package pq

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	log "github.com/Sirupsen/logrus"
	sq "github.com/lann/squirrel"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
	"github.com/oursky/skygear/utils"
)

const batchUserRoleInsertTemplate = `
INSERT INTO {{.UserRoleTable}} (user_id, role_id)
SELECT $1, id
FROM {{.RoleTable}}
WHERE id IN ({{range $i, $_ := .In}}{{inDollar $i}}{{end}});
`

var batchUserRoleInsert = template.Must(
	template.New("batchUserRoleInsert").Funcs(template.FuncMap{
		"inDollar": func(n int) string {
			dollar := "$" + strconv.Itoa(n+2)
			if n > 0 {
				dollar = "," + dollar
			}
			return dollar
		},
	}).Parse(batchUserRoleInsertTemplate))

type batchUserRole struct {
	UserRoleTable string
	RoleTable     string
	In            []string
}

func (c *conn) batchUserRoleSQL(id string, roles []string) (string, []interface{}) {
	b := bytes.Buffer{}
	batchUserRoleInsert.Execute(&b, batchUserRole{
		c.tableName("_user_role"),
		c.tableName("_role"),
		roles,
	})

	args := make([]interface{}, len(roles)+1)
	args[0] = id
	for i := range roles {
		args[i+1] = roles[i]
	}
	return b.String(), args

}

func (c *conn) GetAdminRoles() ([]string, error) {
	// TODO: Implement actual query
	return []string{
		"admin",
	}, nil
}

func (c *conn) SetAdminRoles(roles []string) error {
	log.Debugf("SetAdminRoles %v", roles)
	c.ensureRole(roles)
	return c.setRoleType(roles, "is_admin")
}

func (c *conn) SetDefaultRoles(roles []string) error {
	log.Debugf("SetDefaultRoles %v", roles)
	c.ensureRole(roles)
	return c.setRoleType(roles, "by_default")
}

func (c *conn) setRoleType(roles []string, col string) error {
	resetSQL := psql.Update(c.tableName("_role")).
		Where(col+" = ?", true).Set(col, false)
	_, err := c.ExecWith(resetSQL)
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

	updateSQL := psql.Update(c.tableName("_role")).
		Where("id IN ("+sq.Placeholders(len(roles))+")", roleArgs...).
		Set(col, true)
	_, err = c.ExecWith(updateSQL)
	if err != nil {
		return err
	}
	return nil
}

func (c *conn) UpdateUserRoles(userinfo *skydb.UserInfo) error {
	log.Debugf("UpdateRoles %v", userinfo)
	builder := psql.Delete(c.tableName("_user_role")).Where("user_id = ?", userinfo.ID)
	_, err := c.ExecWith(builder)
	if err != nil {
		return skyerr.NewError(skyerr.ConstraintViolated,
			fmt.Sprintf("Fails to reset user roles %v", userinfo.ID))
	}
	if len(userinfo.Roles) == 0 {
		return nil
	}
	sql, args := c.batchUserRoleSQL(userinfo.ID, userinfo.Roles)
	result, err := c.Exec(sql, args...)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != int64(len(userinfo.Roles)) {
		absenceRoles, err := c.ensureRole(userinfo.Roles)
		if err != nil {
			return err
		}
		sql, args := c.batchUserRoleSQL(userinfo.ID, absenceRoles)
		_, err = c.Exec(sql, args...)
		if err != nil {
			return err
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

func (c *conn) ensureRole(roles []string) ([]string, error) {
	if roles == nil || len(roles) == 0 {
		return nil, nil
	}
	existedRole, err := c.queryRoles(roles)
	if err != nil {
		return nil, err
	}
	if len(existedRole) == len(roles) {
		return nil, nil
	}
	log.Debugf("Diffing the roles not exist in DB")
	absenceRoles := utils.StringSliceExcept(roles, existedRole)
	return absenceRoles, c.createRoles(absenceRoles)
}
