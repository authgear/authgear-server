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
	"strings"
	"time"

	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

func (c *conn) CreateUser(authinfo *skydb.AuthInfo) (err error) {
	var (
		username        *string
		email           *string
		tokenValidSince *time.Time
		lastLoginAt     *time.Time
		lastSeenAt      *time.Time
	)
	if authinfo.Username != "" {
		username = &authinfo.Username
	} else {
		username = nil
	}
	if authinfo.Email != "" {
		email = &authinfo.Email
	} else {
		email = nil
	}
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
		"username",
		"email",
		"password",
		"auth",
		"token_valid_since",
		"last_login_at",
		"last_seen_at",
	).Values(
		authinfo.ID,
		username,
		email,
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

func (c *conn) UpdateUser(authinfo *skydb.AuthInfo) (err error) {
	var (
		username        *string
		email           *string
		tokenValidSince *time.Time
		lastLoginAt     *time.Time
		lastSeenAt      *time.Time
	)
	if authinfo.Username != "" {
		username = &authinfo.Username
	} else {
		username = nil
	}
	if authinfo.Email != "" {
		email = &authinfo.Email
	} else {
		email = nil
	}
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
		Set("username", username).
		Set("email", email).
		Set("password", authinfo.HashedPassword).
		Set("auth", providerInfoValue{authinfo.ProviderInfo, true}).
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
	return psql.Select("id", "username", "email", "password", "auth",
		"token_valid_since", "last_login_at", "last_seen_at",
		"array_to_json(array_agg(role_id)) AS roles").
		From(c.tableName("_auth")).
		LeftJoin(c.tableName("_auth_role") + " ON id = auth_id").
		GroupBy("id")
}

func (c *conn) doScanUser(authinfo *skydb.AuthInfo, scanner sq.RowScanner) error {
	var (
		id              string
		username        sql.NullString
		email           sql.NullString
		tokenValidSince pq.NullTime
		lastLoginAt     pq.NullTime
		lastSeenAt      pq.NullTime
		roles           nullJSONStringSlice
	)
	password, auth := []byte{}, providerInfoValue{}

	err := scanner.Scan(
		&id,
		&username,
		&email,
		&password,
		&auth,
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
	authinfo.Username = username.String
	authinfo.Email = email.String
	authinfo.HashedPassword = password
	authinfo.ProviderInfo = auth.ProviderInfo
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

func (c *conn) GetUser(id string, authinfo *skydb.AuthInfo) error {
	log.Warnf(id)
	builder := c.baseUserBuilder().Where("id = ?", id)
	scanner := c.QueryRowWith(builder)
	return c.doScanUser(authinfo, scanner)
}

func (c *conn) GetUserByUsernameEmail(username string, email string, authinfo *skydb.AuthInfo) error {
	var builder sq.SelectBuilder
	if email == "" {
		builder = c.baseUserBuilder().Where("username = ?", username)
	} else if username == "" {
		builder = c.baseUserBuilder().Where("email = ?", email)
	} else {
		builder = c.baseUserBuilder().Where("username = ? AND email = ?", username, email)
	}
	scanner := c.QueryRowWith(builder)
	return c.doScanUser(authinfo, scanner)
}

func (c *conn) GetUserByPrincipalID(principalID string, authinfo *skydb.AuthInfo) error {
	builder := c.baseUserBuilder().Where("jsonb_exists(auth, ?)", principalID)
	scanner := c.QueryRowWith(builder)
	return c.doScanUser(authinfo, scanner)
}

func (c *conn) QueryUser(emails []string, usernames []string) ([]skydb.AuthInfo, error) {
	emailargs := make([]interface{}, len(emails))
	for i, v := range emails {
		if v == "" {
			continue
		}
		emailargs[i] = interface{}(v)
	}

	usernameargs := make([]interface{}, len(usernames))
	for i, v := range usernames {
		if v == "" {
			continue
		}
		usernameargs[i] = interface{}(v)
	}

	if len(emailargs) == 0 && len(usernameargs) == 0 {
		return []skydb.AuthInfo{}, nil
	}

	var sqls []string
	var args []interface{}
	if len(emailargs) > 0 {
		sqls = append(sqls, fmt.Sprintf("email IN (%s) AND email IS NOT NULL AND email != ''", sq.Placeholders(len(emailargs))))
		args = append(args, emailargs...)
	}
	if len(usernameargs) > 0 {
		sqls = append(sqls, fmt.Sprintf("username IN (%s) AND username IS NOT NULL AND username != ''", sq.Placeholders(len(usernameargs))))
		args = append(args, usernameargs...)
	}

	builder := c.baseUserBuilder().
		Where(strings.Join(sqls, " OR "), args...)

	rows, err := c.QueryWith(builder)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	results := []skydb.AuthInfo{}
	for rows.Next() {
		authinfo := skydb.AuthInfo{}
		if err := c.doScanUser(&authinfo, rows); err != nil {
			panic(err)
		}
		results = append(results, authinfo)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return results, nil
}

func (c *conn) DeleteUser(id string) error {
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
