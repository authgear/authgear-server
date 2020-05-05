package adaptors

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type LoginIDIdentityProvider interface {
	Get(userID, id string) (*loginid.Identity, error)
	GetByLoginID(loginID loginid.LoginID) ([]*loginid.Identity, error)
	ListByClaim(name string, value string) ([]*loginid.Identity, error)
	New(userID string, loginID loginid.LoginID) *loginid.Identity
	Create(i *loginid.Identity) error
	Validate(loginIDs []loginid.LoginID) error
	Normalize(loginID loginid.LoginID) (normalized *loginid.LoginID, typ string, err error)
}

type OAuthIdentityProvider interface {
	Get(userID, id string) (*oauth.Identity, error)
	GetByProviderSubject(provider oauth.ProviderID, subjectID string) (*oauth.Identity, error)
	ListByClaim(name string, value string) ([]*oauth.Identity, error)
	New(
		userID string,
		provider oauth.ProviderID,
		subjectID string,
		profile map[string]interface{},
		claims map[string]interface{},
	) *oauth.Identity
	Create(i *oauth.Identity) error
	Update(i *oauth.Identity) error
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

// GetByClaims return user ID and information about the identity the matches the provided skygear claims.
func (a *IdentityAdaptor) GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *interaction.IdentityInfo, error) {
	switch typ {
	case authn.IdentityTypeLoginID:
		loginID := extractLoginIDClaims(claims)
		l, err := a.LoginID.GetByLoginID(loginID)
		if err != nil {
			return "", nil, err
		} else if len(l) != 1 {
			return "", nil, identity.ErrIdentityNotFound
		}
		return l[0].UserID, loginIDToIdentityInfo(l[0]), nil

	case authn.IdentityTypeOAuth:
		providerID, subjectID := extractOAuthClaims(claims)
		o, err := a.OAuth.GetByProviderSubject(providerID, subjectID)
		if err != nil {
			return "", nil, err
		}
		return o.UserID, oauthToIdentityInfo(o), nil
	}

	panic("interaction_adaptors: unknown identity type " + typ)
}

// ListByClaims return list of identities the matches the provided OIDC standard claims.
func (a *IdentityAdaptor) ListByClaims(claims map[string]string) ([]*interaction.IdentityInfo, error) {
	var all []*interaction.IdentityInfo

	for name, value := range claims {
		ls, err := a.LoginID.ListByClaim(name, value)
		if err != nil {
			return nil, err
		}
		for _, i := range ls {
			all = append(all, loginIDToIdentityInfo(i))
		}

		os, err := a.OAuth.ListByClaim(name, value)
		if err != nil {
			return nil, err
		}
		for _, i := range os {
			all = append(all, oauthToIdentityInfo(i))
		}
	}

	return all, nil
}

func (a *IdentityAdaptor) New(userID string, typ authn.IdentityType, claims map[string]interface{}) *interaction.IdentityInfo {
	switch typ {
	case authn.IdentityTypeLoginID:
		loginID := extractLoginIDClaims(claims)
		l := a.LoginID.New(userID, loginID)
		return loginIDToIdentityInfo(l)

	case authn.IdentityTypeOAuth:
		providerID, subjectID := extractOAuthClaims(claims)
		var profile, oidcClaims map[string]interface{}
		var ok bool
		if profile, ok = claims[interaction.IdentityClaimOAuthProfile].(map[string]interface{}); !ok {
			profile = map[string]interface{}{}
		}
		if oidcClaims, ok = claims[interaction.IdentityClaimOAuthClaims].(map[string]interface{}); !ok {
			oidcClaims = map[string]interface{}{}
		}

		o := a.OAuth.New(userID, providerID, subjectID, profile, oidcClaims)
		return oauthToIdentityInfo(o)
	}

	panic("interaction_adaptors: unknown identity type " + typ)
}

func (a *IdentityAdaptor) WithClaims(userID string, ii *interaction.IdentityInfo, claims map[string]interface{}) *interaction.IdentityInfo {
	switch ii.Type {
	case authn.IdentityTypeLoginID:
		panic("interaction_adaptors: update no support for identity type " + ii.Type)
	case authn.IdentityTypeOAuth:
		var profile map[string]interface{}
		var ok bool
		if profile, ok = claims[interaction.IdentityClaimOAuthProfile].(map[string]interface{}); !ok {
			profile = map[string]interface{}{}
		}
		i := oauthFromIdentityInfo(userID, ii)
		i.UserProfile = profile
		return oauthToIdentityInfo(i)
	}
	panic("interaction_adaptors: unknown identity type " + ii.Type)
}

