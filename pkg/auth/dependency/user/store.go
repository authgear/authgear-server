package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type store interface {
	Create(u *User) error
	Get(userID string) (*User, error)
	UpdateMetadata(u *User) error
	UpdateLoginTime(u *User) error
}

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *Store) Create(u *User) error {
	var metadataBytes []byte
	metadataBytes, err := json.Marshal(u.Metadata)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("user")).
		Columns(
			"id",
			"created_at",
			"updated_at",
			"last_login_at",
			"metadata",
		).
		Values(
			u.ID,
			u.CreatedAt,
			u.UpdatedAt,
			u.LastLoginAt,
			metadataBytes,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
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
			"metadata",
		).
		From(s.SQLBuilder.FullTableName("user")).
		Where("id = ?", userID)
	scanner, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	var metadataBytes []byte
	u := &User{}
	err = scanner.Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
		&metadataBytes,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(metadataBytes, &u.Metadata)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) UpdateMetadata(u *User) error {
	var metadataBytes []byte
	metadataBytes, err := json.Marshal(u.Metadata)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.FullTableName("user")).
		Set("updated_at", u.UpdatedAt).
		Set("data", metadataBytes).
		Where("user_id = ?", u.ID)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateLoginTime(u *User) error {
	builder := s.SQLBuilder.Tenant().
		Update(s.SQLBuilder.FullTableName("user")).
		Set("last_login_at", u.LastLoginAt).
		Where("user_id = ?", u.ID)

	_, err := s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}
