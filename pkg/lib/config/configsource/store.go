package configsource

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type Store struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *Store) selectConfigSourceQuery() db.SelectBuilder {
	return s.SQLBuilder.Global().
		Select(
			"id",
			"app_id",
			"created_at",
			"updated_at",
			"data",
			"plan_name",
		).
		From(s.SQLBuilder.TableName("_portal_config_source"))
}

func (s *Store) scanConfigSource(scn db.Scanner) (*DatabaseSource, error) {
	d := &DatabaseSource{}
	var dataByte []byte

	err := scn.Scan(
		&d.ID,
		&d.AppID,
		&d.CreatedAt,
		&d.UpdatedAt,
		&dataByte,
		&d.PlanName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAppNotFound
	} else if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(dataByte, &d.Data); err != nil {
		return nil, err
	}

	return d, nil
}

func (s *Store) GetAppIDByDomain(domain string) (string, error) {
	builder := s.SQLBuilder.Global().
		Select(
			"app_id",
		).
		From(s.SQLBuilder.TableName("_portal_domain")).
		Where("domain = ?", domain)

	scanner, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return "", err
	}

	var appID string
	if err = scanner.Scan(&appID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrAppNotFound
		}
		return "", err
	}

	return appID, nil
}

func (s *Store) GetDatabaseSourceByAppID(appID string) (*DatabaseSource, error) {
	builder := s.selectConfigSourceQuery()
	builder = builder.Where("app_id = ?", appID)

	scanner, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	dbs, err := s.scanConfigSource(scanner)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}

	return dbs, nil
}

func (s *Store) CreateDatabaseSource(dbs *DatabaseSource) error {
	data, err := json.Marshal(dbs.Data)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.Global().
		Insert(s.SQLBuilder.TableName("_portal_config_source")).
		Columns(
			"id",
			"app_id",
			"data",
			"plan_name",
			"created_at",
			"updated_at",
		).
		Values(
			dbs.ID,
			dbs.AppID,
			data,
			dbs.PlanName,
			dbs.CreatedAt,
			dbs.UpdatedAt,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateDatabaseSource(dbs *DatabaseSource) error {
	data, err := json.Marshal(dbs.Data)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.Global().
		Update(s.SQLBuilder.TableName("_portal_config_source")).
		Set("updated_at", dbs.UpdatedAt).
		Set("data", data).
		Set("plan_name", dbs.PlanName).
		Where("id = ?", dbs.ID)

	result, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAppNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Sprintf("config_source_db: want 1 row updated, got %v", rowsAffected))
	}

	return nil
}

// ListAll is introduced by the need of authgear internal elasticsearch reindex --all.
func (s *Store) ListAll() ([]*DatabaseSource, error) {
	builder := s.selectConfigSourceQuery()

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*DatabaseSource
	for rows.Next() {
		item, err := s.scanConfigSource(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *Store) ListByPlan(planName string) ([]*DatabaseSource, error) {
	builder := s.selectConfigSourceQuery().
		Where("plan_name = ?", planName)

	rows, err := s.SQLExecutor.QueryWith(builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*DatabaseSource
	for rows.Next() {
		item, err := s.scanConfigSource(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
