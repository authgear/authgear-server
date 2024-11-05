package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type RolesGroupsCommands interface {
	CreateRole(ctx context.Context, options *rolesgroups.NewRoleOptions) (*model.Role, error)
	UpdateRole(ctx context.Context, options *rolesgroups.UpdateRoleOptions) (*model.Role, error)
	DeleteRole(ctx context.Context, id string) error

	CreateGroup(ctx context.Context, options *rolesgroups.NewGroupOptions) (*model.Group, error)
	UpdateGroup(ctx context.Context, options *rolesgroups.UpdateGroupOptions) (*model.Group, error)
	DeleteGroup(ctx context.Context, id string) error

	AddRoleToGroups(ctx context.Context, options *rolesgroups.AddRoleToGroupsOptions) (*model.Role, error)
	RemoveRoleFromGroups(ctx context.Context, options *rolesgroups.RemoveRoleFromGroupsOptions) (*model.Role, error)

	AddRoleToUsers(ctx context.Context, options *rolesgroups.AddRoleToUsersOptions) (*model.Role, error)
	RemoveRoleFromUsers(ctx context.Context, options *rolesgroups.RemoveRoleFromUsersOptions) (*model.Role, error)

	AddGroupToUsers(ctx context.Context, options *rolesgroups.AddGroupToUsersOptions) (*model.Group, error)
	RemoveGroupFromUsers(ctx context.Context, options *rolesgroups.RemoveGroupFromUsersOptions) (*model.Group, error)

	AddGroupToRoles(ctx context.Context, options *rolesgroups.AddGroupToRolesOptions) (*model.Group, error)
	RemoveGroupFromRoles(ctx context.Context, options *rolesgroups.RemoveGroupFromRolesOptions) (*model.Group, error)

	AddUserToRoles(ctx context.Context, options *rolesgroups.AddUserToRolesOptions) error
	RemoveUserFromRoles(ctx context.Context, options *rolesgroups.RemoveUserFromRolesOptions) error

	AddUserToGroups(ctx context.Context, options *rolesgroups.AddUserToGroupsOptions) error
	RemoveUserFromGroups(ctx context.Context, options *rolesgroups.RemoveUserFromGroupsOptions) error
}

type RolesGroupsQueries interface {
	GetRole(ctx context.Context, id string) (*model.Role, error)
	GetGroup(ctx context.Context, id string) (*model.Group, error)
	ListRoles(ctx context.Context, options *rolesgroups.ListRolesOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error)
	ListGroups(ctx context.Context, options *rolesgroups.ListGroupsOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error)
	ListGroupsByRoleID(ctx context.Context, roleID string) ([]*model.Group, error)
	ListRolesByGroupID(ctx context.Context, groupID string) ([]*model.Role, error)
	ListRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error)
	ListGroupsByUserID(ctx context.Context, userID string) ([]*model.Group, error)
	ListUserIDsByRoleID(ctx context.Context, roleID string, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error)
	ListUserIDsByGroupID(ctx context.Context, groupID string, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error)
	ListEffectiveRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error)
	ListAllUserIDsByGroupIDs(ctx context.Context, groupIDs []string) ([]string, error)
	ListAllUserIDsByGroupKeys(ctx context.Context, groupKeys []string) ([]string, error)
	ListAllUserIDsByRoleIDs(ctx context.Context, roleIDs []string) ([]string, error)
	ListAllUserIDsByEffectiveRoleIDs(ctx context.Context, roleIDs []string) ([]string, error)
	ListAllRolesByKeys(ctx context.Context, keys []string) ([]*model.Role, error)
	ListAllGroupsByKeys(ctx context.Context, keys []string) ([]*model.Group, error)
	CountRoles(ctx context.Context) (uint64, error)
	CountGroups(ctx context.Context) (uint64, error)
}

type RolesGroupsFacade struct {
	RolesGroupsCommands RolesGroupsCommands
	RolesGroupsQueries  RolesGroupsQueries
}

