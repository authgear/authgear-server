package main

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

// The whole point of this script is to copy the users from one project to another.
// We use a deterministric function to derive a new ID so that this script can run over and over again.

func openDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	return db, nil
}

var onConflictDoNothing = "ON CONFLICT DO NOTHING"

type Copier struct {
	Context    context.Context
	DB         *sql.DB
	Schema     string
	SQLBuilder db.SQLBuilder
	FromAppID  string
	ToAppID    string
}

func NewCopier() (*Copier, error) {
	database, err := openDB()
	if err != nil {
		return nil, err
	}

	schema := os.Getenv("DATABASE_SCHEMA")

	sqlBuilder := db.NewSQLBuilder(schema)

	fromAppIDPtr := flag.String("from-app-id", "", "from app id")
	toAppIDPtr := flag.String("to-app-id", "", "to app id")

	flag.Parse()

	return &Copier{
		Context:    context.Background(),
		DB:         database,
		Schema:     schema,
		SQLBuilder: sqlBuilder,
		FromAppID:  *fromAppIDPtr,
		ToAppID:    *toAppIDPtr,
	}, nil
}

// NewID computes a deterministric ID.
func (c *Copier) NewID(originalID string) string {
	id, err := uuid.Parse(originalID)
	if err != nil {
		panic(err)
	}

	hash := md5.New()
	hash.Write(id[:])
	hash.Write([]byte(c.ToAppID)[:])
	hashed := hash.Sum(nil)

	newID, err := uuid.FromBytes(hashed)
	if err != nil {
		panic(err)
	}

	return newID.String()
}

func (c *Copier) WithTx(f func(tx *sql.Tx) error) (err error) {
	tx, err := c.DB.BeginTx(c.Context, nil)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = f(tx)
	return
}

