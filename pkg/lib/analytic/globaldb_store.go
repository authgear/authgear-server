package analytic

import (
	"context"
	"encoding/json"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type GlobalDBStore struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *GlobalDBStore) GetAppOwners(ctx context.Context, rangeFrom *time.Time, rangeTo *time.Time) ([]*AppCollaborator, error) {
	builder := s.SQLBuilder.
		Select(
			"app_id",
			"user_id",
		).
		From(s.SQLBuilder.TableName("_portal_app_collaborator"))

	if rangeFrom != nil {
		builder = builder.Where("created_at >= ?", rangeFrom)
	}
	if rangeTo != nil {
		builder = builder.Where("created_at < ?", rangeTo)
	}

	builder = builder.
		Where("role = ?", "owner")

	rows, err := s.SQLExecutor.QueryWith(ctx, builder)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*AppCollaborator
	for rows.Next() {
		r := &AppCollaborator{}
		err = rows.Scan(
			&r.AppID,
			&r.UserID,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (s *GlobalDBStore) GetAppIDs(ctx context.Context) (appIDs []string, err error) {
	builder := s.SQLBuilder.
		Select(
			"app_id",
		).
		From(s.SQLBuilder.TableName("_portal_config_source")).
		OrderBy("created_at ASC")

	rows, e := s.SQLExecutor.QueryWith(ctx, builder)
	if e != nil {
		err = e
		return
	}
	defer rows.Close()
	for rows.Next() {
		var appID string
		err = rows.Scan(
			&appID,
		)
		if err != nil {
			return
		}
		appIDs = append(appIDs, appID)
	}
	return
}

func (s *GlobalDBStore) GetCollaboratorCount(ctx context.Context, appID string) (int, error) {
	q := s.SQLBuilder.
		Select(
			"count(*)",
		).
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("app_id = ?", appID)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return 0, err
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *GlobalDBStore) GetAppConfigSource(ctx context.Context, appID string) (*AppConfigSource, error) {
	q := s.SQLBuilder.
		Select(
			"app_id",
			"data",
			"plan_name",
		).
		From(s.SQLBuilder.TableName("_portal_config_source")).
		Where("app_id = ?", appID)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	out := &AppConfigSource{}
	var dataBytes []byte

	err = row.Scan(
		&out.AppID,
		&dataBytes,
		&out.PlanName,
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(dataBytes, &out.Data)
	if err != nil {
		return nil, err
	}

	return out, nil
}
