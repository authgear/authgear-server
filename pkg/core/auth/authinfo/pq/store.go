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

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"

	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/db"
	skySql "github.com/skygeario/skygear-server/pkg/core/db/sql"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type Store struct {
	db          *sqlx.DB
	sqlBuilder  skySql.Builder
	sqlExecutor skySql.Executor
	logger      *logrus.Entry
}

func (s *Store) CreateAuth(authinfo *authinfo.AuthInfo) (err error) {
	var (
		tokenValidSince *time.Time
		lastSeenAt      *time.Time
		disabledReason  *string
		disabledExpiry  *time.Time
	)
	tokenValidSince = authinfo.TokenValidSince
	if tokenValidSince != nil && tokenValidSince.IsZero() {
		tokenValidSince = nil
	}
	lastSeenAt = authinfo.LastSeenAt
	if lastSeenAt != nil && lastSeenAt.IsZero() {
		lastSeenAt = nil
	}
	disabledReason = &authinfo.DisabledMessage
	if *disabledReason == "" {
		disabledReason = nil
	}
	disabledExpiry = authinfo.DisabledExpiry
	if disabledExpiry != nil && disabledExpiry.IsZero() {
		disabledExpiry = nil
	}

	builder := s.sqlBuilder.Insert(s.sqlBuilder.TableName("user")).Columns(
		"id",
		"token_valid_since",
		"last_seen_at",
		"disabled",
		"disabled_message",
		"disabled_expiry",
	).Values(
		authinfo.ID,
		tokenValidSince,
		lastSeenAt,
		authinfo.Disabled,
		disabledReason,
		disabledExpiry,
	)

	_, err = s.sqlExecutor.ExecWith(builder)
	if db.IsUniqueViolated(err) {
		return skydb.ErrUserDuplicated
	}

	// TODO:
	// if err := s.UpdateUserRoles(authinfo); err != nil {
	// 	return skydb.ErrRoleUpdatesFailed
	// }

	return err
}

// nolint: gocyclo
func (s *Store) UpdateAuth(authinfo *authinfo.AuthInfo) (err error) {
	var (
		tokenValidSince *time.Time
		lastSeenAt      *time.Time
		disabledReason  *string
		disabledExpiry  *time.Time
	)
	tokenValidSince = authinfo.TokenValidSince
	if tokenValidSince != nil && tokenValidSince.IsZero() {
		tokenValidSince = nil
	}
	lastSeenAt = authinfo.LastSeenAt
	if lastSeenAt != nil && lastSeenAt.IsZero() {
		lastSeenAt = nil
	}
	disabledReason = &authinfo.DisabledMessage
	if *disabledReason == "" {
		disabledReason = nil
	}
	disabledExpiry = authinfo.DisabledExpiry
	if disabledExpiry != nil && disabledExpiry.IsZero() {
		disabledExpiry = nil
	}

	builder := s.sqlBuilder.Update(s.sqlBuilder.TableName("user")).
		Set("token_valid_since", tokenValidSince).
		Set("last_seen_at", lastSeenAt).
		Set("disabled", authinfo.Disabled).
		Set("disabled_message", disabledReason).
		Set("disabled_expiry", disabledExpiry).
		Where("id = ?", authinfo.ID)

	result, err := s.sqlExecutor.ExecWith(builder)
	if err != nil {
		if db.IsUniqueViolated(err) {
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

	// TODO:
	// if err := c.UpdateUserRoles(authinfo); err != nil {
	// 	return skydb.ErrRoleUpdatesFailed
	// }

	return nil
}

func (s *Store) baseUserBuilder() sq.SelectBuilder {
	// TODO:
	return s.sqlBuilder.Select("id",
		"token_valid_since", "last_seen_at",
		"disabled", "disabled_message", "disabled_expiry",
		"array_to_json(array_agg(role_id)) AS roles").
		From(s.sqlBuilder.TableName("user")).
		LeftJoin(s.sqlBuilder.TableName("_auth_role") + " ON id = auth_id").
		GroupBy("id")
}

func (s *Store) doScanAuth(authinfo *authinfo.AuthInfo, scanner sq.RowScanner) error {
	logger := s.logger
	var (
		id              string
		tokenValidSince pq.NullTime
		lastSeenAt      pq.NullTime
		roles           db.NullJSONStringSlice
		disabled        bool
		disabledReason  sql.NullString
		disabledExpiry  pq.NullTime
	)

	err := scanner.Scan(
		&id,
		&tokenValidSince,
		&lastSeenAt,
		&disabled,
		&disabledReason,
		&disabledExpiry,
		&roles,
	)
	if err != nil {
		logger.Infof(err.Error())
	}
	if err == sql.ErrNoRows {
		return skydb.ErrUserNotFound
	}

	authinfo.ID = id
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
	authinfo.Disabled = disabled
	if disabled {
		if disabledReason.Valid {
			authinfo.DisabledMessage = disabledReason.String
		} else {
			authinfo.DisabledMessage = ""
		}
		if disabledExpiry.Valid {
			expiry := disabledExpiry.Time.UTC()
			authinfo.DisabledExpiry = &expiry
		} else {
			authinfo.DisabledExpiry = nil
		}
	} else {
		authinfo.DisabledMessage = ""
		authinfo.DisabledExpiry = nil
	}

	authinfo.Roles = roles.Slice

	return err
}

func (s *Store) GetAuth(id string, authinfo *authinfo.AuthInfo) error {
	builder := s.baseUserBuilder().Where("id = ?", id)
	scanner := s.sqlExecutor.QueryRowWith(builder)
	return s.doScanAuth(authinfo, scanner)
}

func (s *Store) DeleteAuth(id string) error {
	builder := s.sqlBuilder.Delete(s.sqlBuilder.TableName("user")).
		Where("id = ?", id)

	result, err := s.sqlExecutor.ExecWith(builder)
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
