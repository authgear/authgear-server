package adaptors

import (
	"fmt"
	"reflect"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type LoginIDIdentityProvider interface {
	Get(userID, id string) (*loginid.Identity, error)
	List(userID string) ([]*loginid.Identity, error)
	GetByLoginID(loginID loginid.LoginID) ([]*loginid.Identity, error)
	ListByClaim(name string, value string) ([]*loginid.Identity, error)
	New(userID string, loginID loginid.LoginID) *loginid.Identity
	WithLoginID(iden *loginid.Identity, loginID loginid.LoginID) *loginid.Identity
	Create(i *loginid.Identity) error
	Update(i *loginid.Identity) error
	Delete(i *loginid.Identity) error
	Validate(loginIDs []loginid.LoginID) error
	Normalize(loginID loginid.LoginID) (normalized *loginid.LoginID, typ string, err error)
}

type OAuthIdentityProvider interface {
	Get(userID, id string) (*oauth.Identity, error)
	List(userID string) ([]*oauth.Identity, error)
	GetByProviderSubject(provider oauth.ProviderID, subjectID string) (*oauth.Identity, error)
	GetByUserProvider(userID string, provider oauth.ProviderID) (*oauth.Identity, error)
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
	Delete(i *oauth.Identity) error
}

type AnonymousIdentityProvider interface {
	Get(userID, id string) (*anonymous.Identity, error)
	GetByKeyID(keyID string) (*anonymous.Identity, error)
	ListByClaim(name string, value string) ([]*anonymous.Identity, error)
	New(userID string, keyID string, key []byte) *anonymous.Identity
	Create(i *anonymous.Identity) error
}

type IdentityAdaptor struct {
	LoginID   LoginIDIdentityProvider
	OAuth     OAuthIdentityProvider
	Anonymous AnonymousIdentityProvider
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

	case authn.IdentityTypeAnonymous:
		a, err := a.Anonymous.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return anonymousToIdentityInfo(a), nil
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

	case authn.IdentityTypeAnonymous:
		keyID, _ := extractAnonymousClaims(claims)
		a, err := a.Anonymous.GetByKeyID(keyID)
		if err != nil {
			return "", nil, err
		}
		return a.UserID, anonymousToIdentityInfo(a), nil
	}

	panic("interaction_adaptors: unknown identity type " + typ)
}

// GetByUserAndClaims return user's identity that matches the provide skygear claims.
//
// Given that user id is provided, the matching rule of this function is less strict than GetByClaims.
// For example, login id identity needs match both key and value and oauth identity only needs to match provider id.
// This function is currently in used by remove identity interaction.
func (a *IdentityAdaptor) GetByUserAndClaims(typ authn.IdentityType, userID string, claims map[string]interface{}) (*interaction.IdentityInfo, error) {
	switch typ {
	case authn.IdentityTypeOAuth:
		providerID := extractOAuthProviderClaims(claims)
		o, err := a.OAuth.GetByUserProvider(userID, providerID)
		if err != nil {
			return nil, err
		}
		return oauthToIdentityInfo(o), nil
	default:
		uid, iden, err := a.GetByClaims(typ, claims)
		if err != nil {
			return nil, err
		}
		if uid != userID {
			return nil, identity.ErrIdentityNotFound
		}
		return iden, nil
	}
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

		// Skip anonymous: no standard claims for anonymous identity
	}

	return all, nil
}

func (a *IdentityAdaptor) ListByUser(userID string) ([]*interaction.IdentityInfo, error) {
	iis := []*interaction.IdentityInfo{}

	// login id
	lis, err := a.LoginID.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range lis {
		iis = append(iis, loginIDToIdentityInfo(i))
	}

	// oauth
	ois, err := a.OAuth.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range ois {
		iis = append(iis, oauthToIdentityInfo(i))
	}

	return iis, nil
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

	case authn.IdentityTypeAnonymous:
		keyID, key := extractAnonymousClaims(claims)
		a := a.Anonymous.New(userID, keyID, []byte(key))
		return anonymousToIdentityInfo(a)
	}

	panic("interaction_adaptors: unknown identity type " + typ)
}

