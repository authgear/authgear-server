package oob

import (
	"database/sql"
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *Store) Get(userID string, id string) (*Authenticator, error) {
	builder := s.SQLBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"ao.created_at",
			"ao.channel",
			"ao.phone",
			"ao.email",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_oob"),
			"ao",
			"a.id = ao.id",
		).
		Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	a := &Authenticator{}
	err = row.Scan(
		&a.ID,
		&a.UserID,
		&a.CreatedAt,
		&a.Channel,
		&a.Phone,
		&a.Email,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) GetByChannel(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string) (*Authenticator, error) {
	builder := s.SQLBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"ao.created_at",
			"ao.channel",
			"ao.phone",
			"ao.email",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_oob"),
			"ao",
			"a.id = ao.id",
		)

	switch channel {
	case authn.AuthenticatorOOBChannelSMS:
		builder = builder.Where("a.user_id = ? AND ao.channel = ? AND ao.phone = ?", userID, channel, phone)
	case authn.AuthenticatorOOBChannelEmail:
		builder = builder.Where("a.user_id = ? AND ao.channel = ? AND ao.email = ?", userID, channel, email)
	default:
		panic("oob: unknown channel")
	}

	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	a := &Authenticator{}
	err = row.Scan(
		&a.ID,
		&a.UserID,
		&a.CreatedAt,
		&a.Channel,
		&a.Phone,
		&a.Email,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) List(userID string) ([]*Authenticator, error) {
	builder := s.SQLBuilder.Tenant().
		Select(
			"a.id",
			"a.user_id",
			"ao.created_at",
			"ao.channel",
			"ao.phone",
			"ao.email",
		).
		From(s.SQLBuilder.FullTableName("authenticator"), "a").
		Join(
			s.SQLBuilder.FullTableName("authenticator_oob"),
			"ao",
			"a.id = ao.id",
		).
		Where("a.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authenticators []*Authenticator
	for rows.Next() {
		a := &Authenticator{}
		err = rows.Scan(
			&a.ID,
			&a.UserID,
			&a.CreatedAt,
			&a.Channel,
			&a.Phone,
			&a.Email,
		)
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, a)
	}

	return authenticators, nil
}

func (s *Store) Delete(id string) error {
	q := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("authenticator_oob")).
		Where("id = ?", id)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.FullTableName("authenticator")).
		Where("id = ?", id)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Create(a *Authenticator) error {
	q := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator")).
		Columns(
			"id",
			"type",
			"user_id",
		).
		Values(
			a.ID,
			authn.AuthenticatorTypeOOB,
			a.UserID,
		)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.FullTableName("authenticator_oob")).
		Columns(
			"id",
			"created_at",
			"channel",
			"phone",
			"email",
		).
		Values(
			a.ID,
			a.CreatedAt,
			a.Channel,
			a.Phone,
			a.Email,
		)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}