func (f *RolesGroupsFacade) CreateRole(ctx context.Context, options *rolesgroups.NewRoleOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.CreateRole(ctx, options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) UpdateRole(ctx context.Context, options *rolesgroups.UpdateRoleOptions) (err error) {
	_, err = f.RolesGroupsCommands.UpdateRole(ctx, options)
	if err != nil {
		return
	}

	return
}

func (f *RolesGroupsFacade) DeleteRole(ctx context.Context, id string) (err error) {
	return f.RolesGroupsCommands.DeleteRole(ctx, id)
}

func (f *RolesGroupsFacade) ListRoles(ctx context.Context, options *rolesgroups.ListRolesOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	refs, err := f.RolesGroupsQueries.ListRoles(ctx, options, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	count, err := f.RolesGroupsQueries.CountRoles(ctx)
	if err != nil {
		return nil, nil, err
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		return count, nil
	})), nil
}

func (f *RolesGroupsFacade) ListGroupsByRoleID(ctx context.Context, roleID string) ([]*model.Group, error) {
	return f.RolesGroupsQueries.ListGroupsByRoleID(ctx, roleID)
}

func (f *RolesGroupsFacade) CreateGroup(ctx context.Context, options *rolesgroups.NewGroupOptions) (groupID string, err error) {
	g, err := f.RolesGroupsCommands.CreateGroup(ctx, options)
	if err != nil {
		return
	}

	groupID = g.ID
	return
}

func (f *RolesGroupsFacade) UpdateGroup(ctx context.Context, options *rolesgroups.UpdateGroupOptions) (err error) {
	_, err = f.RolesGroupsCommands.UpdateGroup(ctx, options)
	if err != nil {
		return
	}

	return
}

func (f *RolesGroupsFacade) DeleteGroup(ctx context.Context, id string) (err error) {
	return f.RolesGroupsCommands.DeleteGroup(ctx, id)
}

func (f *RolesGroupsFacade) ListGroups(ctx context.Context, options *rolesgroups.ListGroupsOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	refs, err := f.RolesGroupsQueries.ListGroups(ctx, options, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	count, err := f.RolesGroupsQueries.CountGroups(ctx)
	if err != nil {
		return nil, nil, err
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		return count, nil
	})), nil
}

func (f *RolesGroupsFacade) ListRolesByGroupID(ctx context.Context, groupID string) ([]*model.Role, error) {
	return f.RolesGroupsQueries.ListRolesByGroupID(ctx, groupID)
}