func (a *IdentityAdaptor) WithClaims(userID string, ii *interaction.IdentityInfo, claims map[string]interface{}) *interaction.IdentityInfo {
	switch ii.Type {
	case authn.IdentityTypeLoginID:
		oldIden := loginIDFromIdentityInfo(userID, ii)
		newLoginID := extractLoginIDClaims(claims)
		newIden := a.LoginID.WithLoginID(oldIden, newLoginID)
		return loginIDToIdentityInfo(newIden)

	case authn.IdentityTypeOAuth:
		var profile map[string]interface{}
		var ok bool
		if profile, ok = claims[interaction.IdentityClaimOAuthProfile].(map[string]interface{}); !ok {
			profile = map[string]interface{}{}
		}
		i := oauthFromIdentityInfo(userID, ii)
		i.UserProfile = profile
		return oauthToIdentityInfo(i)
	case authn.IdentityTypeAnonymous:
		panic("interaction_adaptors: update no support for identity type " + ii.Type)
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

		case authn.IdentityTypeAnonymous:
			identity := anonymousFromIdentityInfo(userID, i)
			if err := a.Anonymous.Create(identity); err != nil {
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
			identity := loginIDFromIdentityInfo(userID, i)
			if err := a.LoginID.Update(identity); err != nil {
				return err
			}
		case authn.IdentityTypeOAuth:
			identity := oauthFromIdentityInfo(userID, i)
			if err := a.OAuth.Update(identity); err != nil {
				return err
			}
		case authn.IdentityTypeAnonymous:
			panic("interaction_adaptors: update no support for identity type " + i.Type)
		default:
			panic("interaction_adaptors: unknown identity type " + i.Type)
		}
	}
	return nil
}

func (a *IdentityAdaptor) DeleteAll(userID string, is []*interaction.IdentityInfo) error {
	for _, i := range is {
		switch i.Type {
		case authn.IdentityTypeLoginID:
			identity := loginIDFromIdentityInfo(userID, i)
			if err := a.LoginID.Delete(identity); err != nil {
				return err
			}
		case authn.IdentityTypeOAuth:
			identity := oauthFromIdentityInfo(userID, i)
			if err := a.OAuth.Delete(identity); err != nil {
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
	var oauthProviderIDs []oauth.ProviderID
	for _, i := range is {
		if i.Type == authn.IdentityTypeLoginID {
			loginID := extractLoginIDClaims(i.Claims)
			loginIDs = append(loginIDs, loginID)
		} else if i.Type == authn.IdentityTypeOAuth {
			providerID, _ := extractOAuthClaims(i.Claims)
			oauthProviderIDs = append(oauthProviderIDs, providerID)
		}
	}

	// if there is IdentityInfo with type is loginid
	if len(loginIDs) > 0 {
		if err := a.LoginID.Validate(loginIDs); err != nil {
			return err
		}
	}

	// oauth identity check duplicate provider
	if len(oauthProviderIDs) > 0 {
		for i, l := range oauthProviderIDs {
			for j, r := range oauthProviderIDs {
				if i != j && reflect.DeepEqual(l, r) {
					return identity.ErrIdentityAlreadyExists
				}
			}
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
	case authn.IdentityTypeAnonymous:
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
	providerID = extractOAuthProviderClaims(claims)

	subjectID, ok := claims[interaction.IdentityClaimOAuthSubjectID].(string)
	if !ok {
		panic(fmt.Sprintf("interaction_adaptors: expect string subject ID claim, got %T", claims[interaction.IdentityClaimOAuthSubjectID]))
	}

	return
}

func extractOAuthProviderClaims(claims map[string]interface{}) oauth.ProviderID {
	provider, ok := claims[interaction.IdentityClaimOAuthProvider].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("interaction_adaptors: expect map provider claim, got %T", claims[interaction.IdentityClaimOAuthProvider]))
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

	return providerID
}

func extractAnonymousClaims(claims map[string]interface{}) (keyID string, key string) {
	if v, ok := claims[interaction.IdentityClaimAnonymousKeyID]; ok {
		if keyID, ok = v.(string); !ok {
			panic(fmt.Sprintf("interaction_adaptors: expect string key ID, got %T", claims[interaction.IdentityClaimAnonymousKeyID]))
		}
	}
	if v, ok := claims[interaction.IdentityClaimAnonymousKey]; ok {
		if key, ok = v.(string); !ok {
			panic(fmt.Sprintf("interaction_adaptors: expect string key, got %T", claims[interaction.IdentityClaimAnonymousKey]))
		}
	}
	return
}
