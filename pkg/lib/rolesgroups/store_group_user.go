package rolesgroups

import (
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func (s *Store) ListGroupsByUserIDs(userIDs []string) (map[string][]*Group, error) {
	q := s.SQLBuilder.Select(
		"ug.user_id",
		"g.id",
		"g.created_at",
		"g.updated_at",
		"g.key",
		"g.name",
		"g.description",
	).
		From(s.SQLBuilder.TableName("_auth_user_group"), "ug").
		Join(s.SQLBuilder.TableName("_auth_group"), "g", "ug.group_id = g.id").
		Where("ug.user_id = ANY (?)", pq.Array(userIDs)).
		OrderBy("ug.created_at")

	return s.queryGroupsWithUserID(q)
}

func (s *Store) ListGroupsByUserID(userID string) ([]*Group, error) {
	userGroups, err := s.ListGroupsByUserIDs([]string{userID})
	if err != nil {
		return nil, err
	}

	return userGroups[userID], nil
}

func (s *Store) queryUserIDsByGroupIDs(groupIDs []string) db.SelectBuilder {
	return s.SQLBuilder.Select(
		"u.id",
	).
		From(s.SQLBuilder.TableName("_auth_user_group"), "ug").
		Join(s.SQLBuilder.TableName("_auth_user"), "u", "ug.user_id = u.id").
		Where("ug.group_id = ANY (?)", pq.Array(groupIDs))
}

func (s *Store) ListAllUserIDsByGroupIDs(groupIDs []string) ([]string, error) {
	q := s.queryUserIDsByGroupIDs(groupIDs)

	userIDs, err := s.queryUserIDs(q)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (s *Store) ListUserIDsByGroupID(groupID string, pageArgs graphqlutil.PageArgs) ([]string, uint64, error) {
	q := s.queryUserIDsByGroupIDs([]string{groupID})

	q, offset, err := db.ApplyPageArgs(q, pageArgs)
	if err != nil {
		return nil, 0, err
	}

	userIDs, err := s.queryUserIDs(q)
	if err != nil {
		return nil, 0, err
	}

	return userIDs, offset, nil
}

type ResetUserGroupOptions struct {
	UserID    string
	GroupKeys []string
}

func (s *Store) ResetUserGroup(options *ResetUserGroupOptions) error {
	currentGroups, err := s.ListGroupsByUserID(options.UserID)
	if err != nil {
		return err
	}

	originalKeys := make([]string, len(currentGroups))
	for i, v := range currentGroups {
		originalKeys[i] = v.Key
	}
	keysToAdd, keysToRemove := computeKeyDifference(originalKeys, options.GroupKeys)

	if len(keysToRemove) != 0 {
		err := s.RemoveUserFromGroups(&RemoveUserFromGroupsOptions{
			UserID:    options.UserID,
			GroupKeys: keysToRemove,
		})
		if err != nil {
			return err
		}
	}

	if len(keysToAdd) != 0 {
		err := s.AddUserToGroups(&AddUserToGroupsOptions{
			UserID:    options.UserID,
			GroupKeys: keysToAdd,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) DeleteUserGroup(userID string) error {
	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_user_group")).
		Where("user_id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil

}

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
		err := UserUnknownKeys.NewWithInfo("unknown user ids", apierrors.Details{"ids": missingKeys})
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
		err := UserUnknownKeys.NewWithInfo("unknown user ids", apierrors.Details{"ids": missingKeys})
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
