package pq

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func (s AuthInfoStore) updateUserRoles(authinfo *authinfo.AuthInfo) error {
	s.logger.Debugf("UpdateRoles %v", authinfo)
	builder := s.sqlBuilder.Delete(s.sqlBuilder.TableName("user_role")).Where("user_id = ?", authinfo.ID)
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
		s.sqlBuilder.TableName("user_role"),
		s.sqlBuilder.TableName("role"),
		roles,
	})

	args := make([]interface{}, len(roles)+1)
	args[0] = id
	for i := range roles {
		args[i+1] = roles[i]
	}
	return b.String(), args
}
