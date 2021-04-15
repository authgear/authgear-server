package verification

import (
	"database/sql"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
)

type StorePQ struct {
	SQLBuilder  *tenantdb.SQLBuilder
	SQLExecutor *tenantdb.SQLExecutor
}

func (s *StorePQ) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.Tenant().
		Select(
			"id",
			"user_id",
			"name",
			"value",
			"created_at",
		).
		From(s.SQLBuilder.TableName("_auth_verified_claim"))
}

func (s *StorePQ) scan(scn db.Scanner) (*Claim, error) {
	c := &Claim{}
	err := scn.Scan(
		&c.ID,
		&c.UserID,
		&c.Name,
		&c.Value,
		&c.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrClaimUnverified
	} else if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *StorePQ) ListByUser(userID string) ([]*Claim, error) {
	q := s.selectQuery().Where("user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(q)
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

func (s *StorePQ) ListByClaimName(userID string, claimName string) ([]*Claim, error) {
	q := s.selectQuery().Where("user_id = ? AND name = ?", userID, claimName)

	rows, err := s.SQLExecutor.QueryWith(q)
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

func (s *StorePQ) Get(userID string, claimName string, claimValue string) (*Claim, error) {
	q := s.selectQuery().Where("user_id = ? AND name = ? AND value = ?", userID, claimName, claimValue)

	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(row)
}

func (s *StorePQ) Create(claim *Claim) error {
	q := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.TableName("_auth_verified_claim")).
		Columns(
			"id",
			"user_id",
			"name",
			"value",
			"created_at",
		).
		Values(
			claim.ID,
			claim.UserID,
			claim.Name,
			claim.Value,
			claim.CreatedAt,
		)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *StorePQ) Delete(id string) error {
	q := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.TableName("_auth_verified_claim")).
		Where("id = ?", id)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *StorePQ) DeleteAll(userID string) error {
	q := s.SQLBuilder.Tenant().
		Delete(s.SQLBuilder.TableName("_auth_verified_claim")).
		Where("user_id = ?", userID)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}
