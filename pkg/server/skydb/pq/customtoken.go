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
	"time"

	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

func (c *conn) CreateCustomTokenInfo(tokenInfo *skydb.CustomTokenInfo) (err error) {
	createdAt := time.Now()
	if tokenInfo.CreatedAt != nil && !tokenInfo.CreatedAt.IsZero() {
		createdAt = *tokenInfo.CreatedAt
	}

	builder := psql.Insert(c.tableName("_sso_custom_token")).Columns(
		"user_id",
		"principal_id",
		"_created_at",
	).Values(
		tokenInfo.UserID,
		tokenInfo.PrincipalID,
		createdAt,
	)

	_, err = c.ExecWith(builder)
	if isUniqueViolated(err) {
		return skydb.ErrUserDuplicated
	}
	return err
}

func (c *conn) customTokenBuilder() sq.SelectBuilder {
	return psql.Select("user_id", "principal_id", "_created_at").
		From(c.tableName("_sso_custom_token"))
}

func (c *conn) doScanCustomTokenInfo(tokenInfo *skydb.CustomTokenInfo, scanner sq.RowScanner) error {
	var (
		userID      string
		principalID string
		createdAt   pq.NullTime
	)

	err := scanner.Scan(
		&userID,
		&principalID,
		&createdAt,
	)
	if err != nil {
		log.Infof(err.Error())
	}
	if err == sql.ErrNoRows {
		return skydb.ErrUserNotFound
	}

	tokenInfo.UserID = userID
	tokenInfo.PrincipalID = principalID

	if createdAt.Valid {
		tokenInfo.CreatedAt = &createdAt.Time
	} else {
		tokenInfo.CreatedAt = nil
	}

	return err
}

func (c *conn) GetCustomTokenInfo(principalID string, tokenInfo *skydb.CustomTokenInfo) error {
	builder := c.customTokenBuilder().
		Where("principal_id = ?", principalID)
	scanner := c.QueryRowWith(builder)
	return c.doScanCustomTokenInfo(tokenInfo, scanner)
}

func (c *conn) DeleteCustomTokenInfo(principalID string) error {
	builder := psql.Delete(c.tableName("_sso_custom_token")).
		Where("principal_id = ?", principalID)

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
