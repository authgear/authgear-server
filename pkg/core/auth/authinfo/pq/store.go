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
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	dbPq "github.com/skygeario/skygear-server/pkg/core/db/pq"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type authInfoStore struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
}

func newAuthInfoStore(builder db.SQLBuilder, executor db.SQLExecutor) *authInfoStore {
	return &authInfoStore{
		sqlBuilder:  builder,
		sqlExecutor: executor,
	}
}

func NewAuthInfoStore(builder db.SQLBuilder, executor db.SQLExecutor) authinfo.Store {
	return newAuthInfoStore(builder, executor)
}

func (s authInfoStore) CreateAuth(authinfo *authinfo.AuthInfo) error {
	var (
		lastSeenAt     *time.Time
		lastLoginAt    *time.Time
		disabledReason *string
		disabledExpiry *time.Time
		verifyInfo     dbPq.JSONMapBooleanValue
	)
	lastSeenAt = authinfo.LastSeenAt
	if lastSeenAt != nil && lastSeenAt.IsZero() {
		lastSeenAt = nil
	}
	lastLoginAt = authinfo.LastLoginAt
	if lastLoginAt != nil && lastLoginAt.IsZero() {
		lastLoginAt = nil
	}
	disabledReason = &authinfo.DisabledMessage
	if *disabledReason == "" {
		disabledReason = nil
	}
	disabledExpiry = authinfo.DisabledExpiry
	if disabledExpiry != nil && disabledExpiry.IsZero() {
		disabledExpiry = nil
	}

	verifyInfo = authinfo.VerifyInfo

	builder := s.sqlBuilder.Tenant().
		Insert(s.sqlBuilder.FullTableName("user")).
		Columns(
			"id",
			"last_seen_at",
			"last_login_at",
			"disabled",
			"disabled_message",
			"disabled_expiry",
			"manually_verified",
			"verified",
			"verify_info",
		).
		Values(
			authinfo.ID,
			lastSeenAt,
			lastLoginAt,
			authinfo.Disabled,
			disabledReason,
			disabledExpiry,
			authinfo.ManuallyVerified,
			authinfo.Verified,
			verifyInfo,
		)

	_, err := s.sqlExecutor.ExecWith(builder)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to create user")
	}

	return err
}

// UpdateAuth updates an existing AuthInfo matched by the ID field.
// nolint: gocyclo
func (s authInfoStore) UpdateAuth(info *authinfo.AuthInfo) error {
	var (
		lastSeenAt     *time.Time
		lastLoginAt    *time.Time
		disabledReason *string
		disabledExpiry *time.Time
		verifyInfo     dbPq.JSONMapBooleanValue
	)
	lastSeenAt = info.LastSeenAt
	if lastSeenAt != nil && lastSeenAt.IsZero() {
		lastSeenAt = nil
	}
	lastLoginAt = info.LastLoginAt
	if lastLoginAt != nil && lastLoginAt.IsZero() {
		lastLoginAt = nil
	}
	disabledReason = &info.DisabledMessage
	if *disabledReason == "" {
		disabledReason = nil
	}
	disabledExpiry = info.DisabledExpiry
	if disabledExpiry != nil && disabledExpiry.IsZero() {
		disabledExpiry = nil
	}

	verifyInfo = info.VerifyInfo

	builder := s.sqlBuilder.Tenant().
		Update(s.sqlBuilder.FullTableName("user")).
		Set("last_seen_at", lastSeenAt).
		Set("last_login_at", lastLoginAt).
		Set("disabled", info.Disabled).
		Set("disabled_message", disabledReason).
		Set("disabled_expiry", disabledExpiry).
		Set("manually_verified", info.ManuallyVerified).
		Set("verified", info.Verified).
		Set("verify_info", verifyInfo).
		Where("id = ?", info.ID)

	result, err := s.sqlExecutor.ExecWith(builder)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to update user")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.HandledWithMessage(err, "failed to update user")
	}
	if rowsAffected == 0 {
		return authinfo.ErrNotFound
	}

	return nil
}

func (s authInfoStore) baseUserBuilder() db.SelectBuilder {
	return s.sqlBuilder.Tenant().
		Select(
			"id",
			"last_seen_at",
			"last_login_at",
			"disabled",
			"disabled_message",
			"disabled_expiry",
			"manually_verified",
			"verified",
			"verify_info",
		).
		From(s.sqlBuilder.FullTableName("user"))
}

func (s authInfoStore) doScanAuth(authinfo *authinfo.AuthInfo, scanner sq.RowScanner) error {
	var (
		id               string
		lastSeenAt       pq.NullTime
		lastLoginAt      pq.NullTime
		disabled         bool
		disabledReason   sql.NullString
		disabledExpiry   pq.NullTime
		manuallyVerified bool
		verified         bool
		verifyInfo       dbPq.NullJSONMapBoolean
	)

	err := scanner.Scan(
		&id,
		&lastSeenAt,
		&lastLoginAt,
		&disabled,
		&disabledReason,
		&disabledExpiry,
		&manuallyVerified,
		&verified,
		&verifyInfo,
	)
	if err != nil {
		return err
	}

	authinfo.ID = id
	if lastSeenAt.Valid {
		authinfo.LastSeenAt = &lastSeenAt.Time
	} else {
		authinfo.LastSeenAt = nil
	}
	if lastLoginAt.Valid {
		authinfo.LastLoginAt = &lastLoginAt.Time
	} else {
		authinfo.LastLoginAt = nil
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

	authinfo.ManuallyVerified = manuallyVerified
	authinfo.Verified = verified
	authinfo.VerifyInfo = verifyInfo.JSON

	return nil
}

func (s authInfoStore) GetAuth(id string, info *authinfo.AuthInfo) error {
	builder := s.baseUserBuilder().Where("id = ?", id)
	scanner, err := s.sqlExecutor.QueryRowWith(builder)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to get user")
	}

	err = s.doScanAuth(info, scanner)
	if err == sql.ErrNoRows {
		return authinfo.ErrNotFound
	} else if err != nil {
		return errors.HandledWithMessage(err, "failed to get user")
	}
	return nil
}

func (s authInfoStore) DeleteAuth(id string) error {
	builder := s.sqlBuilder.Tenant().
		Delete(s.sqlBuilder.FullTableName("user")).
		Where("id = ?", id)

	result, err := s.sqlExecutor.ExecWith(builder)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to delete user")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.HandledWithMessage(err, "failed to delete user")
	}
	if rowsAffected == 0 {
		return authinfo.ErrNotFound
	}

	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authinfo.Store = &authInfoStore{}
)
