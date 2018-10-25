package pq

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	sq "github.com/lann/squirrel"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func (s AuthInfoStore) AssignRoles(userIDs []string, roles []string) error {
	s.logger.Debugf("AssignRoles %v to %v", roles, userIDs)
	role.EnsureRole(s.roleStore, s.logger, roles)
	sql, args := s.assignUserRoleSQL(userIDs, roles)
	_, err := s.sqlExecutor.Exec(sql, args...)
	if err != nil {
		return err
	}
	return nil
}

func (s AuthInfoStore) GetRoles(userIDs []string) (map[string][]string, error) {
	userIDArgs := make([]interface{}, len(userIDs))
	for i, v := range userIDs {
		userIDArgs[i] = interface{}(v)
	}
	builder := s.sqlBuilder.Select("user_id", "role_id").
		From(s.sqlBuilder.FullTableName("user_role")).
		Where("user_id IN ("+sq.Placeholders(len(userIDs))+")", userIDArgs...)
	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roleMap := map[string][]string{}
	for _, eachUserID := range userIDs {
		// keep an empty array even no roles found for that user
		roleMap[eachUserID] = []string{}
	}
	for rows.Next() {
		userID := ""
		roleID := ""
		if err := rows.Scan(&userID, &roleID); err != nil {
			panic(err)
		}
		userRoleMap := roleMap[userID]
		roleMap[userID] = append(userRoleMap, roleID)
	}

	return roleMap, nil
}

func (s AuthInfoStore) RevokeRoles(userIDs []string, roles []string) error {
	s.logger.Debugf("RevokeRoles %v to %v", roles, userIDs)
	sql, args := s.revokeUserRoleSQL(userIDs, roles)
	_, err := s.sqlExecutor.Exec(sql, args...)
	if err != nil {
		return err
	}
	return nil
}

func (s AuthInfoStore) updateUserRoles(authinfo *authinfo.AuthInfo) error {
	s.logger.Debugf("UpdateRoles %v", authinfo)
	builder := s.sqlBuilder.Delete(s.sqlBuilder.FullTableName("user_role")).Where("user_id = ?", authinfo.ID)
	_, err := s.sqlExecutor.ExecWith(builder)
	if err != nil {
		return skyerr.NewError(skyerr.ConstraintViolated,
			fmt.Sprintf("Fails to reset user roles %v", authinfo.ID))
	}
	if len(authinfo.Roles) == 0 {
		return nil
	}
	sql, args := s.batchUserRoleSQL(authinfo.ID, authinfo.Roles)
	result, err := s.sqlExecutor.Exec(sql, args...)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != int64(len(authinfo.Roles)) {
		absenceRoles, err := role.EnsureRole(s.roleStore, s.logger, authinfo.Roles)
		if err != nil {
			return err
		}
		sql, args := s.batchUserRoleSQL(authinfo.ID, absenceRoles)
		_, err = s.sqlExecutor.Exec(sql, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

const assignUserRoleInsertTemplate = `
{{ $userLen := len .Users }}
INSERT INTO {{.UserRoleTable}} (user_id, role_id)
SELECT "user"."id", "role"."id"
FROM {{.UserTable}} AS "user", {{.RoleTable}} AS "role"
WHERE
  "user"."id" IN ({{range $i, $_ := .Users}}{{inDollar 0 $i}}{{end}})
  AND
  "role"."id" IN ({{range $i, $_ := .Roles}}{{inDollar $userLen $i}}{{end}})
  AND ("role"."id", "user"."id") NOT IN (
    SELECT role_id, user_id
  FROM {{.UserRoleTable}}
  WHERE
    user_id IN ({{range $i, $_ := .Users}}{{inDollar 0 $i}}{{end}})
  AND
    role_id IN ({{range $i, $_ := .Roles}}{{inDollar $userLen $i}}{{end}})
  );
`

var assignUserRoleInsert = template.Must(
	template.New("assignUserRole").Funcs(template.FuncMap{
		"inDollar": func(offset int, n int) string {
			dollar := "$" + strconv.Itoa(n+offset+1)
			if n > 0 {
				dollar = "," + dollar
			}
			return dollar
		},
	}).Parse(assignUserRoleInsertTemplate))

type assignUserRole struct {
	UserRoleTable string
	RoleTable     string
	UserTable     string
	Roles         []string
	Users         []string
}

func (s AuthInfoStore) assignUserRoleSQL(users []string, roles []string) (string, []interface{}) {
	b := bytes.Buffer{}
	assignUserRoleInsert.Execute(&b, assignUserRole{
		s.sqlBuilder.FullTableName("user_role"),
		s.sqlBuilder.FullTableName("role"),
		s.sqlBuilder.FullTableName("user"),
		roles,
		users,
	})

	args := make([]interface{}, len(users)+len(roles))
	for i := range users {
		args[i] = users[i]
	}
	offset := len(users)
	for i := range roles {
		args[i+offset] = roles[i]
	}
	return b.String(), args

}

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

func (s AuthInfoStore) batchUserRoleSQL(id string, roles []string) (string, []interface{}) {
	b := bytes.Buffer{}
	batchUserRoleInsert.Execute(&b, batchUserRole{
		s.sqlBuilder.FullTableName("user_role"),
		s.sqlBuilder.FullTableName("role"),
		roles,
	})

	args := make([]interface{}, len(roles)+1)
	args[0] = id
	for i := range roles {
		args[i+1] = roles[i]
	}
	return b.String(), args
}

const revokeUserRoleDeleteTemplate = `
{{ $userLen := len .Users }}
DELETE FROM {{.UserRoleTable}}
WHERE
  user_id IN ({{range $i, $_ := .Users}}{{inDollar 0 $i}}{{end}})
  AND
  role_id IN ({{range $i, $_ := .Roles}}{{inDollar $userLen $i}}{{end}})
;
`

var revokeUserRoleDelete = template.Must(
	template.New("revokeUserRole").Funcs(template.FuncMap{
		"inDollar": func(offset int, n int) string {
			dollar := "$" + strconv.Itoa(n+offset+1)
			if n > 0 {
				dollar = "," + dollar
			}
			return dollar
		},
	}).Parse(revokeUserRoleDeleteTemplate))

type revokeUserRole struct {
	UserRoleTable string
	Roles         []string
	Users         []string
}

func (s AuthInfoStore) revokeUserRoleSQL(users []string, roles []string) (string, []interface{}) {
	b := bytes.Buffer{}
	revokeUserRoleDelete.Execute(&b, revokeUserRole{
		s.sqlBuilder.FullTableName("user_role"),
		roles,
		users,
	})

	args := make([]interface{}, len(users)+len(roles))
	for i := range users {
		args[i] = users[i]
	}
	offset := len(users)
	for i := range roles {
		args[i+offset] = roles[i]
	}
	return b.String(), args
}
