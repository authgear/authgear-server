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
