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
	"github.com/skygeario/skygear-server/skydb"
)

func (c *conn) CreateUser(userinfo *skydb.UserInfo) (err error) {
	var (
		username        *string
		email           *string
		tokenValidSince *time.Time
	)
	if userinfo.Username != "" {
		username = &userinfo.Username
	} else {
		username = nil
	}
	if userinfo.Email != "" {
		email = &userinfo.Email
	} else {
		email = nil
	}
	tokenValidSince = userinfo.TokenValidSince
	if tokenValidSince != nil && tokenValidSince.IsZero() {
		tokenValidSince = nil
	}

	builder := psql.Insert(c.tableName("_user")).Columns(
		"id",
		"username",
		"email",
		"password",
		"auth",
		"token_valid_since",
	).Values(
		userinfo.ID,
		username,
		email,
		userinfo.HashedPassword,
		authInfoValue{userinfo.Auth, true},
		tokenValidSince,
	)

	_, err = c.ExecWith(builder)
	if isUniqueViolated(err) {
		return skydb.ErrUserDuplicated
	}

	if err := c.UpdateUserRoles(userinfo); err != nil {
		return skydb.ErrRoleUpdatesFailed
	}
	return err
}

func (c *conn) UpdateUser(userinfo *skydb.UserInfo) (err error) {
	var (
		username        *string
		email           *string
		tokenValidSince *time.Time
	)
	if userinfo.Username != "" {
		username = &userinfo.Username
	} else {
		username = nil
	}
	if userinfo.Email != "" {
		email = &userinfo.Email
	} else {
		email = nil
	}
	tokenValidSince = userinfo.TokenValidSince
	if tokenValidSince != nil && tokenValidSince.IsZero() {
		tokenValidSince = nil
	}

	builder := psql.Update(c.tableName("_user")).
		Set("username", username).
		Set("email", email).
		Set("password", userinfo.HashedPassword).
		Set("auth", authInfoValue{userinfo.Auth, true}).
		Set("token_valid_since", tokenValidSince).
		Where("id = ?", userinfo.ID)

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
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	if err := c.UpdateUserRoles(userinfo); err != nil {
		return skydb.ErrRoleUpdatesFailed
	}
	return nil
}

func (c *conn) baseUserBuilder() sq.SelectBuilder {
	return psql.Select("id", "username", "email", "password", "auth", "token_valid_since",
		"array_to_json(array_agg(role_id)) AS roles").
		From(c.tableName("_user")).
		LeftJoin(c.tableName("_user_role") + " ON id = user_id").
		GroupBy("id")
}

func (c *conn) doScanUser(userinfo *skydb.UserInfo, scanner sq.RowScanner) error {
	var (
		id              string
		username        sql.NullString
		email           sql.NullString
		tokenValidSince pq.NullTime
		roles           nullJSONStringSlice
	)
	password, auth := []byte{}, authInfoValue{}

	err := scanner.Scan(
		&id,
		&username,
		&email,
		&password,
		&auth,
		&tokenValidSince,
		&roles,
	)
	if err != nil {
		log.Infof(err.Error())
	}
	if err == sql.ErrNoRows {
		return skydb.ErrUserNotFound
	}

	userinfo.ID = id
	userinfo.Username = username.String
	userinfo.Email = email.String
	userinfo.HashedPassword = password
	userinfo.Auth = auth.AuthInfo
	if tokenValidSince.Valid {
		userinfo.TokenValidSince = &tokenValidSince.Time
	} else {
		userinfo.TokenValidSince = nil
	}
	userinfo.Roles = roles.slice

	return err
}

func (c *conn) GetUser(id string, userinfo *skydb.UserInfo) error {
	log.Warnf(id)
	builder := c.baseUserBuilder().Where("id = ?", id)
	scanner := c.QueryRowWith(builder)
	return c.doScanUser(userinfo, scanner)
}

func (c *conn) GetUserByUsernameEmail(username string, email string, userinfo *skydb.UserInfo) error {
	var builder sq.SelectBuilder
	if email == "" {
		builder = c.baseUserBuilder().Where("username = ?", username)
	} else if username == "" {
		builder = c.baseUserBuilder().Where("email = ?", email)
	} else {
		builder = c.baseUserBuilder().Where("username = ? AND email = ?", username, email)
	}
	scanner := c.QueryRowWith(builder)
	return c.doScanUser(userinfo, scanner)
}

func (c *conn) GetUserByPrincipalID(principalID string, userinfo *skydb.UserInfo) error {
	builder := c.baseUserBuilder().Where("jsonb_exists(auth, ?)", principalID)
	scanner := c.QueryRowWith(builder)
	return c.doScanUser(userinfo, scanner)
}

func (c *conn) QueryUser(emails []string, usernames []string) ([]skydb.UserInfo, error) {
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
		return []skydb.UserInfo{}, nil
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
	results := []skydb.UserInfo{}
	for rows.Next() {
		userinfo := skydb.UserInfo{}
		if err := c.doScanUser(&userinfo, rows); err != nil {
			panic(err)
		}
		results = append(results, userinfo)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return results, nil
}

func (c *conn) DeleteUser(id string) error {
	builder := psql.Delete(c.tableName("_user")).
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
