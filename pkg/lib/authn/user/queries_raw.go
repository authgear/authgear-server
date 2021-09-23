package user

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type RawQueries struct {
	Store store
}

func (p *RawQueries) GetRaw(id string) (*User, error) {
	return p.Store.Get(id)
}

func (p *RawQueries) GetManyRaw(ids []string) ([]*User, error) {
	return p.Store.GetByIDs(ids)
}

func (p *RawQueries) Count() (uint64, error) {
	return p.Store.Count()
}

func (p *RawQueries) QueryPage(sortOption SortOption, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	users, offset, err := p.Store.QueryPage(sortOption, pageArgs)
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
