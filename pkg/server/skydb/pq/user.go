// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pq

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

func (c *conn) CreateAuth(authinfo *skydb.AuthInfo) (err error) {
	var (
		tokenValidSince *time.Time
		lastLoginAt     *time.Time
		lastSeenAt      *time.Time
	)
	tokenValidSince = authinfo.TokenValidSince
	if tokenValidSince != nil && tokenValidSince.IsZero() {
		tokenValidSince = nil
	}
	lastLoginAt = authinfo.LastLoginAt
	if lastLoginAt != nil && lastLoginAt.IsZero() {
		lastLoginAt = nil
	}
	lastSeenAt = authinfo.LastSeenAt
	if lastSeenAt != nil && lastSeenAt.IsZero() {
		lastSeenAt = nil
	}

	builder := psql.Insert(c.tableName("_auth")).Columns(
		"id",
		"password",
		"provider_info",
		"token_valid_since",
		"last_login_at",
		"last_seen_at",
	).Values(
		authinfo.ID,
		authinfo.HashedPassword,
		providerInfoValue{authinfo.ProviderInfo, true},
		tokenValidSince,
		lastLoginAt,
		lastSeenAt,
	)

	_, err = c.ExecWith(builder)
	if isUniqueViolated(err) {
		return skydb.ErrUserDuplicated
	}

	if err := c.UpdateUserRoles(authinfo); err != nil {
		return skydb.ErrRoleUpdatesFailed
	}
	return err
}