func (c *Copier) ListUsers() ([]string, error) {
	var userIDs []string

	err := c.WithTx(func(tx *sql.Tx) error {
		q := c.SQLBuilder.Select("id").From("_auth_user").Where("app_id = ?", c.FromAppID)

		err := c.Query(tx, q, func(rows *sql.Rows) error {
			var userID string
			err := rows.Scan(&userID)
			if err != nil {
				return err
			}
			userIDs = append(userIDs, userID)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (c *Copier) CopyUser(userID string) error {
	err := c.WithTx(func(tx *sql.Tx) error {
		fmt.Printf("copying user %v to %v\n", userID, c.NewID(userID))

		err := c.CopyUserTable(tx, userID)
		if err != nil {
			return err
		}

		err = c.CopyRecoveryCodeTable(tx, userID)
		if err != nil {
			return err
		}

		err = c.CopyPasswordHistoryTable(tx, userID)
		if err != nil {
			return err
		}

		err = c.CopyOAuthAuthorizationTable(tx, userID)
		if err != nil {
			return err
		}

		err = c.CopyVerifiedClaimTable(tx, userID)
		if err != nil {
			return err
		}

		err = c.CopyIdentity(tx, userID)
		if err != nil {
			return err
		}

		err = c.CopyAuthenticator(tx, userID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return err
}

func (c *Copier) CopyUserTable(tx *sql.Tx, userID string) error {
	columns := []string{
		"id",
		"app_id",
		"created_at",
		"updated_at",
		"last_login_at",
		"login_at",
		"is_disabled",
		"disable_reason",
		"standard_attributes",
		"custom_attributes",
		"is_deactivated",
		"delete_at",
		"is_anonymized",
		"anonymize_at",
		"anonymized_at",
	}
	tableName := "_auth_user"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ?", c.FromAppID, userID)

	var id string
	var appID string
	var created_at time.Time
	var updated_at time.Time
	var last_login_at sql.NullTime
	var login_at sql.NullTime
	var is_disabled bool
	var disable_reason sql.NullString
	var standard_attributes []byte
	var custom_attributes []byte
	var is_deactivated sql.NullBool
	var delete_at sql.NullTime
	var is_anonymized bool
	var anonymize_at sql.NullTime
	var anonymized_at sql.NullTime

	err := c.QueryRow(tx, q, func(row *sql.Row) error {
		return row.Scan(
			&id,
			&appID,
			&created_at,
			&updated_at,
			&last_login_at,
			&login_at,
			&is_disabled,
			&disable_reason,
			&standard_attributes,
			&custom_attributes,
			&is_deactivated,
			&delete_at,
			&is_anonymized,
			&anonymize_at,
			&anonymized_at,
		)
	})
	if err != nil {
		return err
	}

	insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
		c.NewID(id),
		c.ToAppID,
		created_at,
		updated_at,
		last_login_at,
		login_at,
		is_disabled,
		disable_reason,
		standard_attributes,
		custom_attributes,
		is_deactivated,
		delete_at,
		is_anonymized,
		anonymize_at,
		anonymized_at,
	).Suffix(onConflictDoNothing)

	err = c.Insert(tx, insert)
	if err != nil {
		return err
	}

	return nil
}

func (c *Copier) CopyRecoveryCodeTable(tx *sql.Tx, userID string) error {
	columns := []string{
		"id",
		"app_id",
		"user_id",
		"code",
		"created_at",
		"consumed",
		"updated_at",
	}
	tableName := "_auth_recovery_code"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND user_id = ?", c.FromAppID, userID)

	type Record struct {
		id         string
		app_id     string
		user_id    string
		code       string
		created_at time.Time
		consumed   bool
		updated_at time.Time
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.user_id,
			&record.code,
			&record.created_at,
			&record.consumed,
			&record.updated_at,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			c.NewID(userID),
			record.code,
			record.created_at,
			record.consumed,
			record.updated_at,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyPasswordHistoryTable(tx *sql.Tx, userID string) error {
	columns := []string{
		"id",
		"app_id",
		"created_at",
		"user_id",
		"password",
	}
	tableName := "_auth_password_history"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND user_id = ?", c.FromAppID, userID)

	type Record struct {
		id         string
		app_id     string
		created_at time.Time
		user_id    string
		password   string
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.created_at,
			&record.user_id,
			&record.password,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.created_at,
			c.NewID(userID),
			record.password,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyOAuthAuthorizationTable(tx *sql.Tx, userID string) error {
	columns := []string{
		"id",
		"app_id",
		"client_id",
		"user_id",
		"created_at",
		"updated_at",
		"scopes",
	}
	tableName := "_auth_oauth_authorization"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND user_id = ?", c.FromAppID, userID)

	type Record struct {
		id         string
		app_id     string
		client_id  string
		user_id    string
		created_at time.Time
		updated_at time.Time
		scopes     []byte
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.client_id,
			&record.user_id,
			&record.created_at,
			&record.updated_at,
			&record.scopes,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.client_id,
			c.NewID(userID),
			record.created_at,
			record.updated_at,
			record.scopes,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyVerifiedClaimTable(tx *sql.Tx, userID string) error {
	columns := []string{
		"id",
		"app_id",
		"user_id",
		"name",
		"value",
		"created_at",
	}
	tableName := "_auth_verified_claim"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND user_id = ?", c.FromAppID, userID)

	type Record struct {
		id         string
		app_id     string
		user_id    string
		name       string
		value      string
		created_at time.Time
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.user_id,
			&record.name,
			&record.value,
			&record.created_at,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			c.NewID(userID),
			record.name,
			record.value,
			record.created_at,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) ListIdentities(tx *sql.Tx, userID string) ([]string, error) {
	var identityIDs []string

	err := c.Query(tx, c.SQLBuilder.Select(
		"id",
	).
		From("_auth_identity").
		Where("app_id = ? AND user_id = ?", c.FromAppID, userID), func(rows *sql.Rows) error {
		var identityID string
		err := rows.Scan(&identityID)
		if err != nil {
			return err
		}
		identityIDs = append(identityIDs, identityID)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return identityIDs, nil
}

func (c *Copier) CopyIdentity(tx *sql.Tx, userID string) error {
	identityIDs, err := c.ListIdentities(tx, userID)
	if err != nil {
		return err
	}

	err = c.CopyIdentityTable(tx, userID, identityIDs)
	if err != nil {
		return err
	}

	err = c.CopyIdentityAnonymousTable(tx, userID, identityIDs)
	if err != nil {
		return err
	}

	err = c.CopyIdentityBiometricTable(tx, userID, identityIDs)
	if err != nil {
		return err
	}

	err = c.CopyIdentityLoginIDTable(tx, userID, identityIDs)
	if err != nil {
		return err
	}

	err = c.CopyIdentityOAuthTable(tx, userID, identityIDs)
	if err != nil {
		return err
	}

	err = c.CopyIdentityPasskeyTable(tx, userID, identityIDs)
	if err != nil {
		return err
	}

	err = c.CopyIdentitySIWETable(tx, userID, identityIDs)
	if err != nil {
		return err
	}

	return nil
}

func (c *Copier) CopyIdentityTable(tx *sql.Tx, userID string, identityIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"type",
		"user_id",
		"created_at",
		"updated_at",
	}
	tableName := "_auth_identity"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND user_id = ? AND id = ANY (?)", c.FromAppID, userID, pq.Array(identityIDs))

	type Record struct {
		id         string
		app_id     string
		typ        string
		user_id    string
		created_at time.Time
		updated_at time.Time
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.typ,
			&record.user_id,
			&record.created_at,
			&record.updated_at,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.typ,
			c.NewID(userID),
			record.created_at,
			record.updated_at,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyIdentityAnonymousTable(tx *sql.Tx, userID string, identityIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"key_id",
		"key",
	}
	tableName := "_auth_identity_anonymous"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(identityIDs))

	type Record struct {
		id     string
		app_id string
		key_id string
		key    []byte
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.key_id,
			&record.key,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.key_id,
			record.key,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyIdentityBiometricTable(tx *sql.Tx, userID string, identityIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"key_id",
		"key",
		"device_info",
	}
	tableName := "_auth_identity_biometric"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(identityIDs))

	type Record struct {
		id          string
		app_id      string
		key_id      string
		key         []byte
		device_info []byte
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.key_id,
			&record.key,
			&record.device_info,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.key_id,
			record.key,
			record.device_info,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyIdentityLoginIDTable(tx *sql.Tx, userID string, identityIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"login_id_key",
		"login_id",
		"claims",
		"original_login_id",
		"unique_key",
		"login_id_type",
	}
	tableName := "_auth_identity_login_id"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(identityIDs))

	type Record struct {
		id                string
		app_id            string
		login_id_key      string
		login_id          string
		claims            []byte
		original_login_id string
		unique_key        string
		login_id_type     string
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.login_id_key,
			&record.login_id,
			&record.claims,
			&record.original_login_id,
			&record.unique_key,
			&record.login_id_type,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.login_id_key,
			record.login_id,
			record.claims,
			record.original_login_id,
			record.unique_key,
			record.login_id_type,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyIdentityOAuthTable(tx *sql.Tx, userID string, identityIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"provider_type",
		"provider_keys",
		"provider_user_id",
		"claims",
		"profile",
	}
	tableName := "_auth_identity_oauth"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(identityIDs))

	type Record struct {
		id               string
		app_id           string
		provider_type    string
		provider_keys    []byte
		provider_user_id string
		claims           []byte
		profile          []byte
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.provider_type,
			&record.provider_keys,
			&record.provider_user_id,
			&record.claims,
			&record.profile,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.provider_type,
			record.provider_keys,
			record.provider_user_id,
			record.claims,
			record.profile,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyIdentityPasskeyTable(tx *sql.Tx, userID string, identityIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"credential_id",
		"creation_options",
		"attestation_response",
	}
	tableName := "_auth_identity_passkey"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(identityIDs))

	type Record struct {
		id                   string
		app_id               string
		credential_id        string
		creation_options     []byte
		attestation_response []byte
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.credential_id,
			&record.creation_options,
			&record.attestation_response,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.credential_id,
			record.creation_options,
			record.attestation_response,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyIdentitySIWETable(tx *sql.Tx, userID string, identityIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"chain_id",
		"address",
		"data",
	}
	tableName := "_auth_identity_siwe"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(identityIDs))

	type Record struct {
		id       string
		app_id   string
		chain_id int64
		address  string
		data     []byte
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.chain_id,
			&record.address,
			&record.data,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.chain_id,
			record.address,
			record.data,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) ListAuthenticators(tx *sql.Tx, userID string) ([]string, error) {
	var authenticatorIDs []string

	err := c.Query(tx, c.SQLBuilder.Select(
		"id",
	).
		From("_auth_authenticator").
		Where("app_id = ? AND user_id = ?", c.FromAppID, userID), func(rows *sql.Rows) error {
		var authenticatorID string
		err := rows.Scan(&authenticatorID)
		if err != nil {
			return err
		}
		authenticatorIDs = append(authenticatorIDs, authenticatorID)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return authenticatorIDs, nil
}

func (c *Copier) CopyAuthenticator(tx *sql.Tx, userID string) error {
	authenticatorIDs, err := c.ListAuthenticators(tx, userID)
	if err != nil {
		return err
	}

	err = c.CopyAuthenticatorTable(tx, userID, authenticatorIDs)
	if err != nil {
		return err
	}

	err = c.CopyAuthenticatorOOBTable(tx, userID, authenticatorIDs)
	if err != nil {
		return err
	}

	err = c.CopyAuthenticatorPasskeyTable(tx, userID, authenticatorIDs)
	if err != nil {
		return err
	}

	err = c.CopyAuthenticatorPasswordTable(tx, userID, authenticatorIDs)
	if err != nil {
		return err
	}

	err = c.CopyAuthenticatorTOTPTable(tx, userID, authenticatorIDs)
	if err != nil {
		return err
	}

	return nil
}

func (c *Copier) CopyAuthenticatorTable(tx *sql.Tx, userID string, authenticatorIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"type",
		"user_id",
		"created_at",
		"updated_at",
		"is_default",
		"kind",
	}
	tableName := "_auth_authenticator"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND user_id = ? AND id = ANY (?)", c.FromAppID, userID, pq.Array(authenticatorIDs))

	type Record struct {
		id         string
		app_id     string
		typ        string
		user_id    string
		created_at time.Time
		updated_at time.Time
		is_default bool
		kind       string
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.typ,
			&record.user_id,
			&record.created_at,
			&record.updated_at,
			&record.is_default,
			&record.kind,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.typ,
			c.NewID(userID),
			record.created_at,
			record.updated_at,
			record.is_default,
			record.kind,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyAuthenticatorOOBTable(tx *sql.Tx, userID string, authenticatorIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"phone",
		"email",
	}
	tableName := "_auth_authenticator_oob"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(authenticatorIDs))

	type Record struct {
		id     string
		app_id string
		phone  string
		email  string
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.phone,
			&record.email,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.phone,
			record.email,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyAuthenticatorPasskeyTable(tx *sql.Tx, userID string, authenticatorIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"credential_id",
		"creation_options",
		"attestation_response",
		"sign_count",
	}
	tableName := "_auth_authenticator_passkey"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(authenticatorIDs))

	type Record struct {
		id                   string
		app_id               string
		credential_id        string
		creation_options     []byte
		attestation_response []byte
		sign_count           int64
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.credential_id,
			&record.creation_options,
			&record.attestation_response,
			&record.sign_count,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.credential_id,
			record.creation_options,
			record.attestation_response,
			record.sign_count,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyAuthenticatorPasswordTable(tx *sql.Tx, userID string, authenticatorIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"password_hash",
	}
	tableName := "_auth_authenticator_password"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(authenticatorIDs))

	type Record struct {
		id            string
		app_id        string
		password_hash string
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.password_hash,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.password_hash,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) CopyAuthenticatorTOTPTable(tx *sql.Tx, userID string, authenticatorIDs []string) error {
	columns := []string{
		"id",
		"app_id",
		"secret",
		"display_name",
	}
	tableName := "_auth_authenticator_totp"

	q := c.SQLBuilder.Select(columns...).
		From(tableName).
		Where("app_id = ? AND id = ANY (?)", c.FromAppID, pq.Array(authenticatorIDs))

	type Record struct {
		id           string
		app_id       string
		secret       string
		display_name string
	}

	var records []Record

	err := c.Query(tx, q, func(rows *sql.Rows) error {
		var record Record
		err := rows.Scan(
			&record.id,
			&record.app_id,
			&record.secret,
			&record.display_name,
		)
		if err != nil {
			return err
		}

		records = append(records, record)
		return nil
	})
	if err != nil {
		return err
	}

	for _, record := range records {
		insert := c.SQLBuilder.Insert(tableName).Columns(columns...).Values(
			c.NewID(record.id),
			c.ToAppID,
			record.secret,
			record.display_name,
		).Suffix(onConflictDoNothing)

		err = c.Insert(tx, insert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) Query(tx *sql.Tx, q sq.Sqlizer, scan func(rows *sql.Rows) error) error {
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(c.Context, sqlStr, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = scan(rows)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Copier) QueryRow(tx *sql.Tx, q sq.Sqlizer, scan func(row *sql.Row) error) error {
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return err
	}

	row := tx.QueryRowContext(c.Context, sqlStr, args...)
	err = scan(row)
	if err != nil {
		return err
	}

	return nil
}

func (c *Copier) Insert(tx *sql.Tx, q sq.Sqlizer) error {
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(c.Context, sqlStr, args...)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	c, err := NewCopier()
	if err != nil {
		panic(err)
	}

	userIDs, err := c.ListUsers()
	if err != nil {
		panic(err)
	}

	for _, userID := range userIDs {
		err := c.CopyUser(userID)
		if err != nil {
			panic(err)
		}
	}
}
