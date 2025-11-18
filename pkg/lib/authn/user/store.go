package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type store interface {
	Create(ctx context.Context, u *User) error
	Get(ctx context.Context, userID string) (*User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]*User, error)
	Count(ctx context.Context) (uint64, error)
	QueryPage(ctx context.Context, listOption ListOptions, pageArgs graphqlutil.PageArgs) ([]*User, uint64, error)
	QueryForExport(ctx context.Context, offset uint64, limit uint64) ([]*User, error)
	UpdateLoginTime(ctx context.Context, userID string, loginAt time.Time) error
	UpdateMFAEnrollment(ctx context.Context, userID string, endAt *time.Time) error
	UpdateAccountStatus(ctx context.Context, userID string, status AccountStatusWithRefTime) error
	UpdateStandardAttributes(ctx context.Context, userID string, stdAttrs map[string]interface{}) error
	UpdateCustomAttributes(ctx context.Context, userID string, customAttrs map[string]interface{}) error
	UpdateOptOutPasskeyUpselling(ctx context.Context, userID string, optout bool) error
	Delete(ctx context.Context, userID string) error
	SetAllAttributesToNull(ctx context.Context, userID string) error
}

type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
	Clock       clock.Clock
	AppID       config.AppID
}

var _ store = &Store{}

//nolint:gosec
const keyOptOutPasskeyUpselling = "opt_out_passkey_upselling"
const keyScheduleDeletionReason = "schedule_deletion_reason"

func (s *Store) Create(ctx context.Context, u *User) (err error) {
	stdAttrs := u.StandardAttributes
	if stdAttrs == nil {
		stdAttrs = make(map[string]interface{})
	}

	stdAttrsBytes, err := json.Marshal(stdAttrs)
	if err != nil {
		return
	}

	customAttrs := u.CustomAttributes
	if customAttrs == nil {
		customAttrs = make(map[string]interface{})
	}

	customAttrsBytes, err := json.Marshal(customAttrs)
	if err != nil {
		return
	}

	builder := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_user")).
		Columns(
			"id",
			"created_at",
			"updated_at",
			"login_at",
			"last_login_at",

			"standard_attributes",
			"custom_attributes",
			"require_reindex_after",
			"last_indexed_at",
			"mfa_grace_period_end_at",

			"is_disabled",
			"account_status_stale_from",
			"is_indefinitely_disabled",
			"is_deactivated",
			"disable_reason",
			"temporarily_disabled_from",
			"temporarily_disabled_until",
			"account_valid_from",
			"account_valid_until",
			"delete_at",
			"anonymize_at",
			"anonymized_at",
			"is_anonymized",
		).
		Values(
			u.ID,
			u.CreatedAt,
			u.UpdatedAt,
			u.MostRecentLoginAt,
			u.LessRecentLoginAt,

			stdAttrsBytes,
			customAttrsBytes,
			u.RequireReindexAfter,
			u.LastIndexedAt,
			u.MFAGracePeriodtEndAt,

			u.isDisabled,
			u.accountStatusStaleFrom,
			u.isIndefinitelyDisabled,
			u.isDeactivated,
			u.disableReason,
			u.temporarilyDisabledFrom,
			u.temporarilyDisabledUntil,
			u.accountValidFrom,
			u.accountValidUntil,
			u.deleteAt,
			u.anonymizeAt,
			u.anonymizedAt,
			u.isAnonymized,
		)

	_, err = s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) selectQuery(alias string) db.SelectBuilder {
	if alias == "" {
		return s.SQLBuilder.
			Select(
				"id",
				"created_at",
				"updated_at",
				"login_at",
				"last_login_at",

				"standard_attributes",
				"custom_attributes",
				"metadata",
				"require_reindex_after",
				"last_indexed_at",
				"mfa_grace_period_end_at",

				"is_disabled",
				"account_status_stale_from",
				"is_indefinitely_disabled",
				"is_deactivated",
				"disable_reason",
				"temporarily_disabled_from",
				"temporarily_disabled_until",
				"account_valid_from",
				"account_valid_until",
				"delete_at",
				"anonymize_at",
				"anonymized_at",
				"is_anonymized",
			).
			From(s.SQLBuilder.TableName("_auth_user"))
	}
	fieldWithAlias := func(field string) string {
		return fmt.Sprintf("%s.%s", alias, field)
	}
	return s.SQLBuilder.
		Select(
			fieldWithAlias("id"),
			fieldWithAlias("created_at"),
			fieldWithAlias("updated_at"),
			fieldWithAlias("login_at"),
			fieldWithAlias("last_login_at"),

			fieldWithAlias("standard_attributes"),
			fieldWithAlias("custom_attributes"),
			fieldWithAlias("metadata"),
			fieldWithAlias("require_reindex_after"),
			fieldWithAlias("last_indexed_at"),
			fieldWithAlias("mfa_grace_period_end_at"),

			fieldWithAlias("is_disabled"),
			fieldWithAlias("account_status_stale_from"),
			fieldWithAlias("is_indefinitely_disabled"),
			fieldWithAlias("is_deactivated"),
			fieldWithAlias("disable_reason"),
			fieldWithAlias("temporarily_disabled_from"),
			fieldWithAlias("temporarily_disabled_until"),
			fieldWithAlias("account_valid_from"),
			fieldWithAlias("account_valid_until"),
			fieldWithAlias("delete_at"),
			fieldWithAlias("anonymize_at"),
			fieldWithAlias("anonymized_at"),
			fieldWithAlias("is_anonymized"),
		).
		From(s.SQLBuilder.TableName("_auth_user"), alias)
}

