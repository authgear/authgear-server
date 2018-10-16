package pq

import (
	"database/sql"
	"encoding/json"
	"fmt"

	sq "github.com/lann/squirrel"

	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	"github.com/skygeario/skygear-server/pkg/server/skydb/pq/builder"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/utils"
)

func (s *RecordStore) SetRecordAccess(recordType string, acl record.ACL) error {
	creationRoles := []string{}
	for _, ace := range acl {
		if ace.Role != "" {
			creationRoles = append(creationRoles, ace.Role)
		}
	}

	_, err := role.EnsureRole(s.roleStore, s.logger, creationRoles)
	if err != nil {
		return err
	}

	currentCreationAccess, err := s.GetRecordAccess(recordType)
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

	err = s.deleteRecordCreationAccess(recordType, rolesToDelete)
	if err != nil {
		return err
	}

	err = s.insertRecordCreationAccess(recordType, rolesToAdd)

	return err
}

func (s *RecordStore) GetRecordAccess(recordType string) (record.ACL, error) {
	builder := s.sqlBuilder.
		Select("role_id").
		From(s.sqlBuilder.TableName("creation")).
		Where(sq.Eq{"record_type": recordType})

	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	currentCreationRoles := []record.ACLEntry{}
	for rows.Next() {
		roleStr := ""
		if err := rows.Scan(&roleStr); err != nil {
			return nil, err
		}
		currentCreationRoles = append(currentCreationRoles,
			record.NewACLEntryRole(roleStr, record.CreateLevel))
	}

	return record.NewACL(currentCreationRoles), nil
}

func (s *RecordStore) deleteRecordCreationAccess(recordType string, roles []string) error {
	if len(roles) == 0 {
		return nil
	}
	roleArgs := make([]interface{}, len(roles))
	for idx, perRole := range roles {
		roleArgs[idx] = interface{}(perRole)
	}

	builder := s.sqlBuilder.
		Delete(s.sqlBuilder.TableName("creation")).
		Where("role_id IN ("+sq.Placeholders(len(roles))+")", roleArgs...)

	_, err := s.sqlExecutor.ExecWith(builder)
	return err
}

func (s *RecordStore) insertRecordCreationAccess(recordType string, roles []string) error {
	if len(roles) == 0 {
		return nil
	}

	for _, perRole := range roles {
		builder := s.sqlBuilder.
			Insert(s.sqlBuilder.TableName("creation")).
			Columns("record_type", "role_id").
			Values(recordType, perRole)

		_, err := s.sqlExecutor.ExecWith(builder)
		if db.IsForeignKeyViolated(err) {
			return skyerr.NewError(skyerr.ConstraintViolated,
				fmt.Sprintf("Does not have role %s", perRole))
		} else if db.IsUniqueViolated(err) {
			return skyerr.NewError(skyerr.Duplicated,
				fmt.Sprintf("Role %s is already have creation access for Record %s",
					perRole, recordType))
		}
	}

	return nil
}

func (s *RecordStore) SetRecordDefaultAccess(recordType string, acl record.ACL) error {
	pkData := map[string]interface{}{
		"record_type": recordType,
	}
	values := map[string]interface{}{
		"default_access": aclValue(acl),
	}

	upsert := builder.UpsertQuery(s.sqlBuilder.TableName("default_access"), pkData, values)
	_, err := s.sqlExecutor.ExecWith(upsert)

	if err != nil {
		return err
	}

	return nil
}

func (s *RecordStore) GetRecordDefaultAccess(recordType string) (record.ACL, error) {
	builder := s.sqlBuilder.
		Select("default_access").
		From(s.sqlBuilder.TableName("default_access")).
		Where(sq.Eq{"record_type": recordType})

	nullableACLString := sql.NullString{}
	err := s.sqlExecutor.QueryRowWith(builder).Scan(&nullableACLString)
	if err != nil {
		return nil, err
	}

	acl := record.ACL{}
	if nullableACLString.Valid {
		json.Unmarshal([]byte(nullableACLString.String), &acl)
		return acl, nil
	}
	return nil, nil
}

func (s *RecordStore) SetRecordFieldAccess(acl record.FieldACL) (err error) {
	// defer func() {
	// 	c.FieldACL = nil // invalidate cached FieldACL
	// }()

	deleteBuilder := s.sqlBuilder.
		Delete(s.sqlBuilder.TableName("field_access"))

	if _, err = s.sqlExecutor.ExecWith(deleteBuilder); err != nil {
		return
	}

	allEntries := acl.AllEntries()
	if len(allEntries) == 0 {
		// Do not insert if new setting is empty.
		return
	}

	builder := s.sqlBuilder.
		Insert(s.sqlBuilder.TableName("field_access")).
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

	_, err = s.sqlExecutor.ExecWith(builder)
	return
}

func (s *RecordStore) GetRecordFieldAccess() (record.FieldACL, error) {
	// if c.FieldACL != nil {
	// 	return *c.FieldACL, nil
	// }

	builder := s.sqlBuilder.
		Select(
			"record_type",
			"record_field",
			"user_role",
			"writable",
			"readable",
			"comparable",
			"discoverable",
		).
		From(s.sqlBuilder.TableName("field_access"))

	var recordTypeString string
	var recordFieldString string
	var userRoleString string
	var writableBoolean bool
	var readableBoolean bool
	var comparableBoolean bool
	var discoverableBoolean bool

	rows, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		return record.FieldACL{}, err
	}

	entries := []record.FieldACLEntry{}
	var entry record.FieldACLEntry
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
			return record.FieldACL{}, err
		}

		entry.RecordType = recordTypeString
		entry.RecordField = recordFieldString
		entry.UserRole = record.NewFieldUserRole(userRoleString)
		entry.Writable = writableBoolean
		entry.Readable = readableBoolean
		entry.Comparable = comparableBoolean
		entry.Discoverable = discoverableBoolean
		entries = append(entries, entry)
	}

	acl := record.NewFieldACL(record.FieldACLEntryList(entries))

	// c.FieldACL = &acl
	return acl, nil
}
