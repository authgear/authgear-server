package rolesgroups

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type Commands struct {
	Store *Store
}

func (c *Commands) CreateRole(ctx context.Context, options *NewRoleOptions) (*model.Role, error) {
	err := ValidateKey(options.Key)
	if err != nil {
		return nil, err
	}

	role := c.Store.NewRole(options)
	err = c.Store.CreateRole(ctx, role)
	if err != nil {
		return nil, err
	}

	return role.ToModel(), nil
}

func (c *Commands) UpdateRole(ctx context.Context, options *UpdateRoleOptions) (*model.Role, error) {
	if options.RequireUpdate() {
		if options.NewKey != nil {
			err := ValidateKey(*options.NewKey)
			if err != nil {
				return nil, err
			}
		}

		err := c.Store.UpdateRole(ctx, options)
		if err != nil {
			return nil, err
		}
	}

	r, err := c.Store.GetRoleByID(ctx, options.ID)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) DeleteRole(ctx context.Context, id string) error {
	return c.Store.DeleteRole(ctx, id)
}

func (c *Commands) CreateGroup(ctx context.Context, options *NewGroupOptions) (*model.Group, error) {
	err := ValidateKey(options.Key)
	if err != nil {
		return nil, err
	}

	group := c.Store.NewGroup(options)
	err = c.Store.CreateGroup(ctx, group)
	if err != nil {
		return nil, err
	}

	return group.ToModel(), nil
}

func (c *Commands) UpdateGroup(ctx context.Context, options *UpdateGroupOptions) (*model.Group, error) {
	if options.RequireUpdate() {
		if options.NewKey != nil {
			err := ValidateKey(*options.NewKey)
			if err != nil {
				return nil, err
			}
		}

		err := c.Store.UpdateGroup(ctx, options)
		if err != nil {
			return nil, err
		}
	}

	r, err := c.Store.GetGroupByID(ctx, options.ID)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) DeleteGroup(ctx context.Context, id string) error {
	return c.Store.DeleteGroup(ctx, id)
}

func (c *Commands) AddRoleToGroups(ctx context.Context, options *AddRoleToGroupsOptions) (*model.Role, error) {
	r, err := c.Store.AddRoleToGroups(ctx, options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) RemoveRoleFromGroups(ctx context.Context, options *RemoveRoleFromGroupsOptions) (*model.Role, error) {
	r, err := c.Store.RemoveRoleFromGroups(ctx, options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) AddRoleToUsers(ctx context.Context, options *AddRoleToUsersOptions) (*model.Role, error) {
	r, err := c.Store.AddRoleToUsers(ctx, options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) RemoveRoleFromUsers(ctx context.Context, options *RemoveRoleFromUsersOptions) (*model.Role, error) {
	r, err := c.Store.RemoveRoleFromUsers(ctx, options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) AddGroupToUsers(ctx context.Context, options *AddGroupToUsersOptions) (*model.Group, error) {
	r, err := c.Store.AddGroupToUsers(ctx, options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) RemoveGroupFromUsers(ctx context.Context, options *RemoveGroupFromUsersOptions) (*model.Group, error) {
	r, err := c.Store.RemoveGroupFromUsers(ctx, options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) AddGroupToRoles(ctx context.Context, options *AddGroupToRolesOptions) (*model.Group, error) {
	r, err := c.Store.AddGroupToRoles(ctx, options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) RemoveGroupFromRoles(ctx context.Context, options *RemoveGroupFromRolesOptions) (*model.Group, error) {
	r, err := c.Store.RemoveGroupFromRoles(ctx, options)
	if err != nil {
		return nil, err
	}

	return r.ToModel(), nil
}

func (c *Commands) AddUserToRoles(ctx context.Context, options *AddUserToRolesOptions) error {
	err := c.Store.AddUserToRoles(ctx, options)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) RemoveUserFromRoles(ctx context.Context, options *RemoveUserFromRolesOptions) error {
	err := c.Store.RemoveUserFromRoles(ctx, options)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) AddUserToGroups(ctx context.Context, options *AddUserToGroupsOptions) error {
	err := c.Store.AddUserToGroups(ctx, options)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) RemoveUserFromGroups(ctx context.Context, options *RemoveUserFromGroupsOptions) error {
	err := c.Store.RemoveUserFromGroups(ctx, options)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) DeleteUserGroup(ctx context.Context, userID string) error {
	err := c.Store.DeleteUserGroup(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) DeleteUserRole(ctx context.Context, userID string) error {
	err := c.Store.DeleteUserRole(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) ResetUserGroup(ctx context.Context, options *ResetUserGroupOptions) error {
	err := c.Store.ResetUserGroup(ctx, options)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) ResetUserRole(ctx context.Context, options *ResetUserRoleOptions) error {
	err := c.Store.ResetUserRole(ctx, options)
	if err != nil {
		return err
	}

	return nil
}