func (c *conn) UpdateAuth(authinfo *skydb.AuthInfo) (err error) {
	var (
		tokenValidSince *time.Time
		lastLoginAt     *time.Time
		lastSeenAt      *time.Time
	)
	tokenValidSince = authinfo.TokenValidSince
	if tokenValidSince != nil && tokenValidSince.IsZero() {
		tokenValidSince = nil
	}
	lastLoginAt = authinfo.LastLoginAt
	if lastLoginAt != nil && lastLoginAt.IsZero() {
		lastLoginAt = nil
	}
	lastSeenAt = authinfo.LastSeenAt
	if lastSeenAt != nil && lastSeenAt.IsZero() {
		lastSeenAt = nil
	}

	builder := psql.Update(c.tableName("_auth")).
		Set("password", authinfo.HashedPassword).
		Set("provider_info", providerInfoValue{authinfo.ProviderInfo, true}).
		Set("token_valid_since", tokenValidSince).
		Set("last_login_at", lastLoginAt).
		Set("last_seen_at", lastSeenAt).
		Where("id = ?", authinfo.ID)

	result, err := c.ExecWith(builder)
	if err != nil {
		if isUniqueViolated(err) {
			return skydb.ErrUserDuplicated
		}
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	if err := c.UpdateUserRoles(authinfo); err != nil {
		return skydb.ErrRoleUpdatesFailed
	}
	return nil
}

func (c *conn) baseUserBuilder() sq.SelectBuilder {
	return psql.Select("id", "password", "provider_info",
		"token_valid_since", "last_login_at", "last_seen_at",
		"array_to_json(array_agg(role_id)) AS roles").
		From(c.tableName("_auth")).
		LeftJoin(c.tableName("_auth_role") + " ON id = auth_id").
		GroupBy("id")
}

func (c *conn) doScanAuth(authinfo *skydb.AuthInfo, scanner sq.RowScanner) error {
	var (
		id              string
		tokenValidSince pq.NullTime
		lastLoginAt     pq.NullTime
		lastSeenAt      pq.NullTime
		roles           nullJSONStringSlice
	)
	password, providerInfo := []byte{}, providerInfoValue{}

	err := scanner.Scan(
		&id,
		&password,
		&providerInfo,
		&tokenValidSince,
		&lastLoginAt,
		&lastSeenAt,
		&roles,
	)
	if err != nil {
		log.Infof(err.Error())
	}
	if err == sql.ErrNoRows {
		return skydb.ErrUserNotFound
	}

	authinfo.ID = id
	authinfo.HashedPassword = password
	authinfo.ProviderInfo = providerInfo.ProviderInfo
	if tokenValidSince.Valid {
		authinfo.TokenValidSince = &tokenValidSince.Time
	} else {
		authinfo.TokenValidSince = nil
	}
	if lastLoginAt.Valid {
		authinfo.LastLoginAt = &lastLoginAt.Time
	} else {
		authinfo.LastLoginAt = nil
	}
	if lastSeenAt.Valid {
		authinfo.LastSeenAt = &lastSeenAt.Time
	} else {
		authinfo.LastSeenAt = nil
	}
	authinfo.Roles = roles.slice

	return err
}

func (c *conn) GetAuth(id string, authinfo *skydb.AuthInfo) error {
	log.Warnf(id)
	builder := c.baseUserBuilder().Where("id = ?", id)
	scanner := c.QueryRowWith(builder)
	return c.doScanAuth(authinfo, scanner)
}

func (c *conn) GetAuthByPrincipalID(principalID string, authinfo *skydb.AuthInfo) error {
	builder := c.baseUserBuilder().Where("jsonb_exists(provider_info, ?)", principalID)
	scanner := c.QueryRowWith(builder)
	return c.doScanAuth(authinfo, scanner)
}

func (c *conn) DeleteAuth(id string) error {
	builder := psql.Delete(c.tableName("_auth")).
		Where("id = ?", id)

	result, err := c.ExecWith(builder)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows deleted, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) SetAuthRecordKeys(authRecordKeys [][]string) {
	c.authRecordKeys = authRecordKeys
}

func (c *conn) GetAuthRecordKeys() [][]string {
	return c.authRecordKeys
}

func (c *conn) EnsureAuthRecordKeysValid() error {
	authRecordKeys := c.authRecordKeys
	db := c.PublicDB().(*database)
	schema, err := db.GetSchema(db.UserRecordType())
	if err != nil {
		return fmt.Errorf("Unable to retrieve user record schema")
	}

	// check if auth record keys exist in user record
	// if not and canMigrate = true, just pass, since not enough information to migrate
	// if not and canMigrate = false, return error
	missingKeys := c.findMissingRecordKeys(schema, authRecordKeys)
	if len(missingKeys) > 0 && !c.canMigrate {
		return fmt.Errorf("%v are missing in user record schema", missingKeys)
	}

	// check if auth record keys that have invalid record type
	// return error if found
	for _, keys := range authRecordKeys {
		for _, key := range keys {
			err = c.ensureRecordKeyValidType(schema, key)
			if err != nil {
				return err
			}
		}
	}

	// check if auth record keys have unique constraints
	// if not and canMigrate = true, create a special unique constraint, which would be managed by skygear
	// if not and canMigrate = false, return error
	for _, keys := range authRecordKeys {
		indexes, err := db.getIndexes(db.UserRecordType())
		if err != nil {
			return err
		}

		err = c.ensureRecordKeysUnique(indexes, keys)
		if err != nil {
			if !c.canMigrate {
				return fmt.Errorf("%v must be unique in user record schema", keys)
			}

			if schema.HasFields(keys) {
				if err = c.createManagedUniqueConstraint(keys); err != nil {
					return err
				}
			}
		}
	}

	// if canMigrate = true, clean up unique constraint managed by skygear, which is no longer exist
	if c.canMigrate {
		if err = c.removeManagedUniqueConstraints(authRecordKeys); err != nil {
			return err
		}
	}

	return nil
}

func (c *conn) findMissingRecordKeys(schema skydb.RecordSchema, authRecordKeys [][]string) []string {
	recordKeyMap := map[string]bool{}
	for _, keys := range authRecordKeys {
		for _, key := range keys {
			recordKeyMap[key] = schema.HasField(key)
		}
	}

	missing := []string{}
	for key, ok := range recordKeyMap {
		if !ok {
			missing = append(missing, key)
		}
	}

	return missing
}

func (c *conn) ensureRecordKeyValidType(schema skydb.RecordSchema, key string) error {
	if !schema.HasField(key) {
		return nil
	}

	fieldType := schema[key].Type

	var valid bool
	switch fieldType {
	case skydb.TypeString, skydb.TypeNumber, skydb.TypeInteger, skydb.TypeBoolean, skydb.TypeSequence:
		valid = true
	default:
		valid = false
	}

	if !valid {
		return fmt.Errorf("%v must be string, number, integer, boolean or sequence", key)
	}

	return nil
}

func (c *conn) ensureRecordKeysUnique(indexes []dbIndex, authRecordKeys []string) error {
	for _, index := range indexes {
		columns := index.columns

		sort.Strings(authRecordKeys)
		sort.Strings(columns)

		if strings.Join(authRecordKeys, ",") == strings.Join(columns, ",") {
			return nil
		}
	}

	return fmt.Errorf("Unique index not found on keys: %v", authRecordKeys)
}

func (c *conn) indexName(tableName string, columns []string) string {
	return fmt.Sprintf("auth_record_keys_%s_%s_key", tableName, strings.Join(columns, "_"))
}

func (c *conn) createManagedUniqueConstraint(columns []string) error {
	db := c.PublicDB().(*database)
	tableName := db.UserRecordType()
	return db.createIndex(dbIndex{
		table:   tableName,
		name:    c.indexName(tableName, columns),
		columns: columns,
	})
}

func (c *conn) removeManagedUniqueConstraints(exceptColumns [][]string) error {
	db := c.PublicDB().(*database)
	indexes, err := db.getIndexes(db.UserRecordType())
	if err != nil {
		return err
	}

	exceptColumnNames := []string{}
	for _, cols := range exceptColumns {
		exceptColumnNames = append(exceptColumnNames, c.indexName(db.UserRecordType(), cols))
	}

	for _, index := range indexes {
		if index.name != c.indexName(index.table, index.columns) {
			continue
		}

		isException := false
		for _, exceptName := range exceptColumnNames {
			if index.name == exceptName {
				isException = true
				break
			}
		}

		if isException {
			continue
		}

		db.dropIndex(index.table, index.name)
	}

	return nil
}
