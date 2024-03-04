package rolesgroups

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Queries struct {
	Store *Store
}

func (q *Queries) GetManyRoles(ids []string) ([]*model.Role, error) {
	roles, err := q.Store.GetManyRoles(ids)
	if err != nil {
		return nil, err
	}

	roleModels := make([]*model.Role, len(roles))
	for i, r := range roles {
		roleModels[i] = r.ToModel()
	}

	return roleModels, nil
}

func (q *Queries) GetManyGroups(ids []string) ([]*model.Group, error) {
	groups, err := q.Store.GetManyGroups(ids)
	if err != nil {
		return nil, err
	}

	groupModels := make([]*model.Group, len(groups))
	for i, r := range groups {
		groupModels[i] = r.ToModel()
	}

	return groupModels, nil
}

func (q *Queries) ListGroupsByRoleID(roleID string) ([]*model.Group, error) {
	groups, err := q.Store.ListGroupsByRoleID(roleID)
	if err != nil {
		return nil, err
	}

	groupModels := make([]*model.Group, len(groups))
	for i, r := range groups {
		groupModels[i] = r.ToModel()
	}

	return groupModels, nil
}

func (q *Queries) ListRolesByGroupID(groupID string) ([]*model.Role, error) {
	roles, err := q.Store.ListRolesByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	roleModels := make([]*model.Role, len(roles))
	for i, r := range roles {
		roleModels[i] = r.ToModel()
	}

	return roleModels, nil
}

type ListRolesOptions struct {
	SearchKeyword string
}

func (q *Queries) ListRoles(options *ListRolesOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	roles, offset, err := q.Store.ListRoles(options, pageArgs)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItemRef, len(roles))
	for i, r := range roles {
		pageKey := db.PageKey{Offset: offset + uint64(i)}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: r.ID, Cursor: cursor}
	}
	return models, nil
}

type ListGroupsOptions struct {
	SearchKeyword string
}

func (q *Queries) ListGroups(options *ListGroupsOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	groups, offset, err := q.Store.ListGroups(options, pageArgs)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItemRef, len(groups))
	for i, r := range groups {
		pageKey := db.PageKey{Offset: offset + uint64(i)}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: r.ID, Cursor: cursor}
	}
	return models, nil
}

func (q *Queries) ListRolesByUserID(userID string) ([]*model.Role, error) {
	roles, err := q.Store.ListRolesByUserID(userID)
	if err != nil {
		return nil, err
	}

	roleModels := make([]*model.Role, len(roles))
	for i, r := range roles {
		roleModels[i] = r.ToModel()
	}

	return roleModels, nil
}

func (q *Queries) ListGroupsByUserID(userID string) ([]*model.Group, error) {
	groups, err := q.Store.ListGroupsByUserID(userID)
	if err != nil {
		return nil, err
	}

	groupModels := make([]*model.Group, len(groups))
	for i, r := range groups {
		groupModels[i] = r.ToModel()
	}

	return groupModels, nil
}

func (q *Queries) ListUserIDsByRoleID(roleID string, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	userIDs, offset, err := q.Store.ListUserIDsByRoleID(roleID, pageArgs)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItemRef, len(userIDs))
	for i, r := range userIDs {
		pageKey := db.PageKey{Offset: offset + uint64(i)}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: r, Cursor: cursor}
	}
	return models, nil
}

func (q *Queries) ListUserIDsByGroupID(groupID string, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	userIDs, offset, err := q.Store.ListUserIDsByGroupID(groupID, pageArgs)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItemRef, len(userIDs))
	for i, r := range userIDs {
		pageKey := db.PageKey{Offset: offset + uint64(i)}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: r, Cursor: cursor}
	}
	return models, nil
}

func (q *Queries) ListEffectiveRolesByUserID(userID string) ([]*model.Role, error) {
	roles, err := q.Store.ListEffectiveRolesByUserID(userID)
	if err != nil {
		return nil, err
	}

	roleModels := make([]*model.Role, len(roles))
	for i, r := range roles {
		roleModels[i] = r.ToModel()
	}

	return roleModels, nil
}
