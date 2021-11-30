package service

import (
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
}

func (s *Store) Count(userID string) (uint64, error) {
	builder := s.SQLBuilder.
		Select("count(*)").
		Where("user_id = ?", userID).
		From(s.SQLBuilder.TableName("_auth_identity"))
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

func (s *Store) ListRefsByUsers(userIDs []string) ([]*model.IdentityRef, error) {
	builder := s.SQLBuilder.
		Select("id", "type", "user_id", "created_at", "updated_at").
		Where("user_id = ANY (?)", pq.Array(userIDs)).
		From(s.SQLBuilder.TableName("_auth_identity"))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*model.IdentityRef
	for rows.Next() {
		ref := &model.IdentityRef{}
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
