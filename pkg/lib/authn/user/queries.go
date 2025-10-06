package user

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type IdentityService interface {
	ListByUserIDs(ctx context.Context, userIDs []string) (map[string][]*identity.Info, error)
}

type AuthenticatorService interface {
	ListByUserIDs(ctx context.Context, userIDs []string, filters ...authenticator.Filter) (map[string][]*authenticator.Info, error)
}

type VerificationService interface {
	IsUserVerified(ctx context.Context, identities []*identity.Info) (bool, error)
	AreUsersVerified(ctx context.Context, identitiesByUserIDs map[string][]*identity.Info) (map[string]bool, error)
}

type StandardAttributesService interface {
	DeriveStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
	DeriveStandardAttributesForUsers(
		ctx context.Context,
		role accesscontrol.Role,
		userIDs []string,
		updatedAts []time.Time,
		attrsList []map[string]interface{},
	) (map[string]map[string]interface{}, error)
}

type CustomAttributesService interface {
	ReadCustomAttributesInStorageForm(ctx context.Context, role accesscontrol.Role, userID string, storageForm map[string]interface{}) (map[string]interface{}, error)
	ReadCustomAttributesInStorageFormForUsers(
		ctx context.Context,
		role accesscontrol.Role,
		userIDs []string,
		storageForms []map[string]interface{},
	) (map[string]map[string]interface{}, error)
}

type RolesAndGroupsService interface {
	ListRolesByUserID(ctx context.Context, userID string) ([]*model.Role, error)
	ListGroupsByUserID(ctx context.Context, userID string) ([]*model.Group, error)
	ListRolesByUserIDs(ctx context.Context, userIDs []string) (map[string][]*model.Role, error)
	ListGroupsByUserIDs(ctx context.Context, userIDs []string) (map[string][]*model.Group, error)
}

type Queries struct {
	*RawQueries
	Store              store
	Identities         IdentityService
	Authenticators     AuthenticatorService
	Verification       VerificationService
	StandardAttributes StandardAttributesService
	CustomAttributes   CustomAttributesService
	RolesAndGroups     RolesAndGroupsService
	Clock              clock.Clock
}

func (p *Queries) Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error) {
	users, err := p.GetMany(ctx, []string{id}, role)
	if err != nil {
		return nil, err
	}

	if len(users) != 1 {
		return nil, ErrUserNotFound
	}

	return users[0], nil
}

func (p *Queries) GetMany(ctx context.Context, ids []string, role accesscontrol.Role) (users []*model.User, err error) {
	rawUsers, err := p.GetManyRaw(ctx, ids)
	if err != nil {
		return nil, err
	}

	userIDs := []string{}
	updatedAts := []time.Time{}
	stdAttrsList := []map[string]interface{}{}
	customAttrsList := []map[string]interface{}{}
	for _, rawUser := range rawUsers {
		if rawUser == nil {
			continue
		}
		userIDs = append(userIDs, rawUser.ID)
		updatedAts = append(updatedAts, rawUser.UpdatedAt)
		stdAttrsList = append(stdAttrsList, rawUser.StandardAttributes)
		customAttrsList = append(customAttrsList, rawUser.CustomAttributes)
	}

	identitiesByUserID, err := p.Identities.ListByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	authenticatorsByUserID, err := p.Authenticators.ListByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	isVerifiedByUserID, err := p.Verification.AreUsersVerified(ctx, identitiesByUserID)
	if err != nil {
		return nil, err
	}

	stdAttrsByUserID, err := p.StandardAttributes.DeriveStandardAttributesForUsers(
		ctx,
		role,
		userIDs,
		updatedAts,
		stdAttrsList,
	)
	if err != nil {
		return nil, err
	}

	customAttrsByUserID, err := p.CustomAttributes.ReadCustomAttributesInStorageFormForUsers(
		ctx,
		role,
		userIDs,
		customAttrsList)
	if err != nil {
		return nil, err
	}

	rolesByUserID, err := p.RolesAndGroups.ListRolesByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	groupsByUserID, err := p.RolesAndGroups.ListGroupsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	now := p.Clock.NowUTC()
	for _, rawUser := range rawUsers {
		if rawUser == nil {
			users = append(users, nil)
		} else {
			identities := identitiesByUserID[rawUser.ID]
			authenticators := authenticatorsByUserID[rawUser.ID]
			isVerified := isVerifiedByUserID[rawUser.ID]
			stdAttrs := stdAttrsByUserID[rawUser.ID]
			customAttrs := customAttrsByUserID[rawUser.ID]
			roles := rolesByUserID[rawUser.ID]
			roleKeys := make([]string, len(roles))
			for i, v := range roles {
				roleKeys[i] = v.Key
			}

			groups := groupsByUserID[rawUser.ID]
			groupKeys := make([]string, len(groups))
			for i, v := range groups {
				groupKeys[i] = v.Key
			}
			u := newUserModel(
				rawUser,
				now,
				identities,
				authenticators,
				isVerified,
				stdAttrs,
				customAttrs,
				roleKeys,
				groupKeys,
			)

			users = append(users, u)
		}
	}

	return
}

