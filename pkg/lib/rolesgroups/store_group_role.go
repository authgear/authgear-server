package rolesgroups

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func (s *Store) ListGroupsByRoleID(ctx context.Context, roleID string) ([]*Group, error) {
	q := s.SQLBuilder.
		Select(
			"g.id",
			"g.created_at",
			"g.updated_at",
			"g.key",
			"g.name",
			"g.description",
		).
		From(s.SQLBuilder.TableName("_auth_group_role"), "gr").
		Join(s.SQLBuilder.TableName("_auth_group"), "g", "gr.group_id = g.id").
		Where("gr.role_id = ?", roleID)

	return s.queryGroups(ctx, q)
}

func (s *Store) ListRolesByGroupID(ctx context.Context, groupID string) ([]*Role, error) {
	q := s.SQLBuilder.
		Select(
			"r.id",
			"r.created_at",
			"r.updated_at",
			"r.key",
			"r.name",
			"r.description",
		).
		From(s.SQLBuilder.TableName("_auth_group_role"), "gr").
		Join(s.SQLBuilder.TableName("_auth_role"), "r", "gr.role_id = r.id").
		Where("gr.group_id = ?", groupID)

	return s.queryRoles(ctx, q)
}

type AddRoleToGroupsOptions struct {
	RoleKey   string
	GroupKeys []string
}

func (s *Store) AddRoleToGroups(ctx context.Context, options *AddRoleToGroupsOptions) (*Role, error) {
	r, err := s.GetRoleByKey(ctx, options.RoleKey)
	if err != nil {
		return nil, err
	}

	gs, err := s.GetManyGroupsByKeys(ctx, options.GroupKeys)
	if err != nil {
		return nil, err
	}

	var seenKeys []string
	now := s.Clock.NowUTC()
	for _, g := range gs {
		id := uuid.New()
		q := s.SQLBuilder.
			Insert(s.SQLBuilder.TableName("_auth_group_role")).
			Columns(
				"id",
				"created_at",
				"updated_at",
				"group_id",
				"role_id",
			).
			Values(
				id,
				now,
				now,
				g.ID,
				r.ID,
			).Suffix("ON CONFLICT DO NOTHING")

		_, err := s.SQLExecutor.ExecWith(ctx, q)
		if err != nil {
			return nil, err
		}

		seenKeys = append(seenKeys, g.Key)
	}

	missingKeys := slice.ExceptStrings(options.GroupKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := GroupUnknownKeys.NewWithInfo("unknown group keys", apierrors.Details{"keys": missingKeys})
		return nil, err
	}

	return r, nil
}

type RemoveRoleFromGroupsOptions struct {
	RoleKey   string
	GroupKeys []string
}

func (s *Store) RemoveRoleFromGroups(ctx context.Context, options *RemoveRoleFromGroupsOptions) (*Role, error) {
	r, err := s.GetRoleByKey(ctx, options.RoleKey)
	if err != nil {
		return nil, err
	}

	gs, err := s.GetManyGroupsByKeys(ctx, options.GroupKeys)
	if err != nil {
		return nil, err
	}

	var seenKeys []string
	for _, g := range gs {
		q := s.SQLBuilder.
			Delete(s.SQLBuilder.TableName("_auth_group_role")).
			Where("role_id = ? AND group_id = ?", r.ID, g.ID)
		_, err := s.SQLExecutor.ExecWith(ctx, q)
		if err != nil {
			return nil, err
		}

		seenKeys = append(seenKeys, g.Key)
	}

	missingKeys := slice.ExceptStrings(options.GroupKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := GroupUnknownKeys.NewWithInfo("unknown group keys", apierrors.Details{"keys": missingKeys})
		return nil, err
	}

	return r, nil
}

type AddGroupToRolesOptions struct {
	GroupKey string
	RoleKeys []string
}

func (s *Store) AddGroupToRoles(ctx context.Context, options *AddGroupToRolesOptions) (*Group, error) {
	g, err := s.GetGroupByKey(ctx, options.GroupKey)
	if err != nil {
		return nil, err
	}

	rs, err := s.GetManyRolesByKeys(ctx, options.RoleKeys)
	if err != nil {
		return nil, err
	}

	var seenKeys []string
	now := s.Clock.NowUTC()
	for _, r := range rs {
		id := uuid.New()
		q := s.SQLBuilder.
			Insert(s.SQLBuilder.TableName("_auth_group_role")).
			Columns(
				"id",
				"created_at",
				"updated_at",
				"group_id",
				"role_id",
			).
			Values(
				id,
				now,
				now,
				g.ID,
				r.ID,
			).Suffix("ON CONFLICT DO NOTHING")

		_, err := s.SQLExecutor.ExecWith(ctx, q)
		if err != nil {
			return nil, err
		}

		seenKeys = append(seenKeys, r.Key)
	}

	missingKeys := slice.ExceptStrings(options.RoleKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := GroupUnknownKeys.NewWithInfo("unknown role keys", apierrors.Details{"keys": missingKeys})
		return nil, err
	}

	return g, nil
}

type RemoveGroupFromRolesOptions struct {
	GroupKey string
	RoleKeys []string
}

func (s *Store) RemoveGroupFromRoles(ctx context.Context, options *RemoveGroupFromRolesOptions) (*Group, error) {
	g, err := s.GetGroupByKey(ctx, options.GroupKey)
	if err != nil {
		return nil, err
	}

	rs, err := s.GetManyRolesByKeys(ctx, options.RoleKeys)
	if err != nil {
		return nil, err
	}

	var seenKeys []string
	for _, r := range rs {
		q := s.SQLBuilder.
			Delete(s.SQLBuilder.TableName("_auth_group_role")).
			Where("role_id = ? AND group_id = ?", r.ID, g.ID)

		_, err := s.SQLExecutor.ExecWith(ctx, q)
		if err != nil {
			return nil, err
		}

		seenKeys = append(seenKeys, r.Key)
	}

	missingKeys := slice.ExceptStrings(options.RoleKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := GroupUnknownKeys.NewWithInfo("unknown role keys", apierrors.Details{"keys": missingKeys})
		return nil, err
	}

	return g, nil
}
