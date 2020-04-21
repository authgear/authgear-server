package adaptors

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type LoginIDIdentityProvider interface {
	Get(userID, id string) (*loginid.Identity, error)
	GetByLoginID(loginID loginid.LoginID) ([]*loginid.Identity, error)
}

type OAuthIdentityProvider interface {
	Get(userID, id string) (*oauth.Identity, error)
	GetByProviderSubject(provider oauth.ProviderID, subjectID string) (*oauth.Identity, error)
}

type IdentityAdaptor struct {
	LoginID LoginIDIdentityProvider
	OAuth   OAuthIdentityProvider
}

func (a *IdentityAdaptor) Get(userID string, typ authn.IdentityType, id string) (*interaction.IdentityInfo, error) {
	switch typ {
	case authn.IdentityTypeLoginID:
		l, err := a.LoginID.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return loginIDToIdentityInfo(l), nil

	case authn.IdentityTypeOAuth:
		o, err := a.OAuth.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return oauthToIdentityInfo(o), nil
	}

	panic("interaction_adaptors: unknown identity type " + typ)
}

func (a *IdentityAdaptor) GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *interaction.IdentityInfo, error) {
	switch typ {
	case authn.IdentityTypeLoginID:
		loginIDKey := ""
		if v, ok := claims[interaction.IdentityClaimLoginIDKey]; ok {
			if loginIDKey, ok = v.(string); !ok {
				panic(fmt.Sprintf("interaction_adaptors: expect string login ID key, got %T", claims[interaction.IdentityClaimLoginIDKey]))
			}
		}
		loginID, ok := claims[interaction.IdentityClaimLoginIDValue].(string)
		if !ok {
			panic(fmt.Sprintf("interaction_adaptors: expect string login ID value, got %T", claims[interaction.IdentityClaimLoginIDValue]))
		}

		l, err := a.LoginID.GetByLoginID(loginid.LoginID{Key: loginIDKey, Value: loginID})
		if err != nil {
			return "", nil, err
		} else if len(l) != 1 {
			return "", nil, identity.ErrIdentityNotFound
		}
		return l[0].UserID, loginIDToIdentityInfo(l[0]), nil

	case authn.IdentityTypeOAuth:
		provider, ok := claims[interaction.IdentityClaimOAuthProvider].(map[string]interface{})
		if !ok {
			panic(fmt.Sprintf("interaction_adaptors: expect map provider claim, got %T", claims[interaction.IdentityClaimOAuthProvider]))
		}
		subjectID, ok := claims[interaction.IdentityClaimOAuthSubjectID].(string)
		if !ok {
			panic(fmt.Sprintf("interaction_adaptors: expect string subject ID claim, got %T", claims[interaction.IdentityClaimOAuthSubjectID]))
		}

		providerID := oauth.ProviderID{Keys: map[string]interface{}{}}
		for k, v := range provider {
			if k == "type" {
				providerID.Type, ok = v.(string)
				if !ok {
					panic(fmt.Sprintf("interaction_adaptors: expect string provider type, got %T", v))
				}
			} else {
				providerID.Keys[k] = v
			}
		}

		o, err := a.OAuth.GetByProviderSubject(providerID, subjectID)
		if err != nil {
			return "", nil, err
		}
		return o.UserID, oauthToIdentityInfo(o), nil
	}

	panic("interaction_adaptors: unknown identity type " + typ)
}
