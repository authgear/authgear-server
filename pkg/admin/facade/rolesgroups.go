package facade

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
)

type RolesGroupsCommands interface {
	CreateRole(options *rolesgroups.NewRoleOptions) (*model.Role, error)
	UpdateRole(options *rolesgroups.UpdateRoleOptions) (*model.Role, error)
}

type RolesGroupsFacade struct {
	RolesGroupsCommands RolesGroupsCommands
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
