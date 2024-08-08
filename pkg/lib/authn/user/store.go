package user

import (
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
	Create(u *User) error
	Get(userID string) (*User, error)
	GetByIDs(userIDs []string) ([]*User, error)
	Count() (uint64, error)
	QueryPage(listOption ListOptions, pageArgs graphqlutil.PageArgs) ([]*User, uint64, error)
	UpdateLoginTime(userID string, loginAt time.Time) error
	UpdateMFAEnrollment(userID string, endAt *time.Time) error
	UpdateAccountStatus(userID string, status AccountStatus) error
	UpdateStandardAttributes(userID string, stdAttrs map[string]interface{}) error
	UpdateCustomAttributes(userID string, customAttrs map[string]interface{}) error
	Delete(userID string) error
	Anonymize(userID string) error
}

type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
	Clock       clock.Clock
	AppID       config.AppID
}

func (s *Store) Create(u *User) (err error) {
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
			"is_disabled",
			"disable_reason",
			"is_deactivated",
			"delete_at",
			"is_anonymized",
			"anonymize_at",
			"standard_attributes",
			"custom_attributes",
			"require_reindex_after",
			"mfa_grace_period_end_at",
		).
		Values(
			u.ID,
			u.CreatedAt,
			u.UpdatedAt,
			u.MostRecentLoginAt,
			u.LessRecentLoginAt,
			u.IsDisabled,
			u.DisableReason,
			u.IsDeactivated,
			u.DeleteAt,
			u.IsAnonymized,
			u.AnonymizeAt,
			stdAttrsBytes,
			customAttrsBytes,
			u.RequireReindexAfter,
			u.MFAGracePeriodtEndAt,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
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
				"is_disabled",
				"disable_reason",
				"is_deactivated",
				"delete_at",
				"is_anonymized",
				"anonymize_at",
				"last_indexed_at",
				"require_reindex_after",
				"standard_attributes",
				"custom_attributes",
				"mfa_grace_period_end_at",
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
			fieldWithAlias("is_disabled"),
			fieldWithAlias("disable_reason"),
			fieldWithAlias("is_deactivated"),
			fieldWithAlias("delete_at"),
			fieldWithAlias("is_anonymized"),
			fieldWithAlias("anonymize_at"),
			fieldWithAlias("last_indexed_at"),
			fieldWithAlias("require_reindex_after"),
			fieldWithAlias("standard_attributes"),
			fieldWithAlias("custom_attributes"),
			fieldWithAlias("mfa_grace_period_end_at"),
		).
		From(s.SQLBuilder.TableName("_auth_user"), alias)
}

func (s *Store) scan(scn db.Scanner) (*User, error) {
	u := &User{}
	var stdAttrsBytes []byte
	var customAttrsBytes []byte
	var isDeactivated sql.NullBool

	if err := scn.Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.MostRecentLoginAt,
		&u.LessRecentLoginAt,
		&u.IsDisabled,
		&u.DisableReason,
		&isDeactivated,
		&u.DeleteAt,
		&u.IsAnonymized,
		&u.AnonymizeAt,
		&u.LastIndexedAt,
		&u.RequireReindexAfter,
		&stdAttrsBytes,
		&customAttrsBytes,
		&u.MFAGracePeriodtEndAt,
	); err != nil {
		return nil, err
	}
	u.IsDeactivated = isDeactivated.Bool

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

	if u.StandardAttributes == nil {
		u.StandardAttributes = make(map[string]interface{})
	}
	if u.CustomAttributes == nil {
		u.CustomAttributes = make(map[string]interface{})
	}

	return u, nil
}

func (s *Store) Get(userID string) (*User, error) {
	builder := s.selectQuery("").Where("id = ?", userID)
	scanner, err := s.SQLExecutor.QueryRowWith(builder)
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

func (s *Store) GetByIDs(userIDs []string) ([]*User, error) {
	builder := s.selectQuery("").Where("id = ANY (?)", pq.Array(userIDs))

	rows, err := s.SQLExecutor.QueryWith(builder)
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

func (s *Store) Count() (uint64, error) {
	builder := s.SQLBuilder.
		Select("count(*)").
		From(s.SQLBuilder.TableName("_auth_user"))
	scanner, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return 0, err
	}

	var count uint64
	if err = scanner.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) QueryPage(listOption ListOptions, pageArgs graphqlutil.PageArgs) ([]*User, uint64, error) {
	query := s.selectQuery("u")

	query = listOption.SortOption.Apply(query)

	query, offset, err := db.ApplyPageArgs(query, pageArgs)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.SQLExecutor.QueryWith(query)
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

func (s *Store) UpdateLoginTime(userID string, loginAt time.Time) error {
	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("last_login_at", sq.Expr("login_at")).
		Set("login_at", loginAt).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateMFAEnrollment(userID string, endAt *time.Time) error {
	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("mfa_grace_period_end_at", endAt).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateAccountStatus(userID string, accountStatus AccountStatus) error {
	now := s.Clock.NowUTC()

	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("is_disabled", accountStatus.IsDisabled).
		Set("disable_reason", accountStatus.DisableReason).
		Set("is_deactivated", accountStatus.IsDeactivated).
		Set("delete_at", accountStatus.DeleteAt).
		Set("is_anonymized", accountStatus.IsAnonymized).
		Set("anonymize_at", accountStatus.AnonymizeAt).
		Set("updated_at", now).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateStandardAttributes(userID string, stdAttrs map[string]interface{}) error {
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

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateCustomAttributes(userID string, customAttrs map[string]interface{}) error {
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

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(userID string) error {
	builder := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_user")).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Anonymize(userID string) error {
	now := s.Clock.NowUTC()

	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("is_disabled", true).
		Set("is_anonymized", true).
		Set("anonymized_at", now).
		Set("standard_attributes", nil).
		Set("custom_attributes", nil).
		Set("updated_at", now).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) MarkAsReindexRequired(userIDs []string) error {
	now := s.Clock.NowUTC()
	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("require_reindex_after", now).
		Where("id = ANY (?)", pq.Array(userIDs))

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateLastIndexedAt(userIDs []string, at time.Time) error {
	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("last_indexed_at", at).
		Where("id = ANY (?)", pq.Array(userIDs))

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}
