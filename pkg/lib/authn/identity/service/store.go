package service

import (
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
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

	return s.listRefs(builder)
}
func (s *Store) ListRefsByIDs(ids []string) ([]*model.IdentityRef, error) {
	builder := s.SQLBuilder.
		Select("id", "type", "user_id", "created_at", "updated_at").
		Where("id = ANY (?)", pq.Array(ids)).
		From(s.SQLBuilder.TableName("_auth_identity"))

	return s.listRefs(builder)
}

func (s *Store) GetRefByID(id string) (*model.IdentityRef, error) {
	builder := s.SQLBuilder.
		Select("id", "type", "user_id", "created_at", "updated_at").
		Where("id = ?", id).
		From(s.SQLBuilder.TableName("_auth_identity"))

	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	ref, err := s.scan(row)
	if err != nil {
		return nil, err
	}

	return ref, nil
}

func (s *Store) listRefs(builder db.SelectBuilder) ([]*model.IdentityRef, error) {
	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*model.IdentityRef
	for rows.Next() {
		var ref *model.IdentityRef
		ref, err = s.scan(rows)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}

	return refs, nil
}

func (s *Store) scan(scanner db.Scanner) (*model.IdentityRef, error) {
	ref := &model.IdentityRef{}
	err := scanner.Scan(
		&ref.ID,
		&ref.Type,
		&ref.UserID,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return ref, nil
}
