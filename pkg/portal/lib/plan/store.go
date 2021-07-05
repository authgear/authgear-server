package plan

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type Store struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *Store) GetPlan(name string) (*model.Plan, error) {
	q := s.selectQuery().Where("name = ?", name)
	row, err := s.SQLExecutor.QueryRowWith(q)
	if err != nil {
		return nil, err
	}
	return s.scan(row)
}

func (s *Store) selectQuery() db.SelectBuilder {
	return s.SQLBuilder.Global().
		Select(
			"id",
			"name",
			"feature_config",
		).
		From(s.SQLBuilder.TableName("_portal_plan"))
}

func (s *Store) scan(scn db.Scanner) (*model.Plan, error) {
	p := &model.Plan{}

	var data []byte
	err := scn.Scan(
		&p.ID,
		&p.Name,
		&data,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPlanNotFound
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &p.RawFeatureConfig)
	if err != nil {
		return nil, err
	}

	return p, nil
}
