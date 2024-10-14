package opencvfr

import (
	"database/sql"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type AuthgearAppIDOpenCVFRCollectionIDMap struct {
	ID                   string `json:"id"`
	AppID                string `json:"app_id"`
	OpenCVFRCollectionID string `json:"opencv_fr_collection_id"`
}

// Store contains a mapping between authgear-app-id and opencvfr-collection-id
type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
}

func (s *Store) Get(appID string) (m *AuthgearAppIDOpenCVFRCollectionIDMap, err error) {
	builder := s.selectQuery() // no need specify app_id because db.SQLBuilder does that already

	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	return s.scan(row)
}

func (s *Store) Create(m *AuthgearAppIDOpenCVFRCollectionIDMap) (err error) {
	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_authenticator_face_recognition_opencvfr_collection_map")).
		Columns(
			"id",
			"app_id",
			"opencv_fr_collection_id",
		).
		Values(
			m.ID,
			m.AppID,
			m.OpenCVFRCollectionID,
		)

	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) scan(scn db.Scanner) (*AuthgearAppIDOpenCVFRCollectionIDMap, error) {
	m := &AuthgearAppIDOpenCVFRCollectionIDMap{}

	err := scn.Scan(
		&m.ID,
		&m.AppID,
		&m.OpenCVFRCollectionID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // expected, not an error
	} else if err != nil {
		return nil, err
	}

	return m, nil

}
func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"m.id",
			"m.app_id",
			"m.opencv_fr_collection_id",
		).
		From(s.SQLBuilder.TableName("_auth_authenticator_face_recognition_opencvfr_collection_map"), "m")
}