func (p *Queries) GetPageForExport(ctx context.Context, offset uint64, limit uint64) (users []*UserForExport, err error) {
	rawUsers, err := p.Store.QueryForExport(ctx, offset, limit)
	if err != nil {
		return
	}

	userIDs := []string{}
	updatedAts := []time.Time{}
	stdAttrsList := []map[string]interface{}{}
	customAttrsList := []map[string]interface{}{}
	for _, rawUser := range rawUsers {
		if rawUser == nil {
			continue
		}
		userIDs = append(userIDs, rawUser.ID)
		updatedAts = append(updatedAts, rawUser.UpdatedAt)
		stdAttrsList = append(stdAttrsList, rawUser.StandardAttributes)
		customAttrsList = append(customAttrsList, rawUser.CustomAttributes)
	}

	identitiesByUserID, err := p.Identities.ListByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	authenticatorsByUserID, err := p.Authenticators.ListByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	isVerifiedByUserID, err := p.Verification.AreUsersVerified(ctx, identitiesByUserID)
	if err != nil {
		return nil, err
	}

	stdAttrsByUserID, err := p.StandardAttributes.DeriveStandardAttributesForUsers(
		ctx,
		"",
		userIDs,
		updatedAts,
		stdAttrsList,
	)
	if err != nil {
		return nil, err
	}

	customAttrsByUserID, err := p.CustomAttributes.ReadCustomAttributesInStorageFormForUsers(
		ctx,
		"",
		userIDs,
		customAttrsList)
	if err != nil {
		return nil, err
	}

	rolesByUserID, err := p.RolesAndGroups.ListRolesByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	groupsByUserID, err := p.RolesAndGroups.ListGroupsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	now := p.Clock.NowUTC()
	for _, rawUser := range rawUsers {
		if rawUser == nil {
			users = append(users, nil)
		} else {
			identities := identitiesByUserID[rawUser.ID]
			authenticators := authenticatorsByUserID[rawUser.ID]
			isVerified := isVerifiedByUserID[rawUser.ID]
			stdAttrs := stdAttrsByUserID[rawUser.ID]
			customAttrs := customAttrsByUserID[rawUser.ID]
			roles := rolesByUserID[rawUser.ID]
			roleKeys := make([]string, len(roles))
			for i, v := range roles {
				roleKeys[i] = v.Key
			}

			groups := groupsByUserID[rawUser.ID]
			groupKeys := make([]string, len(groups))
			for i, v := range groups {
				groupKeys[i] = v.Key
			}
			u := newUserModel(
				rawUser,
				now,
				identities,
				authenticators,
				isVerified,
				stdAttrs,
				customAttrs,
				roleKeys,
				groupKeys,
			)

			userForExport := UserForExport{
				User:           *u,
				Identities:     identities,
				Authenticators: authenticators,
			}

			users = append(users, &userForExport)
		}
	}

	return users, nil
}

func (p *Queries) CountAll(ctx context.Context) (count uint64, err error) {
	count, err = p.Store.Count(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}
