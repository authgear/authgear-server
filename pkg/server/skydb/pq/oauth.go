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

func (c *conn) CreateOAuthInfo(oauthinfo *skydb.OAuthInfo) (err error) {
	var (
		createdAt *time.Time
		updatedAt *time.Time
	)
	createdAt = oauthinfo.CreatedAt
	if createdAt != nil && createdAt.IsZero() {
		createdAt = nil
	}
	updatedAt = oauthinfo.UpdatedAt
	if updatedAt != nil && updatedAt.IsZero() {
		updatedAt = nil
	}

	builder := psql.Insert(c.tableName("_sso_oauth")).Columns(
		"user_id",
		"provider",
		"principal_id",
		"token_response",
		"profile",
		"_created_at",
		"_updated_at",
	).Values(
		oauthinfo.UserID,
		oauthinfo.Provider,
		oauthinfo.PrincipalID,
		tokenResponseValue{oauthinfo.TokenResponse, true},
		providerProfileValue{oauthinfo.ProviderProfile, true},
		createdAt,
		updatedAt,
	)

	_, err = c.ExecWith(builder)
	if isUniqueViolated(err) {
		return skydb.ErrUserDuplicated
	}
	return err
}

func (c *conn) UpdateOAuthInfo(oauthinfo *skydb.OAuthInfo) (err error) {
	var (
		updatedAt *time.Time
	)
	updatedAt = oauthinfo.UpdatedAt
	if updatedAt != nil && updatedAt.IsZero() {
		updatedAt = nil
	}

	builder := psql.Update(c.tableName("_sso_oauth")).
		Set("token_response", tokenResponseValue{oauthinfo.TokenResponse, true}).
		Set("profile", providerProfileValue{oauthinfo.ProviderProfile, true}).
		Set("_updated_at", updatedAt).
		Where("provider = ? and principal_id = ?", oauthinfo.Provider, oauthinfo.PrincipalID)

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
	return nil
}

func (c *conn) oauthBuilder() sq.SelectBuilder {
	return psql.Select("user_id", "provider", "principal_id",
		"token_response", "profile", "_created_at", "_updated_at").
		From(c.tableName("_sso_oauth"))
}

func (c *conn) doScanOAuthInfo(oauthinfo *skydb.OAuthInfo, scanner sq.RowScanner) error {
	var (
		userID      string
		provider    string
		principalID string
		createdAt   pq.NullTime
		updatedAt   pq.NullTime
	)
	tokenResponse := tokenResponseValue{}
	providerProfile := providerProfileValue{}

	err := scanner.Scan(
		&userID,
		&provider,
		&principalID,
		&tokenResponse,
		&providerProfile,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		log.Infof(err.Error())
	}
	if err == sql.ErrNoRows {
		return skydb.ErrUserNotFound
	}

	oauthinfo.UserID = userID
	oauthinfo.Provider = provider
	oauthinfo.PrincipalID = principalID
	oauthinfo.TokenResponse = tokenResponse.TokenResponse
	oauthinfo.ProviderProfile = providerProfile.ProviderProfile

	if createdAt.Valid {
		oauthinfo.CreatedAt = &createdAt.Time
	} else {
		oauthinfo.CreatedAt = nil
	}
	if updatedAt.Valid {
		oauthinfo.UpdatedAt = &updatedAt.Time
	} else {
		oauthinfo.UpdatedAt = nil
	}

	return err
}

func (c *conn) GetOAuthInfo(provider string, principalID string, oauthinfo *skydb.OAuthInfo) error {
	builder := c.oauthBuilder().
		Where("provider = ? and principal_id = ?", provider, principalID)
	scanner := c.QueryRowWith(builder)
	return c.doScanOAuthInfo(oauthinfo, scanner)
}

func (c *conn) GetOAuthInfoByProviderAndUserID(provider string, userID string, oauthinfo *skydb.OAuthInfo) error {
	builder := c.oauthBuilder().
		Where("provider = ? and user_id = ?", provider, userID)
	scanner := c.QueryRowWith(builder)
	return c.doScanOAuthInfo(oauthinfo, scanner)
}

func (c *conn) DeleteOAuth(provider string, principalID string) error {
	builder := psql.Delete(c.tableName("_sso_oauth")).
		Where("provider = ? and principal_id = ?", provider, principalID)

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
