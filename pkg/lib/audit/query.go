package audit

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Query struct {
	Store *Store
}

func (q *Query) GetByIDs(ids []string) ([]*Log, error) {
	return q.Store.GetByIDs(ids)
}

func (q *Query) Count(opts QueryPageOptions) (uint64, error) {
	return q.Store.Count(opts)
}

func (q *Query) QueryPage(opts QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	logs, offset, err := q.Store.QueryPage(opts, pageArgs)
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
