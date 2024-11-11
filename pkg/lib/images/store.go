package images

import (
	"context"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
}

func (s *Store) Create(ctx context.Context, i *File) error {
	metadata, err := json.Marshal(i.Metadata)
	if err != nil {
		return err
	}

	builder := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_images_file")).
		Columns(
			"id",
			"size",
			"metadata",
			"created_at",
		).
		Values(
			i.ID,
			i.Size,
			metadata,
			i.CreatedAt,
		)

	_, err = s.SQLExecutor.ExecWith(ctx, builder)
	if err != nil {
		return err
	}

	return nil
}
