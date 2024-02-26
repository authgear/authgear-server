package facade

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
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
}

type RolesGroupsQueries interface {
	ListGroupsByRoleID(roleID string) ([]*model.Group, error)
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
