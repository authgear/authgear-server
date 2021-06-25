package plan

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Store struct {
	Clock       clock.Clock
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

func (s *Store) Create(plan *model.Plan) error {
	configData, err := json.Marshal(plan.RawFeatureConfig)
	if err != nil {
		return err
	}
	q := s.SQLBuilder.Global().
		Insert(s.SQLBuilder.TableName("_portal_plan")).
		Columns(
			"id",
			"name",
			"feature_config",
			"created_at",
			"updated_at",
		).
		Values(
			plan.ID,
			plan.Name,
			configData,
			s.Clock.NowUTC(),
			s.Clock.NowUTC(),
		)
	_, err = s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}
	return nil
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
