package rolesgroups

import (
	"sort"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/setutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func (s *Store) ListRolesByUserIDs(userIDs []string) (map[string][]*Role, error) {
	q := s.SQLBuilder.Select(
		"ur.user_id",
		"r.id",
		"r.created_at",
		"r.updated_at",
		"r.key",
		"r.name",
		"r.description",
	).
		From(s.SQLBuilder.TableName("_auth_user_role"), "ur").
		Join(s.SQLBuilder.TableName("_auth_role"), "r", "ur.role_id = r.id").
		Where("ur.user_id = ANY (?)", pq.Array(userIDs)).
		OrderBy("ur.created_at")

	return s.queryRolesWithUserID(q)
}

func (s *Store) DeleteUserRole(userID string) error {
	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_user_role")).
		Where("user_id = ?", userID)

	_, err := s.SQLExecutor.ExecWith(q)
	if err != nil {
		return err
	}

	return nil

}

func (s *Store) ListEffectiveRolesByUserID(userID string) ([]*Role, error) {
	roleFromGroupsQuery := s.SQLBuilder.Select(
		"r.id",
		"r.created_at",
		"r.updated_at",
		"r.key",
		"r.name",
		"r.description",
	).
		From(s.SQLBuilder.TableName("_auth_user"), "u").
		Join(s.SQLBuilder.TableName("_auth_user_group"), "ug", "u.id = ug.user_id").
		Join(s.SQLBuilder.TableName("_auth_group_role"), "gr", "ug.group_id = gr.group_id").
		Join(s.SQLBuilder.TableName("_auth_role"), "r", "r.id = gr.role_id").
		Where("ug.user_id = ?", userID).
		OrderBy("ug.created_at")

	roleFromUserQuery := s.SQLBuilder.Select(
		"r.id",
		"r.created_at",
		"r.updated_at",
		"r.key",
		"r.name",
		"r.description",
	).
		From(s.SQLBuilder.TableName("_auth_user_role"), "ur").
		Join(s.SQLBuilder.TableName("_auth_role"), "r", "ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		OrderBy("ur.created_at")

	roleFromGroups, err := s.queryRoles(roleFromGroupsQuery)
	if err != nil {
		return nil, err
	}
	roleFromUser, err := s.queryRoles(roleFromUserQuery)
	if err != nil {
		return nil, err
	}

	mergedList := append(roleFromGroups, roleFromUser...)
	sort.Slice(mergedList, func(i, j int) bool {
		return mergedList[i].CreatedAt.Unix() < mergedList[j].CreatedAt.Unix()
	})
	deduplicatedList := make([]*Role, 0)
	check := make(map[string]bool, len(mergedList))
	for i := range mergedList {
		if !check[mergedList[i].Key] {
			deduplicatedList = append(deduplicatedList, mergedList[i])
			check[mergedList[i].Key] = true
		}
	}

	return deduplicatedList, nil
}

func (s *Store) ListRolesByUserID(userID string) ([]*Role, error) {
	userRoles, err := s.ListRolesByUserIDs([]string{userID})
	if err != nil {
		return nil, err
	}

	return userRoles[userID], nil
}

func (s *Store) ListUserIDsByRoleID(roleID string, pageArgs graphqlutil.PageArgs) ([]string, uint64, error) {
	q := s.SQLBuilder.Select(
		"u.id",
	).
		From(s.SQLBuilder.TableName("_auth_user_role"), "ur").
		Join(s.SQLBuilder.TableName("_auth_user"), "u", "ur.user_id = u.id").
		Where("ur.role_id = ?", roleID)

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

func (s *Store) ListAllUserIDsByRoleID(roleIDs []string) ([]string, error) {
	q := s.SQLBuilder.Select(
		"u.id",
	).
		From(s.SQLBuilder.TableName("_auth_user_role"), "ur").
		Join(s.SQLBuilder.TableName("_auth_user"), "u", "ur.user_id = u.id").
		Where("ur.role_id = ANY (?)", pq.Array(roleIDs))

	userIDs, err := s.queryUserIDs(q)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (s *Store) ListAllUserIDsByEffectiveRoleIDs(roleIDs []string) ([]string, error) {
	userRoleUserIDsQuery := s.SQLBuilder.Select(
		"ur.user_id",
	).
		From(s.SQLBuilder.TableName("_auth_user_role"), "ur").
		Where("ur.role_id = ANY (?)", pq.Array(roleIDs))

	userRoleUserIDs, err := s.queryUserIDs(userRoleUserIDsQuery)
	if err != nil {
		return nil, err
	}

	userGroupRoleUserIDsQuery := s.SQLBuilder.Select(
		"ug.user_id",
	).
		From(s.SQLBuilder.TableName("_auth_user_group"), "ug").
		Join(s.SQLBuilder.TableName("_auth_group_role"), "gr", "ug.group_id = gr.group_id").
		Where("gr.role_id = ANY (?)", pq.Array(roleIDs))

	userGroupRoleUserIDs, err := s.queryUserIDs(userGroupRoleUserIDsQuery)
	if err != nil {
		return nil, err
	}

	combinedUserIDs := []string{}
	combinedUserIDs = append(combinedUserIDs, userRoleUserIDs...)
	combinedUserIDs = append(combinedUserIDs, userGroupRoleUserIDs...)

	userIDsSet := setutil.NewSetFromSlice(combinedUserIDs, setutil.Identity[string])
	return setutil.SetToSlice(combinedUserIDs, userIDsSet, setutil.Identity[string]), nil
}

type AddRoleToUsersOptions struct {
	RoleKey string
	UserIDs []string
}

type ResetUserRoleOptions struct {
	UserID   string
	RoleKeys []string
}

func (s *Store) ResetUserRole(options *ResetUserRoleOptions) error {
	currentRoles, err := s.ListRolesByUserID(options.UserID)
	if err != nil {
		return err
	}
	originalKeys := make([]string, len(currentRoles))
	for i, v := range currentRoles {
		originalKeys[i] = v.Key
	}
	keysToAdd, keysToRemove := computeKeyDifference(originalKeys, options.RoleKeys)

	if len(keysToRemove) != 0 {
		err := s.RemoveUserFromRoles(&RemoveUserFromRolesOptions{
			UserID:   options.UserID,
			RoleKeys: keysToRemove,
		})
		if err != nil {
			return err
		}
	}

	if len(keysToAdd) != 0 {
		err := s.AddUserToRoles(&AddUserToRolesOptions{
			UserID:   options.UserID,
			RoleKeys: keysToAdd,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) AddRoleToUsers(options *AddRoleToUsersOptions) (*Role, error) {
	r, err := s.GetRoleByKey(options.RoleKey)
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
			Insert(s.SQLBuilder.TableName("_auth_user_role")).
			Columns(
				"id",
				"created_at",
				"updated_at",
				"user_id",
				"role_id",
			).
			Values(
				id,
				now,
				now,
				u,
				r.ID,
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

	return r, nil
}

type RemoveRoleFromUsersOptions struct {
	RoleKey string
	UserIDs []string
}

func (s *Store) RemoveRoleFromUsers(options *RemoveRoleFromUsersOptions) (*Role, error) {
	r, err := s.GetRoleByKey(options.RoleKey)
	if err != nil {
		return nil, err
	}

	userIds, err := s.GetManyUsersByIds(options.UserIDs)
	if err != nil {
		return nil, err
	}

	var seenKeys []string
	for _, u := range userIds {
		q := s.SQLBuilder.
			Delete(s.SQLBuilder.TableName("_auth_user_role")).
			Where("role_id = ? AND user_id = ?", r.ID, u)

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

type AddUserToRolesOptions struct {
	UserID   string
	RoleKeys []string
}

func (s *Store) AddUserToRoles(options *AddUserToRolesOptions) error {
	u, err := s.GetUserByID(options.UserID)
	if err != nil {
		return err
	}

	rs, err := s.GetManyRolesByKeys(options.RoleKeys)
	if err != nil {
		return err
	}

	var seenKeys []string
	now := s.Clock.NowUTC()
	for _, r := range rs {
		id := uuid.New()
		q := s.SQLBuilder.
			Insert(s.SQLBuilder.TableName("_auth_user_role")).
			Columns(
				"id",
				"created_at",
				"updated_at",
				"user_id",
				"role_id",
			).
			Values(
				id,
				now,
				now,
				u,
				r.ID,
			).Suffix("ON CONFLICT DO NOTHING")

		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return err
		}

		seenKeys = append(seenKeys, r.Key)
	}

	missingKeys := slice.ExceptStrings(options.RoleKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := RoleUnknownKeys.NewWithInfo("unknown role keys", apierrors.Details{"keys": missingKeys})
		return err
	}

	return nil
}

type RemoveUserFromRolesOptions struct {
	UserID   string
	RoleKeys []string
}

func (s *Store) RemoveUserFromRoles(options *RemoveUserFromRolesOptions) error {
	u, err := s.GetUserByID(options.UserID)
	if err != nil {
		return err
	}

	rs, err := s.GetManyRolesByKeys(options.RoleKeys)
	if err != nil {
		return err
	}

	var seenKeys []string
	for _, r := range rs {
		q := s.SQLBuilder.
			Delete(s.SQLBuilder.TableName("_auth_user_role")).
			Where("role_id = ? AND user_id = ?", r.ID, u)

		_, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return err
		}

		seenKeys = append(seenKeys, r.Key)
	}

	missingKeys := slice.ExceptStrings(options.RoleKeys, seenKeys)
	if len(missingKeys) > 0 {
		err := RoleUnknownKeys.NewWithInfo("unknown role keys", apierrors.Details{"keys": missingKeys})
		return err
	}

	return nil
}
