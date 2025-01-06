package rolesgroups

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type Queries struct {
	Store *Store
}

func (q *Queries) GetRole(ctx context.Context, id string) (*model.Role, error) {
	role, err := q.Store.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return role.ToModel(), nil
}

func (q *Queries) GetGroup(ctx context.Context, id string) (*model.Group, error) {
	group, err := q.Store.GetGroupByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return group.ToModel(), nil
}

func (q *Queries) GetManyRoles(ctx context.Context, ids []string) ([]*model.Role, error) {
	roles, err := q.Store.GetManyRoles(ctx, ids)
	if err != nil {
		return nil, err
	}

	roleModels := make([]*model.Role, len(roles))
	for i, r := range roles {
		roleModels[i] = r.ToModel()
	}

	return roleModels, nil
}

func (q *Queries) GetManyGroups(ctx context.Context, ids []string) ([]*model.Group, error) {
	groups, err := q.Store.GetManyGroups(ctx, ids)
	if err != nil {
		return nil, err
	}

	groupModels := make([]*model.Group, len(groups))
	for i, r := range groups {
		groupModels[i] = r.ToModel()
	}

	return groupModels, nil
}

func (q *Queries) ListGroupsByRoleID(ctx context.Context, roleID string) ([]*model.Group, error) {
	groups, err := q.Store.ListGroupsByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	groupModels := make([]*model.Group, len(groups))
	for i, r := range groups {
		groupModels[i] = r.ToModel()
	}

	return groupModels, nil
}

func (q *Queries) ListRolesByGroupID(ctx context.Context, groupID string) ([]*model.Role, error) {
	roles, err := q.Store.ListRolesByGroupID(ctx, groupID)
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
	ExcludedIDs   []string
}

func (q *Queries) ListRoles(ctx context.Context, options *ListRolesOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	roles, offset, err := q.Store.ListRoles(ctx, options, pageArgs)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItemRef, len(roles))
	for i, r := range roles {
		//nolint:gosec // G115
		i_uint64 := uint64(i)
		pageKey := db.PageKey{Offset: offset + i_uint64}
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
	ExcludedIDs   []string
}

func (q *Queries) ListGroups(ctx context.Context, options *ListGroupsOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	groups, offset, err := q.Store.ListGroups(ctx, options, pageArgs)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItemRef, len(groups))
	for i, r := range groups {
		//nolint:gosec // G115
		i_uint64 := uint64(i)
		pageKey := db.PageKey{Offset: offset + i_uint64}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: r.ID, Cursor: cursor}
	}
	return models, nil
}

func (q *Queries) ListRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error) {
	roles, err := q.Store.ListRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleModels := make([]*model.Role, len(roles))
	for i, r := range roles {
		roleModels[i] = r.ToModel()
	}

	return roleModels, nil
}

func (q *Queries) ListRolesByUserIDs(ctx context.Context, userIDs []string) (map[string][]*model.Role, error) {
	rolesByUserID, err := q.Store.ListRolesByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	roleModelsByUserID := make(map[string][]*model.Role)
	for k, v := range rolesByUserID {
		for _, r := range v {
			roleModelsByUserID[k] = append(roleModelsByUserID[k], r.ToModel())
		}
	}

	return roleModelsByUserID, nil
}

func (q *Queries) ListGroupsByUserID(ctx context.Context, userID string) ([]*model.Group, error) {
	groups, err := q.Store.ListGroupsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	groupModels := make([]*model.Group, len(groups))
	for i, r := range groups {
		groupModels[i] = r.ToModel()
	}

	return groupModels, nil
}

func (q *Queries) ListGroupsByUserIDs(ctx context.Context, userIDs []string) (map[string][]*model.Group, error) {
	groupsByUserID, err := q.Store.ListGroupsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	groupModelsByUserID := make(map[string][]*model.Group)
	for k, v := range groupsByUserID {
		for _, g := range v {
			groupModelsByUserID[k] = append(groupModelsByUserID[k], g.ToModel())
		}
	}

	return groupModelsByUserID, nil
}

func (q *Queries) ListUserIDsByRoleID(ctx context.Context, roleID string, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	userIDs, offset, err := q.Store.ListUserIDsByRoleID(ctx, roleID, pageArgs)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItemRef, len(userIDs))
	for i, r := range userIDs {
		//nolint:gosec // G115
		i_uint64 := uint64(i)
		pageKey := db.PageKey{Offset: offset + i_uint64}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: r, Cursor: cursor}
	}
	return models, nil
}

func (q *Queries) ListAllUserIDsByRoleIDs(ctx context.Context, roleIDs []string) ([]string, error) {
	return q.Store.ListAllUserIDsByRoleID(ctx, roleIDs)
}

func (q *Queries) ListAllUserIDsByGroupKeys(ctx context.Context, groupKeys []string) ([]string, error) {
	groups, err := q.Store.GetManyGroupsByKeys(ctx, groupKeys)
	if err != nil {
		return nil, err
	}
	groupIDs := slice.Map(groups, func(group *Group) string { return group.ID })
	return q.Store.ListAllUserIDsByGroupIDs(ctx, groupIDs)
}

func (q *Queries) ListAllUserIDsByGroupIDs(ctx context.Context, groupIDs []string) ([]string, error) {
	return q.Store.ListAllUserIDsByGroupIDs(ctx, groupIDs)
}

func (q *Queries) ListUserIDsByGroupID(ctx context.Context, groupID string, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	userIDs, offset, err := q.Store.ListUserIDsByGroupID(ctx, groupID, pageArgs)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItemRef, len(userIDs))
	for i, r := range userIDs {
		//nolint:gosec // G115
		i_uint64 := uint64(i)
		pageKey := db.PageKey{Offset: offset + i_uint64}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: r, Cursor: cursor}
	}
	return models, nil
}

func (q *Queries) ListEffectiveRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error) {
	roles, err := q.Store.ListEffectiveRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleModels := make([]*model.Role, len(roles))
	for i, r := range roles {
		roleModels[i] = r.ToModel()
	}

	return roleModels, nil
}

func (q *Queries) ListAllUserIDsByEffectiveRoleIDs(ctx context.Context, roleIDs []string) ([]string, error) {
	return q.Store.ListAllUserIDsByEffectiveRoleIDs(ctx, roleIDs)
}

func (f *Queries) ListAllRolesByKeys(ctx context.Context, keys []string) ([]*model.Role, error) {
	roles, err := f.Store.GetManyRolesByKeys(ctx, keys)
	if err != nil {
		return nil, err
	}
	return slice.Map(roles, func(r *Role) *model.Role { return r.ToModel() }), nil
}

func (f *Queries) ListAllGroupsByKeys(ctx context.Context, keys []string) ([]*model.Group, error) {
	groups, err := f.Store.GetManyGroupsByKeys(ctx, keys)
	if err != nil {
		return nil, err
	}
	return slice.Map(groups, func(g *Group) *model.Group { return g.ToModel() }), nil
}

func (q *Queries) CountRoles(ctx context.Context) (uint64, error) {
	return q.Store.CountRoles(ctx)
}

func (q *Queries) CountGroups(ctx context.Context) (uint64, error) {
	return q.Store.CountGroups(ctx)
}
