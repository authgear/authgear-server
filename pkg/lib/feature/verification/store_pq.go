package verification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type StorePQ struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
}

func (s *StorePQ) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"id",
			"user_id",
			"name",
			"value",
			"created_at",
			"metadata",
		).
		From(s.SQLBuilder.TableName("_auth_verified_claim"))
}

func (s *StorePQ) scan(scn db.Scanner) (*Claim, error) {
	c := &Claim{}
	var rawMetadata []byte
	err := scn.Scan(
		&c.ID,
		&c.UserID,
		&c.Name,
		&c.Value,
		&c.CreatedAt,
		&rawMetadata,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrClaimUnverified
	} else if err != nil {
		return nil, err
	}

	if rawMetadata != nil {
		err = json.Unmarshal(rawMetadata, &c.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (s *StorePQ) ListByUserIDs(ctx context.Context, userIDs []string) ([]*Claim, error) {
	q := s.selectQuery().Where("user_id = ANY (?)", pq.Array(userIDs))

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []*Claim
	for rows.Next() {
		a, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, a)
	}

	return claims, nil
}

func (s *StorePQ) ListByUser(ctx context.Context, userID string) ([]*Claim, error) {
	return s.ListByUserIDs(ctx, []string{userID})
}

func (s *StorePQ) ListByUserIDsAndClaimNames(ctx context.Context, userIDs []string, claimNames []string) ([]*Claim, error) {
	q := s.selectQuery().Where("user_id = ANY (?) AND name = ANY (?)", pq.Array(userIDs), pq.Array(claimNames))

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []*Claim
	for rows.Next() {
		a, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, a)
	}

	return claims, nil
}

func (s *StorePQ) ListByClaimName(ctx context.Context, userID string, claimName string) ([]*Claim, error) {
	return s.ListByUserIDsAndClaimNames(ctx, []string{userID}, []string{claimName})
}

func (s *StorePQ) Get(ctx context.Context, userID string, claimName string, claimValue string) (*Claim, error) {
	q := s.selectQuery().Where("user_id = ? AND name = ? AND value = ?", userID, claimName, claimValue)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	return s.scan(row)
}

func (s *StorePQ) Create(ctx context.Context, claim *Claim) (err error) {
	rawMetadata, err := json.Marshal(claim.Metadata)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_verified_claim")).
		Columns(
			"id",
			"user_id",
			"name",
			"value",
			"created_at",
			"metadata",
		).
		Values(
			claim.ID,
			claim.UserID,
			claim.Name,
			claim.Value,
			claim.CreatedAt,
			rawMetadata,
		)
	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *StorePQ) Delete(ctx context.Context, id string) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_verified_claim")).
		Where("id = ?", id)
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *StorePQ) DeleteAll(ctx context.Context, userID string) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_verified_claim")).
		Where("user_id = ?", userID)
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	return nil
}
