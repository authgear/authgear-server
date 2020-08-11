package user

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/db"
)

type store interface {
	Create(u *User) error
	Get(userID string) (*User, error)
	UpdateLoginTime(userID string, loginAt time.Time) error
}

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *Store) Create(u *User) error {
	builder := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("user")).
		Columns(
			"id",
			"created_at",
			"updated_at",
			"last_login_at",
		).
		Values(
			u.ID,
			u.CreatedAt,
			u.UpdatedAt,
			u.LastLoginAt,
		)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Get(userID string) (*User, error) {
	builder := s.SQLBuilder.Tenant().
		Select(
			"id",
			"created_at",
			"updated_at",
			"last_login_at",
		).
		From(s.SQLBuilder.FullTableName("user")).
		Where("id = ?", userID)
	scanner, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	u := &User{}
	err = scanner.Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) UpdateLoginTime(userID string, loginAt time.Time) error {
	builder := s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.FullTableName("user")).
		Set("last_login_at", squirrel.Expr("login_at")).
		Set("login_at", loginAt).
		Where("id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}
