package tutorial

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type StoreImpl struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *StoreImpl) Get(ctx context.Context, appID string) (*Entry, error) {
	builder := s.SQLBuilder.
		Select(
			"data",
		).
		From(s.SQLBuilder.TableName("_portal_tutorial_progress")).
		Where("app_id = ?", appID)

	scanner, err := s.SQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return nil, err
	}

	entry := NewEntry(appID)
	var dataBytes []byte
	err = scanner.Scan(
		&dataBytes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return entry, nil
	}
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(dataBytes, &entry.Data)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *StoreImpl) Save(ctx context.Context, entry *Entry) error {
	data, err := json.Marshal(entry.Data)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_portal_tutorial_progress")).
		Columns(
			"app_id",
			"data",
		).
		Values(
			entry.AppID,
			data,
		).
		Suffix("ON CONFLICT (app_id) DO UPDATE SET data = excluded.data")

	_, err = s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}
