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
		lastSeenAt      *time.Time
	)
	tokenValidSince = authinfo.TokenValidSince
	if tokenValidSince != nil && tokenValidSince.IsZero() {
		tokenValidSince = nil
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
		"last_seen_at",
	).Values(
		authinfo.ID,
		authinfo.HashedPassword,
		providerInfoValue{authinfo.ProviderInfo, true},
		tokenValidSince,
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
		lastSeenAt      *time.Time
	)
	tokenValidSince = authinfo.TokenValidSince
	if tokenValidSince != nil && tokenValidSince.IsZero() {
		tokenValidSince = nil
	}
	lastSeenAt = authinfo.LastSeenAt
	if lastSeenAt != nil && lastSeenAt.IsZero() {
		lastSeenAt = nil
	}

	builder := psql.Update(c.tableName("_auth")).
		Set("password", authinfo.HashedPassword).
		Set("provider_info", providerInfoValue{authinfo.ProviderInfo, true}).
		Set("token_valid_since", tokenValidSince).
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
		"token_valid_since", "last_seen_at",
		"array_to_json(array_agg(role_id)) AS roles").
		From(c.tableName("_auth")).
		LeftJoin(c.tableName("_auth_role") + " ON id = auth_id").
		GroupBy("id")
}

func (c *conn) doScanAuth(authinfo *skydb.AuthInfo, scanner sq.RowScanner) error {
	var (
		id              string
		tokenValidSince pq.NullTime
		lastSeenAt      pq.NullTime
		roles           nullJSONStringSlice
	)
	password, providerInfo := []byte{}, providerInfoValue{}

	err := scanner.Scan(
		&id,
		&password,
		&providerInfo,
		&tokenValidSince,
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

func (c *conn) EnsureAuthRecordKeysExist(authRecordKeys [][]string) error {
	db := c.PublicDB().(*database)
	userRecordType := db.UserRecordType()
	schema, err := db.GetSchema(userRecordType)
	if err != nil {
		return fmt.Errorf("Unable to retrieve user record schema")
	}

	requiredKeys := getAllAuthRecordKeys(authRecordKeys)
	schemaToExtend := skydb.RecordSchema{}
	for _, key := range requiredKeys {
		if _, ok := schema[key]; ok {
			continue
		}

		schemaToExtend[key] = skydb.FieldType{
			Type: skydb.TypeString,
		}
	}

	if _, err := db.Extend(userRecordType, schemaToExtend); err != nil {
		return err
	}

	return nil
}

func (c *conn) EnsureAuthRecordKeysIndexesMatch(authRecordKeys [][]string) error {
	db := c.PublicDB().(*database)
	userRecordType := db.UserRecordType()

	requiredIndexes := []skydb.Index{}
	for _, keys := range authRecordKeys {
		requiredIndexes = append(requiredIndexes, skydb.Index{
			Fields: keys,
		})
	}

	indexesByName, err := db.GetIndexesByRecordType(userRecordType)
	if err != nil {
		return err
	}

	indexes := []skydb.Index{}
	for _, index := range indexesByName {
		indexes = append(indexes, index)
	}

	requiredIndexesByFields := groupIndexesByFields(requiredIndexes)
	indexesByFields := groupIndexesByFields(indexes)

	for fieldsString := range requiredIndexesByFields {
		if _, ok := indexesByFields[fieldsString]; !ok {
			index := requiredIndexesByFields[fieldsString]
			if !c.canMigrate {
				return fmt.Errorf("Index of %v is required in user record schema", index.Fields)
			}

			if err = db.SaveIndex(userRecordType, managedIndexName(userRecordType, index), index); err != nil {
				return err
			}
		}
	}

	// cleanup unused unique constraint
	if c.canMigrate {
		for indexName, index := range indexesByName {
			fieldsString := joinFields(index.Fields)
			_, isRequired := requiredIndexesByFields[fieldsString]
			if !isRequired && indexName == managedIndexName(userRecordType, index) {
				db.DeleteIndex(userRecordType, indexName)
			}
		}
	}

	return nil
}

func getAllAuthRecordKeys(authRecordKeys [][]string) []string {
	recordKeyMap := map[string]bool{}
	for _, keys := range authRecordKeys {
		for _, key := range keys {
			recordKeyMap[key] = true
		}
	}

	allKeys := []string{}
	for key := range recordKeyMap {
		allKeys = append(allKeys, key)
	}

	return allKeys
}

func joinFields(fields []string) string {
	sort.Strings(fields)
	return strings.Join(fields, ",")
}

func groupIndexesByFields(indexes []skydb.Index) map[string]skydb.Index {
	output := map[string]skydb.Index{}
	for _, index := range indexes {
		fieldsString := joinFields(index.Fields)
		output[fieldsString] = index
	}
	return output
}

func managedIndexName(recordType string, index skydb.Index) string {
	fields := index.Fields
	sort.Strings(fields)
	return fmt.Sprintf("auth_record_keys_%s_%s_key", recordType, strings.Join(fields, "_"))
}
