package dcr

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

func (s *Store) NewInitialAccessToken(options *NewInitialAccessTokenOptions, tokenHash string) *InitialAccessToken {
	now := s.Clock.NowUTC()

	expiresIn := DefaultInitialAccessTokenExpiresIn
	if options.ExpiresIn != nil {
		expiresIn = *options.ExpiresIn
	}

	return &InitialAccessToken{
		ID:        uuid.NewString(),
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(expiresIn) * time.Second),
		Type:      options.Type,
		TokenHash: tokenHash,
	}
}

func (s *Store) CreateInitialAccessToken(ctx context.Context, t *InitialAccessToken) error {
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_oauth_initial_access_token")).
		Columns(
			"id",
			"created_at",
			"expires_at",
			"token_type",
			"token_hash",
		).
		Values(
			t.ID,
			t.CreatedAt,
			t.ExpiresAt,
			string(t.Type),
			t.TokenHash,
		)

	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetInitialAccessTokenByID(ctx context.Context, id string) (*InitialAccessToken, error) {
	q := s.selectInitialAccessTokenQuery().Where("id = ?", id)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	t, err := s.scanInitialAccessToken(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInitialAccessTokenNotFound
		}
		return nil, err
	}

	return t, nil
}

func (s *Store) GetInitialAccessTokenByHash(ctx context.Context, tokenHash string) (*InitialAccessToken, error) {
	q := s.selectInitialAccessTokenQuery().Where("token_hash = ?", tokenHash)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	t, err := s.scanInitialAccessToken(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInitialAccessTokenNotFound
		}
		return nil, err
	}

	return t, nil
}

func (s *Store) DeleteInitialAccessToken(ctx context.Context, id string) error {
	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_oauth_initial_access_token")).
		Where("id = ?", id)

	result, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrInitialAccessTokenNotFound
	}

	return nil
}

func (s *Store) ListActiveInitialAccessTokens(ctx context.Context) ([]*InitialAccessToken, error) {
	now := s.Clock.NowUTC()
	q := s.selectInitialAccessTokenQuery().
		Where("expires_at > ?", now).
		OrderBy("created_at DESC")

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*InitialAccessToken
	for rows.Next() {
		t, err := s.scanInitialAccessToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}

	return tokens, nil
}

func (s *Store) selectInitialAccessTokenQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"id",
			"created_at",
			"expires_at",
			"token_type",
			"token_hash",
		).
		From(s.SQLBuilder.TableName("_auth_oauth_initial_access_token"))
}

func (s *Store) scanInitialAccessToken(scanner db.Scanner) (*InitialAccessToken, error) {
	t := &InitialAccessToken{}

	err := scanner.Scan(
		&t.ID,
		&t.CreatedAt,
		&t.ExpiresAt,
		&t.Type,
		&t.TokenHash,
	)
	if err != nil {
		return nil, err
	}

	return t, nil
}
