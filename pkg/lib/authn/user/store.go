package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"

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
	QueryPage(sortOption SortOption, pageArgs graphqlutil.PageArgs) ([]*User, uint64, error)
	UpdateLoginTime(userID string, loginAt time.Time) error
	UpdateDisabledStatus(userID string, isDisabled bool, reason *string) error
	UpdateStandardAttributes(userID string, stdAttrs map[string]interface{}) error
	Delete(userID string) error
}

type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
	Clock       clock.Clock
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
			"standard_attributes",
		).
		Values(
			u.ID,
			u.CreatedAt,
			u.UpdatedAt,
			u.MostRecentLoginAt,
			u.LessRecentLoginAt,
			u.IsDisabled,
			u.DisableReason,
			stdAttrsBytes,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"id",
			"created_at",
			"updated_at",
			"login_at",
			"last_login_at",
			"is_disabled",
			"disable_reason",
			"standard_attributes",
		).
		From(s.SQLBuilder.TableName("_auth_user"))
}

func (s *Store) scan(scn db.Scanner) (*User, error) {
	u := &User{}
	var stdAttrsBytes []byte

	if err := scn.Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.MostRecentLoginAt,
		&u.LessRecentLoginAt,
		&u.IsDisabled,
		&u.DisableReason,
		&stdAttrsBytes,
	); err != nil {
		return nil, err
	}

	if len(stdAttrsBytes) > 0 {
		if err := json.Unmarshal(stdAttrsBytes, &u.StandardAttributes); err != nil {
			return nil, err
		}
	}
	if u.StandardAttributes == nil {
		u.StandardAttributes = make(map[string]interface{})
	}

	return u, nil
}

func (s *Store) Get(userID string) (*User, error) {
	builder := s.selectQuery().Where("id = ?", userID)
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
	builder := s.selectQuery().Where("id = ANY (?)", pq.Array(userIDs))

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

func (s *Store) QueryPage(sortOption SortOption, pageArgs graphqlutil.PageArgs) ([]*User, uint64, error) {
	query := s.selectQuery()

	query = sortOption.Apply(query)

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
		Set("last_login_at", squirrel.Expr("login_at")).
		Set("login_at", loginAt).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateDisabledStatus(userID string, isDisabled bool, reason *string) error {
	builder := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_auth_user")).
		Set("is_disabled", isDisabled).
		Set("disable_reason", reason).
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