func (s *Store) scan(scn db.Scanner) (*User, error) {
	u := &User{}
	var stdAttrsBytes []byte
	var customAttrsBytes []byte
	var metadataBytes []byte

	if err := scn.Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.MostRecentLoginAt,
		&u.LessRecentLoginAt,

		&stdAttrsBytes,
		&customAttrsBytes,
		&metadataBytes,
		&u.RequireReindexAfter,
		&u.LastIndexedAt,
		&u.MFAGracePeriodtEndAt,

		&u.isDisabled,
		&u.accountStatusStaleFrom,
		&u.isIndefinitelyDisabled,
		&u.isDeactivated,
		&u.disableReason,
		&u.temporarilyDisabledFrom,
		&u.temporarilyDisabledUntil,
		&u.accountValidFrom,
		&u.accountValidUntil,
		&u.deleteAt,
		&u.anonymizeAt,
		&u.anonymizedAt,
		&u.isAnonymized,
	); err != nil {
		return nil, err
	}

	if len(stdAttrsBytes) > 0 {
		if err := json.Unmarshal(stdAttrsBytes, &u.StandardAttributes); err != nil {
			return nil, err
		}
	}
	if len(customAttrsBytes) > 0 {
		if err := json.Unmarshal(customAttrsBytes, &u.CustomAttributes); err != nil {
			return nil, err
		}
	}

	var metadata map[string]interface{}
	if len(metadataBytes) > 0 {
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, err
		}
	}
	if v, ok := metadata[keyOptOutPasskeyUpselling].(bool); ok {
		u.OptOutPasskeyUpsell = v
	} else {
		u.OptOutPasskeyUpsell = false
	}

	if v, ok := metadata[keyScheduleDeletionReason].(string); ok {
		u.deleteReason = &v
	} else {
		u.deleteReason = nil
	}

	if u.StandardAttributes == nil {
		u.StandardAttributes = make(map[string]interface{})
	}
	if u.CustomAttributes == nil {
		u.CustomAttributes = make(map[string]interface{})
	}

	return u, nil
}

