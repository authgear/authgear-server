package pq

import (
	"database/sql"
	"encoding/json"
	"fmt"

	sq "github.com/lann/squirrel"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/pq/builder"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/utils"
)

func (c *conn) SetRecordAccess(recordType string, acl skydb.RecordACL) error {
	creationRoles := []string{}
	for _, ace := range acl {
		if ace.Role != "" {
			creationRoles = append(creationRoles, ace.Role)
		}
	}

	_, err := c.ensureRole(creationRoles)
	if err != nil {
		return err
	}

	currentCreationAccess, err := c.GetRecordAccess(recordType)
	if err != nil {
		return err
	}

	currentCreationRoles := []string{}
	for _, perACE := range currentCreationAccess {
		if perACE.Role != "" {
			currentCreationRoles = append(currentCreationRoles, perACE.Role)
		}
	}

	rolesToDelete := utils.StringSliceExcept(currentCreationRoles, creationRoles)
	rolesToAdd := utils.StringSliceExcept(creationRoles, currentCreationRoles)

	err = c.deleteRecordCreationAccess(recordType, rolesToDelete)
	if err != nil {
		return err
	}

	err = c.insertRecordCreationAccess(recordType, rolesToAdd)

	return err
}

func (c *conn) GetRecordAccess(recordType string) (skydb.RecordACL, error) {
	builder := psql.
		Select("role_id").
		From(c.tableName("_record_creation")).
		Where(sq.Eq{"record_type": recordType}).
		Join(fmt.Sprintf("%s ON %s.role_id = id",
		c.tableName("_role"),
		c.tableName("_record_creation")))

	rows, err := c.QueryWith(builder)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	currentCreationRoles := []skydb.RecordACLEntry{}
	for rows.Next() {
		roleStr := ""
		if err := rows.Scan(&roleStr); err != nil {
			return nil, err
		}
		currentCreationRoles = append(currentCreationRoles,
			skydb.NewRecordACLEntryRole(roleStr, skydb.CreateLevel))
	}

	return skydb.NewRecordACL(currentCreationRoles), nil
}

func (c *conn) deleteRecordCreationAccess(recordType string, roles []string) error {
	if len(roles) == 0 {
		return nil
	}
	roleArgs := make([]interface{}, len(roles))
	for idx, perRole := range roles {
		roleArgs[idx] = interface{}(perRole)
	}

	builder := psql.
		Delete(c.tableName("_record_creation")).
		Where("role_id IN ("+sq.Placeholders(len(roles))+")", roleArgs...)

	_, err := c.ExecWith(builder)
	return err
}

func (c *conn) insertRecordCreationAccess(recordType string, roles []string) error {
	if len(roles) == 0 {
		return nil
	}

	for _, perRole := range roles {
		builder := psql.
			Insert(c.tableName("_record_creation")).
			Columns("record_type", "role_id").
			Values(recordType, perRole)

		_, err := c.ExecWith(builder)
		if isForeignKeyViolated(err) {
			return skyerr.NewError(skyerr.ConstraintViolated,
				fmt.Sprintf("Does not have role %s", perRole))
		} else if isUniqueViolated(err) {
			return skyerr.NewError(skyerr.Duplicated,
				fmt.Sprintf("Role %s is already have creation access for Record %s",
					perRole, recordType))
		}
	}

	return nil
}

func (c *conn) SetRecordDefaultAccess(recordType string, acl skydb.RecordACL) error {
	pkData := map[string]interface{}{
		"record_type": recordType,
	}
	values := map[string]interface{}{
		"default_access": aclValue(acl),
	}

	upsert := builder.UpsertQuery(c.tableName("_record_default_access"), pkData, values)
	_, err := c.ExecWith(upsert)

	if err != nil {
		return err
	}

	return nil
}

func (c *conn) GetRecordDefaultAccess(recordType string) (skydb.RecordACL, error) {
	builder := psql.
		Select("default_access").
		From(c.tableName("_record_default_access")).
		Where(sq.Eq{"record_type": recordType})

	nullableACLString := sql.NullString{}
	err := c.QueryRowWith(builder).Scan(&nullableACLString)
	if err != nil {
		return nil, err
	}

	acl := skydb.RecordACL{}
	if nullableACLString.Valid {
		json.Unmarshal([]byte(nullableACLString.String), &acl)
		return acl, nil
	}
	return nil, nil
}
