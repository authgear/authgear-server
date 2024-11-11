package audit

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Query struct {
	Database *auditdb.ReadHandle
	Store    *ReadStore
}

func (q *Query) GetByIDs(ctx context.Context, ids []string) ([]*Log, error) {
	if q.Database == nil {
		return make([]*Log, len(ids)), nil
	}

	return q.Store.GetByIDs(ctx, ids)
}

func (q *Query) Count(ctx context.Context, opts QueryPageOptions) (uint64, error) {
	if q.Database == nil {
		return 0, nil
	}

	return q.Store.Count(ctx, opts)
}

func (q *Query) QueryPage(ctx context.Context, opts QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	if q.Database == nil {
		return nil, nil
	}

	logs, offset, err := q.Store.QueryPage(ctx, opts, pageArgs)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItemRef, len(logs))
	for i, l := range logs {
		pageKey := db.PageKey{Offset: offset + uint64(i)}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: l.ID, Cursor: cursor}
	}

	return models, nil
}
