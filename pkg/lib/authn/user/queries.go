package user

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type VerificationService interface {
	IsUserVerified(identities []*identity.Info) (bool, error)
}

type StandardAttributesService interface {
	DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
}

type CustomAttributesService interface {
	ReadCustomAttributesInStorageForm(role accesscontrol.Role, userID string, storageForm map[string]interface{}) (map[string]interface{}, error)
}

type Web3Service interface {
	GetWeb3Info(addresses []string) (map[string]interface{}, error)
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
}

func (p *Queries) Get(id string, role accesscontrol.Role) (*model.User, error) {
	user, err := p.RawQueries.GetRaw(id)
	if err != nil {
		return nil, err
	}

	identities, err := p.Identities.ListByUser(id)
	if err != nil {
		return nil, err
	}

	authenticators, err := p.Authenticators.List(id)
	if err != nil {
		return nil, err
	}

	isVerified, err := p.Verification.IsUserVerified(identities)
	if err != nil {
		return nil, err
	}

	stdAttrs, err := p.StandardAttributes.DeriveStandardAttributes(role, id, user.UpdatedAt, user.StandardAttributes)
	if err != nil {
		return nil, err
	}

	customAttrs, err := p.CustomAttributes.ReadCustomAttributesInStorageForm(role, id, user.CustomAttributes)
	if err != nil {
		return nil, err
	}

	web3Addresses := make([]string, 0)
	for _, i := range identities {
		if i.Type == model.IdentityTypeSIWE && i.SIWE != nil {
			web3Addresses = append(web3Addresses, i.SIWE.Address)
		}
	}
	web3Info := map[string]interface{}{}
	if len(web3Addresses) > 0 {
		info, err := p.Web3.GetWeb3Info(web3Addresses)
		if err != nil {
			return nil, err
		}

		web3Info = info
	}

	return newUserModel(user, identities, authenticators, isVerified, stdAttrs, customAttrs, web3Info), nil
}

func (p *Queries) GetMany(ids []string) (users []*model.User, err error) {
	rawUsers, err := p.GetManyRaw(ids)
	if err != nil {
		return nil, err
	}

	for _, rawUser := range rawUsers {
		if rawUser == nil {
			users = append(users, nil)
		} else {
			var u *model.User
			u, err = p.Get(rawUser.ID, accesscontrol.RoleGreatest)
			if err != nil {
				return
			}
			users = append(users, u)
		}
	}

	return
}
