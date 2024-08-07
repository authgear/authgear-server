package user

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type IdentityService interface {
	ListByUserIDs(userIDs []string) (map[string][]*identity.Info, error)
}

type AuthenticatorService interface {
	ListByUserIDs(userIDs []string, filters ...authenticator.Filter) (map[string][]*authenticator.Info, error)
}

type VerificationService interface {
	IsUserVerified(identities []*identity.Info) (bool, error)
	AreUsersVerified(identitiesByUserIDs map[string][]*identity.Info) (map[string]bool, error)
}

type StandardAttributesService interface {
	DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
	DeriveStandardAttributesForUsers(
		role accesscontrol.Role,
		userIDs []string,
		updatedAts []time.Time,
		attrsList []map[string]interface{},
	) (map[string]map[string]interface{}, error)
}

type CustomAttributesService interface {
	ReadCustomAttributesInStorageForm(role accesscontrol.Role, userID string, storageForm map[string]interface{}) (map[string]interface{}, error)
	ReadCustomAttributesInStorageFormForUsers(
		role accesscontrol.Role,
		userIDs []string,
		storageForms []map[string]interface{},
	) (map[string]map[string]interface{}, error)
}

type Web3Service interface {
	GetWeb3Info(identities []*identity.Info) (*model.UserWeb3Info, error)
}

type RolesAndGroupsService interface {
	ListRolesByUserID(userID string) ([]*model.Role, error)
	ListGroupsByUserID(userID string) ([]*model.Group, error)
	ListRolesByUserIDs(userIDs []string) (map[string][]*model.Role, error)
	ListGroupsByUserIDs(userIDs []string) (map[string][]*model.Group, error)
}

type Queries struct {
	*RawQueries
	Store              store
	Identities         IdentityService
	Authenticators     AuthenticatorService
	Verification       VerificationService
	StandardAttributes StandardAttributesService
	CustomAttributes   CustomAttributesService
	Web3               Web3Service
	RolesAndGroups     RolesAndGroupsService
}

func (p *Queries) Get(id string, role accesscontrol.Role) (*model.User, error) {
	users, err := p.GetMany([]string{id}, role)
	if err != nil {
		return nil, err
	}

	if len(users) != 1 {
		return nil, ErrUserNotFound
	}

	return users[0], nil
}

func (p *Queries) GetMany(ids []string, role accesscontrol.Role) (users []*model.User, err error) {
	rawUsers, err := p.GetManyRaw(ids)
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

	identitiesByUserID, err := p.Identities.ListByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}

	authenticatorsByUserID, err := p.Authenticators.ListByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}

	isVerifiedByUserID, err := p.Verification.AreUsersVerified(identitiesByUserID)
	if err != nil {
		return nil, err
	}

	stdAttrsByUserID, err := p.StandardAttributes.DeriveStandardAttributesForUsers(
		role,
		userIDs,
		updatedAts,
		stdAttrsList,
	)
	if err != nil {
		return nil, err
	}

	customAttrsByUserID, err := p.CustomAttributes.ReadCustomAttributesInStorageFormForUsers(
		role,
		userIDs,
		customAttrsList)
	if err != nil {
		return nil, err
	}

	rolesByUserID, err := p.RolesAndGroups.ListRolesByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}
	groupsByUserID, err := p.RolesAndGroups.ListGroupsByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}

	for _, rawUser := range rawUsers {
		if rawUser == nil {
			users = append(users, nil)
		} else {
			identities := identitiesByUserID[rawUser.ID]
			authenticators := authenticatorsByUserID[rawUser.ID]
			isVerified := isVerifiedByUserID[rawUser.ID]
			stdAttrs := stdAttrsByUserID[rawUser.ID]
			customAttrs := customAttrsByUserID[rawUser.ID]
			web3Info, web3err := p.Web3.GetWeb3Info(identities)
			if web3err != nil {
				return nil, web3err
			}
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
				identities,
				authenticators,
				isVerified,
				stdAttrs,
				customAttrs,
				web3Info,
				roleKeys,
				groupKeys,
			)

			users = append(users, u)
		}
	}

	return
}
