package face_recognition

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"a.id",
			"a.user_id",
			"a.created_at",
			"a.updated_at",
			"a.is_default",
			"a.kind",
			"afr.opencv_fr_person_id",
		).
		From(s.SQLBuilder.TableName("_auth_authenticator"), "a").
		Join(
			s.SQLBuilder.TableName("_auth_authenticator_face_recognition"),
			"afr",
			"a.id = afr.id",
		)
}

func (s *Store) scan(scn db.Scanner) (*authenticator.FaceRecognition, error) {
	a := &authenticator.FaceRecognition{}

	err := scn.Scan(
		&a.ID,
		&a.UserID,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.IsDefault,
		&a.Kind,
		&a.OpenCVFRPersonID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authenticator.ErrAuthenticatorNotFound
	} else if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *Store) GetMany(ids []string) ([]*authenticator.FaceRecognition, error) {
	builder := s.selectQuery().Where("a.id = ANY (?)", pq.Array(ids))

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var as []*authenticator.FaceRecognition
	for rows.Next() {
		a, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}

	return as, nil
}

func (s *Store) Get(userID string, id string) (*authenticator.FaceRecognition, error) {
	q := s.selectQuery().Where("a.user_id = ? AND a.id = ?", userID, id)

	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}

	return s.scan(row)
}

func (s *Store) List(userID string) ([]*authenticator.FaceRecognition, error) {
	q := s.selectQuery().Where("a.user_id = ?", userID)

	rows, err := s.SQLExecutor.QueryWith(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authenticators []*authenticator.FaceRecognition
	for rows.Next() {
		a, err := s.scan(rows)
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, a)
	}

	return authenticators, nil
}

func (s *Store) Delete(id string) error {
	q := s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_authenticator_face_recognition")).
		Where("id = ?", id)
	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Delete(s.SQLBuilder.TableName("_auth_authenticator")).
		Where("id = ?", id)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Create(a *authenticator.FaceRecognition) (err error) {
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_authenticator")).
		Columns(
			"id",
			"type",
			"user_id",
			"created_at",
			"updated_at",
			"is_default",
			"kind",
		).
		Values(
			a.ID,
			model.AuthenticatorTypeFaceRecognition,
			a.UserID,
			a.CreatedAt,
			a.UpdatedAt,
			a.IsDefault,
			a.Kind,
		)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	q = s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_authenticator_face_recognition")).
		Columns(
			"id",
			"opencv_fr_person_id",
		).
		Values(
			a.ID,
			a.OpenCVFRPersonID,
		)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}