func (a *IdentityAdaptor) CreateAll(userID string, is []*interaction.IdentityInfo) error {
	for _, i := range is {
		switch i.Type {
		case authn.IdentityTypeLoginID:
			identity := loginIDFromIdentityInfo(userID, i)
			if err := a.LoginID.Create(identity); err != nil {
				return err
			}

		case authn.IdentityTypeOAuth:
			identity := oauthFromIdentityInfo(userID, i)
			if err := a.OAuth.Create(identity); err != nil {
				return err
			}
		default:

			panic("interaction_adaptors: unknown identity type " + i.Type)
		}
	}
	return nil
}

func (a *IdentityAdaptor) UpdateAll(userID string, is []*interaction.IdentityInfo) error {
	for _, i := range is {
		switch i.Type {
		case authn.IdentityTypeLoginID:
			panic("interaction_adaptors: update no support for identity type " + i.Type)
		case authn.IdentityTypeOAuth:
			identity := oauthFromIdentityInfo(userID, i)
			if err := a.OAuth.Update(identity); err != nil {
				return err
			}
		default:
			panic("interaction_adaptors: unknown identity type " + i.Type)
		}
	}
	return nil
}

func (a *IdentityAdaptor) Validate(is []*interaction.IdentityInfo) error {
	var loginIDs []loginid.LoginID
	for _, i := range is {
		if i.Type == authn.IdentityTypeLoginID {
			loginID := extractLoginIDClaims(i.Claims)
			loginIDs = append(loginIDs, loginID)
		}
	}

	// if there is IdentityInfo with type is loginid
	if len(loginIDs) > 0 {
		if err := a.LoginID.Validate(loginIDs); err != nil {
			return err
		}
	}

	return nil
}

func (a *IdentityAdaptor) RelateIdentityToAuthenticator(is interaction.IdentitySpec, as *interaction.AuthenticatorSpec) *interaction.AuthenticatorSpec {
	switch is.Type {
	case authn.IdentityTypeLoginID:
		// Early return for other authenticators.
		if as.Type != authn.AuthenticatorTypeOOB {
			return as
		}

		loginID, loginIDType, err := a.LoginID.Normalize(extractLoginIDClaims(is.Claims))
		if err != nil {
			return nil
		}

		switch loginIDType {
		case string(metadata.Email):
			as.Props[interaction.AuthenticatorPropOOBOTPChannelType] = string(authn.AuthenticatorOOBChannelEmail)
			as.Props[interaction.AuthenticatorPropOOBOTPEmail] = loginID.Value
			return as
		case string(metadata.Phone):
			as.Props[interaction.AuthenticatorPropOOBOTPChannelType] = string(authn.AuthenticatorOOBChannelSMS)
			as.Props[interaction.AuthenticatorPropOOBOTPPhone] = loginID.Value
			return as
		default:
			return nil
		}
	case authn.IdentityTypeOAuth:
		return nil
	}

	panic("interaction_adaptors: unknown identity type " + is.Type)
}

func extractLoginIDClaims(claims map[string]interface{}) loginid.LoginID {
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

	return loginid.LoginID{Key: loginIDKey, Value: loginID}
}

func extractOAuthClaims(claims map[string]interface{}) (providerID oauth.ProviderID, subjectID string) {
	provider, ok := claims[interaction.IdentityClaimOAuthProvider].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("interaction_adaptors: expect map provider claim, got %T", claims[interaction.IdentityClaimOAuthProvider]))
	}
	subjectID, ok = claims[interaction.IdentityClaimOAuthSubjectID].(string)
	if !ok {
		panic(fmt.Sprintf("interaction_adaptors: expect string subject ID claim, got %T", claims[interaction.IdentityClaimOAuthSubjectID]))
	}

	providerID = oauth.ProviderID{Keys: map[string]interface{}{}}
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

	return
}
