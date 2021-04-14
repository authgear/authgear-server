package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type store interface {
	Create(u *User) error
	Get(userID string) (*User, error)
	GetByIDs(userIDs []string) ([]*User, error)
	Count() (uint64, error)
	QueryPage(after, before model.PageCursor, first, last *uint64) ([]*User, uint64, error)
	UpdateLoginTime(userID string, loginAt time.Time) error
	UpdateDisabledStatus(userID string, isDisabled bool, reason *string) error
	Delete(userID string) error
}

var queryPage = db.QueryPage(db.QueryPageConfig{
	KeyColumn: "created_at",
	IDColumn:  "id",
})

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *Store) Create(u *User) error {
	labels, err := json.Marshal(u.Labels)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.TableName("_auth_user")).
		Columns(
			"id",
			"labels",
			"created_at",
			"updated_at",
			"last_login_at",
			"is_disabled",
			"disable_reason",
		).
		Values(
			u.ID,
			labels,
			u.CreatedAt,
			u.UpdatedAt,
			u.LastLoginAt,
			u.IsDisabled,
			u.DisableReason,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.Tenant().
		Select(
			"id",
			"labels",
			"created_at",
			"updated_at",
			"last_login_at",
			"is_disabled",
			"disable_reason",
		).
		From(s.SQLBuilder.TableName("_auth_user"))
}

func (s *Store) scan(scn db.Scanner) (*User, error) {
	u := &User{}

	var labels []byte

	if err := scn.Scan(
		&u.ID,
		&labels,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
		&u.IsDisabled,
		&u.DisableReason,
	); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(labels, &u.Labels); err != nil {
		return nil, err
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
	builder := s.SQLBuilder.Tenant().
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

func (s *Store) QueryPage(after, before model.PageCursor, first, last *uint64) ([]*User, uint64, error) {
	afterKey, err := db.NewFromPageCursor(after)
	if err != nil {
		return nil, 0, err
	}
	beforeKey, err := db.NewFromPageCursor(before)
	if err != nil {
		return nil, 0, err
	}

	selectQuery := s.selectQuery()

	query, offset, err := queryPage(selectQuery, afterKey, beforeKey, first, last)
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
	builder := s.SQLBuilder.Tenant().
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
	builder := s.SQLBuilder.Tenant().
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

func (s *Store) Delete(userID string) error {
	builder := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.TableName("_auth_user")).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}
