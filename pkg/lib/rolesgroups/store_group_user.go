package rolesgroups

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type AddGroupToUsersOptions struct {
	GroupKey string
	UserIDs  []string
}

func (s *Store) AddGroupToUsers(options *AddGroupToUsersOptions) (*Group, error) {
	g, err := s.GetGroupByKey(options.GroupKey)
	if err != nil {
		return nil, err
	}

	userIds, err := s.GetManyUsersByIds(options.UserIDs)
	if err != nil {
		return nil, err
	}

	var seenKeys []string
	now := s.Clock.NowUTC()
	for _, u := range userIds {
		id := uuid.New()
		q := s.SQLBuilder.
			Insert(s.SQLBuilder.TableName("_auth_user_group")).
			Columns(
				"id",
				"created_at",
				"updated_at",
				"user_id",
				"group_id",
			).
			Values(
				id,
				now,
				now,
				u,
				g.ID,
			).Suffix("ON CONFLICT DO NOTHING")

		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return nil, err
		}

		seenKeys = append(seenKeys, u)
	}

	missingKeys := slice.ExceptStrings(options.UserIDs, seenKeys)
	if len(missingKeys) > 0 {
		err := UserUnknownKeys.NewWithInfo("unknown user ids", apierrors.Details{"keys": missingKeys})
		return nil, err
	}

	return g, nil
}

type RemoveGroupFromUsersOptions struct {
	GroupKey string
	UserIDs  []string
}

func (s *Store) RemoveGroupFromUsers(options *RemoveGroupFromUsersOptions) (*Group, error) {
	r, err := s.GetGroupByKey(options.GroupKey)
	if err != nil {
		return nil, err
	}

	users, err := s.GetManyUsersByIds(options.UserIDs)
	if err != nil {
		return nil, err
	}

	var seenKeys []string
	for _, u := range users {
		q := s.SQLBuilder.
			Delete(s.SQLBuilder.TableName("_auth_user_group")).
			Where("group_id = ? AND user_id = ?", r.ID, u)
		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return nil, err
		}

		seenKeys = append(seenKeys, u)
	}

	missingKeys := slice.ExceptStrings(options.UserIDs, seenKeys)
	if len(missingKeys) > 0 {
		err := UserUnknownKeys.NewWithInfo("unknown group keys", apierrors.Details{"keys": missingKeys})
		return nil, err
	}

	return r, nil
}

type AddUserToGroupsOptions struct {
	UserID    string
	GroupKeys []string
}

func (s *Store) AddUserToGroups(options *AddUserToGroupsOptions) error {
	u, err := s.GetUserByID(options.UserID)
	if err != nil {
		return err
	}

	gs, err := s.GetManyGroupsByKeys(options.GroupKeys)
	if err != nil {
		return err
	}

	var seenKeys []string
	now := s.Clock.NowUTC()
	for _, g := range gs {
		id := uuid.New()
		q := s.SQLBuilder.
			Insert(s.SQLBuilder.TableName("_auth_user_group")).
			Columns(
				"id",
				"created_at",
				"updated_at",
				"user_id",
				"group_id",
			).
			Values(
				id,
				now,
				now,
				u,
				g.ID,
			).Suffix("ON CONFLICT DO NOTHING")

		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return err
		}

		seenKeys = append(seenKeys, g.Key)
	}

	missingKeys := slice.ExceptStrings(options.GroupKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := GroupUnknownKeys.NewWithInfo("unknown group keys", apierrors.Details{"keys": missingKeys})
		return err
	}

	return nil
}

type RemoveUserFromGroupsOptions struct {
	UserID    string
	GroupKeys []string
}

func (s *Store) RemoveUserFromGroups(options *RemoveUserFromGroupsOptions) error {
	u, err := s.GetUserByID(options.UserID)
	if err != nil {
		return err
	}

	gs, err := s.GetManyGroupsByKeys(options.GroupKeys)
	if err != nil {
		return err
	}

	var seenKeys []string
	for _, g := range gs {
		q := s.SQLBuilder.
			Delete(s.SQLBuilder.TableName("_auth_user_group")).
			Where("group_id = ? AND user_id = ?", g.ID, u)

		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return err
		}

		seenKeys = append(seenKeys, g.Key)
	}

	missingKeys := slice.ExceptStrings(options.GroupKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := GroupUnknownKeys.NewWithInfo("unknown group keys", apierrors.Details{"keys": missingKeys})
		return err
	}

	return nil
}