func (f *RolesGroupsFacade) AddRoleToGroups(ctx context.Context, options *rolesgroups.AddRoleToGroupsOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.AddRoleToGroups(ctx, options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) RemoveRoleFromGroups(ctx context.Context, options *rolesgroups.RemoveRoleFromGroupsOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.RemoveRoleFromGroups(ctx, options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) AddRoleToUsers(ctx context.Context, options *rolesgroups.AddRoleToUsersOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.AddRoleToUsers(ctx, options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) RemoveRoleFromUsers(ctx context.Context, options *rolesgroups.RemoveRoleFromUsersOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.RemoveRoleFromUsers(ctx, options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) AddGroupToUsers(ctx context.Context, options *rolesgroups.AddGroupToUsersOptions) (groupID string, err error) {
	r, err := f.RolesGroupsCommands.AddGroupToUsers(ctx, options)
	if err != nil {
		return
	}

	groupID = r.ID
	return
}

func (f *RolesGroupsFacade) RemoveGroupFromUsers(ctx context.Context, options *rolesgroups.RemoveGroupFromUsersOptions) (groupID string, err error) {
	r, err := f.RolesGroupsCommands.RemoveGroupFromUsers(ctx, options)
	if err != nil {
		return
	}

	groupID = r.ID
	return
}

func (f *RolesGroupsFacade) AddGroupToRoles(ctx context.Context, options *rolesgroups.AddGroupToRolesOptions) (groupID string, err error) {
	r, err := f.RolesGroupsCommands.AddGroupToRoles(ctx, options)
	if err != nil {
		return
	}

	groupID = r.ID
	return
}

func (f *RolesGroupsFacade) RemoveGroupFromRoles(ctx context.Context, options *rolesgroups.RemoveGroupFromRolesOptions) (groupID string, err error) {
	r, err := f.RolesGroupsCommands.RemoveGroupFromRoles(ctx, options)
	if err != nil {
		return
	}

	groupID = r.ID
	return
}

func (f *RolesGroupsFacade) AddUserToRoles(ctx context.Context, options *rolesgroups.AddUserToRolesOptions) (err error) {
	err = f.RolesGroupsCommands.AddUserToRoles(ctx, options)
	if err != nil {
		return
	}

	return
}

func (f *RolesGroupsFacade) RemoveUserFromRoles(ctx context.Context, options *rolesgroups.RemoveUserFromRolesOptions) (err error) {
	err = f.RolesGroupsCommands.RemoveUserFromRoles(ctx, options)
	if err != nil {
		return
	}

	return
}

func (f *RolesGroupsFacade) AddUserToGroups(ctx context.Context, options *rolesgroups.AddUserToGroupsOptions) (err error) {
	err = f.RolesGroupsCommands.AddUserToGroups(ctx, options)
	if err != nil {
		return
	}

	return
}

func (f *RolesGroupsFacade) RemoveUserFromGroups(ctx context.Context, options *rolesgroups.RemoveUserFromGroupsOptions) (err error) {
	err = f.RolesGroupsCommands.RemoveUserFromGroups(ctx, options)
	if err != nil {
		return
	}

	return
}

func (f *RolesGroupsFacade) ListRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error) {
	return f.RolesGroupsQueries.ListRolesByUserID(ctx, userID)
}

func (f *RolesGroupsFacade) ListGroupsByUserID(ctx context.Context, userID string) ([]*model.Group, error) {
	return f.RolesGroupsQueries.ListGroupsByUserID(ctx, userID)
}

func (f *RolesGroupsFacade) ListUserIDsByRoleID(ctx context.Context, roleID string, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	refs, err := f.RolesGroupsQueries.ListUserIDsByRoleID(ctx, roleID, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		// No need to report the total number of groups. So we return nil here.
		return nil, nil
	})), nil
}

func (f *RolesGroupsFacade) ListAllUserIDsByGroupIDs(ctx context.Context, groupIDs []string) ([]string, error) {
	return f.RolesGroupsQueries.ListAllUserIDsByGroupIDs(ctx, groupIDs)
}

func (f *RolesGroupsFacade) ListAllUserIDsByGroupKeys(ctx context.Context, groupKeys []string) ([]string, error) {
	return f.RolesGroupsQueries.ListAllUserIDsByGroupKeys(ctx, groupKeys)
}

func (f *RolesGroupsFacade) ListUserIDsByGroupID(ctx context.Context, groupID string, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	refs, err := f.RolesGroupsQueries.ListUserIDsByGroupID(ctx, groupID, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		// No need to report the total number of groups. So we return nil here.
		return nil, nil
	})), nil
}

func (f *RolesGroupsFacade) ListEffectiveRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error) {
	return f.RolesGroupsQueries.ListEffectiveRolesByUserID(ctx, userID)
}

func (f *RolesGroupsFacade) ListAllUserIDsByEffectiveRoleIDs(ctx context.Context, roleIDs []string) ([]string, error) {
	return f.RolesGroupsQueries.ListAllUserIDsByEffectiveRoleIDs(ctx, roleIDs)
}

func (f *RolesGroupsFacade) ListAllUserIDsByRoleIDs(ctx context.Context, roleIDs []string) ([]string, error) {
	return f.RolesGroupsQueries.ListAllUserIDsByRoleIDs(ctx, roleIDs)
}

func (f *RolesGroupsFacade) ListAllRolesByKeys(ctx context.Context, keys []string) ([]*model.Role, error) {
	return f.RolesGroupsQueries.ListAllRolesByKeys(ctx, keys)
}

func (f *RolesGroupsFacade) ListAllGroupsByKeys(ctx context.Context, keys []string) ([]*model.Group, error) {
	return f.RolesGroupsQueries.ListAllGroupsByKeys(ctx, keys)
}

func (f *RolesGroupsFacade) GetRole(ctx context.Context, roleID string) (*model.Role, error) {
	return f.RolesGroupsQueries.GetRole(ctx, roleID)
}

func (f *RolesGroupsFacade) GetGroup(ctx context.Context, groupID string) (*model.Group, error) {
	return f.RolesGroupsQueries.GetGroup(ctx, groupID)
}
