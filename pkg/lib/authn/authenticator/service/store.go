package service

import (
	"github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor
}

func (s *Store) Count(userID string) (uint64, error) {
	builder := s.SQLBuilder.Tenant().
		Select("count(*)").
		Where("user_id = ?", userID).
		From(s.SQLBuilder.FullTableName("authenticator"))
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

func (s *Store) ListRefsByUsers(userIDs []string) ([]*authenticator.Ref, error) {
	ids := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		ids[i] = id
	}
	builder := s.SQLBuilder.Tenant().
		Select("id", "type", "user_id", "created_at", "updated_at").
		Where("user_id IN ("+squirrel.Placeholders(len(ids))+")", ids...).
		From(s.SQLBuilder.FullTableName("authenticator"))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*authenticator.Ref
	for rows.Next() {
		ref := &authenticator.Ref{}
		if err := rows.Scan(
			&ref.ID,
			&ref.Type,
			&ref.UserID,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}

	return refs, nil
}
