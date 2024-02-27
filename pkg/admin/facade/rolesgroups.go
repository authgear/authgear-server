package facade

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type RolesGroupsCommands interface {
	CreateRole(options *rolesgroups.NewRoleOptions) (*model.Role, error)
	UpdateRole(options *rolesgroups.UpdateRoleOptions) (*model.Role, error)
	DeleteRole(id string) error

	CreateGroup(options *rolesgroups.NewGroupOptions) (*model.Group, error)
	UpdateGroup(options *rolesgroups.UpdateGroupOptions) (*model.Group, error)
	DeleteGroup(id string) error

	AddRoleToGroups(options *rolesgroups.AddRoleToGroupsOptions) (*model.Role, error)
	RemoveRoleFromGroups(options *rolesgroups.RemoveRoleFromGroupsOptions) (*model.Role, error)

	AddRoleToUsers(options *rolesgroups.AddRoleToUsersOptions) (*model.Role, error)
	RemoveRoleFromUsers(options *rolesgroups.RemoveRoleFromUsersOptions) (*model.Role, error)

	AddGroupToUsers(options *rolesgroups.AddGroupToUsersOptions) (*model.Group, error)
	RemoveGroupFromUsers(options *rolesgroups.RemoveGroupFromUsersOptions) (*model.Group, error)

	AddGroupToRoles(options *rolesgroups.AddGroupToRolesOptions) (*model.Group, error)
	RemoveGroupFromRoles(options *rolesgroups.RemoveGroupFromRolesOptions) (*model.Group, error)

	AddUserToRoles(options *rolesgroups.AddUserToRolesOptions) error
}

type RolesGroupsQueries interface {
	ListRoles(options *rolesgroups.ListRolesOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error)
	ListGroups(options *rolesgroups.ListGroupsOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error)
	ListGroupsByRoleID(roleID string) ([]*model.Group, error)
	ListRolesByGroupID(groupID string) ([]*model.Role, error)
}

type RolesGroupsFacade struct {
	RolesGroupsCommands RolesGroupsCommands
	RolesGroupsQueries  RolesGroupsQueries
}

func (f *RolesGroupsFacade) CreateRole(options *rolesgroups.NewRoleOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.CreateRole(options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) UpdateRole(options *rolesgroups.UpdateRoleOptions) (err error) {
	_, err = f.RolesGroupsCommands.UpdateRole(options)
	if err != nil {
		return
	}

	return
}

func (f *RolesGroupsFacade) DeleteRole(id string) (err error) {
	return f.RolesGroupsCommands.DeleteRole(id)
}

func (f *RolesGroupsFacade) ListRoles(options *rolesgroups.ListRolesOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	refs, err := f.RolesGroupsQueries.ListRoles(options, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		// No need to report the total number of roles. So we return nil here.
		return nil, nil
	})), nil
}

func (f *RolesGroupsFacade) ListGroupsByRoleID(roleID string) ([]*model.Group, error) {
	return f.RolesGroupsQueries.ListGroupsByRoleID(roleID)
}

func (f *RolesGroupsFacade) CreateGroup(options *rolesgroups.NewGroupOptions) (groupID string, err error) {
	g, err := f.RolesGroupsCommands.CreateGroup(options)
	if err != nil {
		return
	}

	groupID = g.ID
	return
}

func (f *RolesGroupsFacade) UpdateGroup(options *rolesgroups.UpdateGroupOptions) (err error) {
	_, err = f.RolesGroupsCommands.UpdateGroup(options)
	if err != nil {
		return
	}

	return
}

func (f *RolesGroupsFacade) DeleteGroup(id string) (err error) {
	return f.RolesGroupsCommands.DeleteGroup(id)
}

func (f *RolesGroupsFacade) ListGroups(options *rolesgroups.ListGroupsOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	refs, err := f.RolesGroupsQueries.ListGroups(options, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		// No need to report the total number of groups. So we return nil here.
		return nil, nil
	})), nil
}

func (f *RolesGroupsFacade) ListRolesByGroupID(groupID string) ([]*model.Role, error) {
	return f.RolesGroupsQueries.ListRolesByGroupID(groupID)
}

func (f *RolesGroupsFacade) AddRoleToGroups(options *rolesgroups.AddRoleToGroupsOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.AddRoleToGroups(options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) RemoveRoleFromGroups(options *rolesgroups.RemoveRoleFromGroupsOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.RemoveRoleFromGroups(options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) AddRoleToUsers(options *rolesgroups.AddRoleToUsersOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.AddRoleToUsers(options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) RemoveRoleFromUsers(options *rolesgroups.RemoveRoleFromUsersOptions) (roleID string, err error) {
	r, err := f.RolesGroupsCommands.RemoveRoleFromUsers(options)
	if err != nil {
		return
	}

	roleID = r.ID
	return
}

func (f *RolesGroupsFacade) AddGroupToUsers(options *rolesgroups.AddGroupToUsersOptions) (groupID string, err error) {
	r, err := f.RolesGroupsCommands.AddGroupToUsers(options)
	if err != nil {
		return
	}

	groupID = r.ID
	return
}

func (f *RolesGroupsFacade) RemoveGroupFromUsers(options *rolesgroups.RemoveGroupFromUsersOptions) (groupID string, err error) {
	r, err := f.RolesGroupsCommands.RemoveGroupFromUsers(options)
	if err != nil {
		return
	}

	groupID = r.ID
	return
}

func (f *RolesGroupsFacade) AddGroupToRoles(options *rolesgroups.AddGroupToRolesOptions) (groupID string, err error) {
	r, err := f.RolesGroupsCommands.AddGroupToRoles(options)
	if err != nil {
		return
	}

	groupID = r.ID
	return
}

func (f *RolesGroupsFacade) RemoveGroupFromRoles(options *rolesgroups.RemoveGroupFromRolesOptions) (groupID string, err error) {
	r, err := f.RolesGroupsCommands.RemoveGroupFromRoles(options)
	if err != nil {
		return
	}

	groupID = r.ID
	return
}

func (f *RolesGroupsFacade) AddUserToRoles(options *rolesgroups.AddUserToRolesOptions) (err error) {
	err = f.RolesGroupsCommands.AddUserToRoles(options)
	if err != nil {
		return
	}

	return
}
