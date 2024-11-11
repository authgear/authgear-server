package configsource

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type Store struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *Store) selectConfigSourceQuery() sq.SelectBuilder {
	return s.SQLBuilder.
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

func (s *Store) GetAppIDByDomain(ctx context.Context, domain string) (string, error) {
	builder := s.SQLBuilder.
		Select(
			"app_id",
		).
		From(s.SQLBuilder.TableName("_portal_domain")).
		Where("domain = ?", domain)

	scanner, err := s.SQLExecutor.QueryRowWith(ctx, builder)
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

func (s *Store) GetDomainsByAppID(ctx context.Context, appID string) (domains []string, err error) {
	builder := s.SQLBuilder.
		Select(
			"domain",
		).
		From(s.SQLBuilder.TableName("_portal_domain")).
		Where("app_id = ?", appID)

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var d string
		err := rows.Scan(&d)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}

	return domains, nil
}

func (s *Store) GetDatabaseSourceByAppID(ctx context.Context, appID string) (*DatabaseSource, error) {
	builder := s.selectConfigSourceQuery()
	builder = builder.Where("app_id = ?", appID)

	scanner, err := s.SQLExecutor.QueryRowWith(ctx, builder)
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

func (s *Store) CreateDatabaseSource(ctx context.Context, dbs *DatabaseSource) error {
	data, err := json.Marshal(dbs.Data)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.
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

	_, err = s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateDatabaseSource(ctx context.Context, dbs *DatabaseSource) error {
	data, err := json.Marshal(dbs.Data)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.
		Update(s.SQLBuilder.TableName("_portal_config_source")).
		Set("updated_at", dbs.UpdatedAt).
		Set("data", data).
		Set("plan_name", dbs.PlanName).
		Where("id = ?", dbs.ID)

	result, err := s.SQLExecutor.ExecWith(ctx, q)
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
func (s *Store) ListAll(ctx context.Context) ([]*DatabaseSource, error) {
	builder := s.selectConfigSourceQuery()

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
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

func (s *Store) ListByPlan(ctx context.Context, planName string) ([]*DatabaseSource, error) {
	builder := s.selectConfigSourceQuery().
		Where("plan_name = ?", planName)

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
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