func (s *Store) Get(ctx context.Context, userID string) (*User, error) {
	builder := s.selectQuery("").Where("id = ?", userID)
	scanner, err := s.SQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return nil, err
	}

	u, err := s.scan(scanner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) GetByIDs(ctx context.Context, userIDs []string) ([]*User, error) {
	builder := s.selectQuery("").Where("id = ANY (?)", pq.Array(userIDs))

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (s *Store) Count(ctx context.Context) (uint64, error) {
	builder := s.SQLBuilder.
		Select("count(*)").
		From(s.SQLBuilder.TableName("_auth_user"))
	scanner, err := s.SQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return 0, err
	}

	var count uint64
	if err = scanner.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) QueryPage(ctx context.Context, listOption ListOptions, pageArgs graphqlutil.PageArgs) ([]*User, uint64, error) {
	query := s.selectQuery("u")

	query = listOption.SortOption.Apply(query, "")

	query, offset, err := db.ApplyPageArgs(query, pageArgs)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.SQLExecutor.QueryWith(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := s.scan(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, offset, nil
}

func (s *Store) QueryForExport(ctx context.Context, offset uint64, limit uint64) ([]*User, error) {
	// created_at indexed as DESC NULLS LAST, to re use the index but in invented direction, need to use ASC NULLS FIRST
	query := s.selectQuery("u").Offset(offset).Limit(limit).OrderBy("created_at ASC NULLS FIRST")

	rows, err := s.SQLExecutor.QueryWith(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (s *Store) UpdateLoginTime(ctx context.Context, userID string, loginAt time.Time) error {
	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("last_login_at", sq.Expr("login_at")).
		Set("login_at", loginAt).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateMFAEnrollment(ctx context.Context, userID string, endAt *time.Time) error {
	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("mfa_grace_period_end_at", endAt).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateAccountStatus(ctx context.Context, userID string, accountStatus AccountStatusWithRefTime) error {
	now := s.Clock.NowUTC()

	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("is_disabled", accountStatus.accountStatus.isDisabled).
		Set("account_status_stale_from", accountStatus.accountStatus.accountStatusStaleFrom).
		Set("is_indefinitely_disabled", accountStatus.accountStatus.isIndefinitelyDisabled).
		Set("is_deactivated", accountStatus.accountStatus.isDeactivated).
		Set("disable_reason", accountStatus.accountStatus.disableReason).
		Set("temporarily_disabled_from", accountStatus.accountStatus.temporarilyDisabledFrom).
		Set("temporarily_disabled_until", accountStatus.accountStatus.temporarilyDisabledUntil).
		Set("account_valid_from", accountStatus.accountStatus.accountValidFrom).
		Set("account_valid_until", accountStatus.accountStatus.accountValidUntil).
		Set("delete_at", accountStatus.accountStatus.deleteAt).
		Set("metadata", sq.Expr(
			fmt.Sprintf("jsonb_set(coalesce(metadata, '{}'::jsonb), '{%s}', ?::jsonb, true)", keyScheduleDeletionReason),
			accountStatus.accountStatus.deleteReason,
		)).
		Set("anonymize_at", accountStatus.accountStatus.anonymizeAt).
		Set("anonymized_at", accountStatus.accountStatus.anonymizedAt).
		Set("is_anonymized", accountStatus.accountStatus.isAnonymized).
		Set("updated_at", now).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateStandardAttributes(ctx context.Context, userID string, stdAttrs map[string]interface{}) error {
	now := s.Clock.NowUTC()

	if stdAttrs == nil {
		stdAttrs = make(map[string]interface{})
	}

	stdAttrsBytes, err := json.Marshal(stdAttrs)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("standard_attributes", stdAttrsBytes).
		Set("updated_at", now).
		Where("id = ?", userID)

	_, err = s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateCustomAttributes(ctx context.Context, userID string, customAttrs map[string]interface{}) error {
	now := s.Clock.NowUTC()

	if customAttrs == nil {
		customAttrs = make(map[string]interface{})
	}

	customAttrsBytes, err := json.Marshal(customAttrs)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("custom_attributes", customAttrsBytes).
		Set("updated_at", now).
		Where("id = ?", userID)

	_, err = s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, userID string) error {
	builder := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_user")).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SetAllAttributesToNull(ctx context.Context, userID string) error {
	now := s.Clock.NowUTC()

	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("standard_attributes", nil).
		Set("custom_attributes", nil).
		Set("updated_at", now).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) MarkAsReindexRequired(ctx context.Context, userIDs []string) error {
	now := s.Clock.NowUTC()
	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("require_reindex_after", now).
		Where("id = ANY (?)", pq.Array(userIDs))

	_, err := s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateLastIndexedAt(ctx context.Context, userIDs []string, at time.Time) error {
	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("last_indexed_at", at).
		Where("id = ANY (?)", pq.Array(userIDs))

	_, err := s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateOptOutPasskeyUpselling(ctx context.Context, userID string, optout bool) error {
	now := s.Clock.NowUTC()

	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("metadata", sq.Expr(
			fmt.Sprintf("jsonb_set(coalesce(metadata, '{}'::jsonb), '{%s}', ?::jsonb, true)", keyOptOutPasskeyUpselling),
			optout,
		)).
		Set("updated_at", now).
		Where("id = ?", userID)
	_, err := s.SQLExecutor.ExecWith(ctx, builder)
	return err
}
