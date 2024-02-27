package rolesgroups

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type Commands struct {
	Store *Store
}

func (c *Commands) CreateRole(options *NewRoleOptions) (*model.Role, error) {
	err := ValidateKey(options.Key)
	if err != nil {
		return nil, err
	}

	role := c.Store.NewRole(options)
	err = c.Store.CreateRole(role)
	if err != nil {
		return nil, err
	}

	return role.ToModel(), nil
}

func (c *Commands) UpdateRole(options *UpdateRoleOptions) (*model.Role, error) {
	if options.RequireUpdate() {
		if options.NewKey != nil {
			err := ValidateKey(*options.NewKey)
			if err != nil {
				return nil, err
			}
		}

		err := c.Store.UpdateRole(options)
		if err != nil {
			return nil, err
		}
	}

	r, err := c.Store.GetRoleByID(options.ID)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) DeleteRole(id string) error {
	return c.Store.DeleteRole(id)
}

func (c *Commands) CreateGroup(options *NewGroupOptions) (*model.Group, error) {
	err := ValidateKey(options.Key)
	if err != nil {
		return nil, err
	}

	group := c.Store.NewGroup(options)
	err = c.Store.CreateGroup(group)
	if err != nil {
		return nil, err
	}

	return group.ToModel(), nil
}

func (c *Commands) UpdateGroup(options *UpdateGroupOptions) (*model.Group, error) {
	if options.RequireUpdate() {
		if options.NewKey != nil {
			err := ValidateKey(*options.NewKey)
			if err != nil {
				return nil, err
			}
		}

		err := c.Store.UpdateGroup(options)
		if err != nil {
			return nil, err
		}
	}

	r, err := c.Store.GetGroupByID(options.ID)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) DeleteGroup(id string) error {
	return c.Store.DeleteGroup(id)
}

func (c *Commands) AddRoleToGroups(options *AddRoleToGroupsOptions) (*model.Role, error) {
	r, err := c.Store.AddRoleToGroups(options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) RemoveRoleFromGroups(options *RemoveRoleFromGroupsOptions) (*model.Role, error) {
	r, err := c.Store.RemoveRoleFromGroups(options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) AddRoleToUsers(options *AddRoleToUsersOptions) (*model.Role, error) {
	r, err := c.Store.AddRoleToUsers(options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) RemoveRoleFromUsers(options *RemoveRoleFromUsersOptions) (*model.Role, error) {
	r, err := c.Store.RemoveRoleFromUsers(options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) AddGroupToUsers(options *AddGroupToUsersOptions) (*model.Group, error) {
	r, err := c.Store.AddGroupToUsers(options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}
