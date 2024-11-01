package user

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type RawQueries struct {
	Store store
}

func (p *RawQueries) GetRaw(ctx context.Context, id string) (*User, error) {
	return p.Store.Get(ctx, id)
}

func (p *RawQueries) GetManyRaw(ctx context.Context, ids []string) ([]*User, error) {
	return p.Store.GetByIDs(ctx, ids)
}

func (p *RawQueries) Count(ctx context.Context) (uint64, error) {
	return p.Store.Count(ctx)
}

func (p *RawQueries) QueryPage(ctx context.Context, listOption ListOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	users, offset, err := p.Store.QueryPage(ctx, listOption, pageArgs)
	if err != nil {
		return nil, err
	}

	var models = make([]model.PageItemRef, len(users))
	for i, u := range users {
		pageKey := db.PageKey{Offset: offset + uint64(i)}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: u.ID, Cursor: cursor}
	}
	return models, nil
}
