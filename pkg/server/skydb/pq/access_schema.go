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

func (c *conn) SetRecordFieldAccess(acl skydb.FieldACL) (err error) {
	tx, err := c.db.Beginx()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit(); err != nil {
			err = fmt.Errorf("unable to commit transaction: %s", err)
			return
		}

		c.FieldACL = nil // invalidate cached FieldACL
	}()

	deleteBuilder := psql.
		Delete(c.tableName("_record_field_access"))

	if _, err = c.ExecWith(deleteBuilder); err != nil {
		return
	}

	allEntries := acl.AllEntries()
	if len(allEntries) == 0 {
		// Do not insert if new setting is empty.
		return
	}

	builder := psql.
		Insert(c.tableName("_record_field_access")).
		Columns(
			"record_type",
			"record_field",
			"user_role",
			"writable",
			"readable",
			"comparable",
			"discoverable",
		)

	for _, entry := range allEntries {
		builder = builder.Values(
			entry.RecordType,
			entry.RecordField,
			entry.UserRole.String(),
			entry.Writable,
			entry.Readable,
			entry.Comparable,
			entry.Discoverable,
		)
	}

	_, err = c.ExecWith(builder)
	return
}

func (c *conn) GetRecordFieldAccess() (skydb.FieldACL, error) {
	if c.FieldACL != nil {
		return *c.FieldACL, nil
	}

	builder := psql.
		Select(
			"record_type",
			"record_field",
			"user_role",
			"writable",
			"readable",
			"comparable",
			"discoverable",
		).
		From(c.tableName("_record_field_access"))

	var recordTypeString string
	var recordFieldString string
	var userRoleString string
	var writableBoolean bool
	var readableBoolean bool
	var comparableBoolean bool
	var discoverableBoolean bool

	rows, err := c.QueryWith(builder)
	if err != nil {
		return skydb.FieldACL{}, err
	}

	entries := []skydb.FieldACLEntry{}
	var entry skydb.FieldACLEntry
	for rows.Next() {
		err := rows.Scan(
			&recordTypeString,
			&recordFieldString,
			&userRoleString,
			&writableBoolean,
			&readableBoolean,
			&comparableBoolean,
			&discoverableBoolean,
		)
		if err != nil {
			return skydb.FieldACL{}, err
		}

		entry.RecordType = recordTypeString
		entry.RecordField = recordFieldString
		entry.UserRole = skydb.NewFieldUserRole(userRoleString)
		entry.Writable = writableBoolean
		entry.Readable = readableBoolean
		entry.Comparable = comparableBoolean
		entry.Discoverable = discoverableBoolean
		entries = append(entries, entry)
	}

	acl := skydb.NewFieldACL(skydb.FieldACLEntryList(entries))

	c.FieldACL = &acl
	return acl, nil
}
