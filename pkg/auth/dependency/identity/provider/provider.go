package provider

import (
	"fmt"
	"reflect"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

//go:generate mockgen -source=provider.go -destination=provider_mock_test.go -package provider

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
	List(userID string) ([]*anonymous.Identity, error)
	ListByClaim(name string, value string) ([]*anonymous.Identity, error)
	New(userID string, keyID string, key []byte) *anonymous.Identity
	Create(i *anonymous.Identity) error
	Delete(i *anonymous.Identity) error
}

type Provider struct {
	Authentication *config.AuthenticationConfiguration
	Identity       *config.IdentityConfiguration
	LoginID        LoginIDIdentityProvider
	OAuth          OAuthIdentityProvider
	Anonymous      AnonymousIdentityProvider
}

func (a *Provider) Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error) {
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
func (a *Provider) GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *identity.Info, error) {
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
func (a *Provider) GetByUserAndClaims(typ authn.IdentityType, userID string, claims map[string]interface{}) (*identity.Info, error) {
	switch typ {
	case authn.IdentityTypeOAuth:
		providerID := extractOAuthProviderClaims(claims)
		o, err := a.OAuth.GetByUserProvider(userID, providerID)
		if err != nil {
			return nil, err
		}
		return oauthToIdentityInfo(o), nil
	case authn.IdentityTypeAnonymous:
		as, err := a.Anonymous.List(userID)
		if err != nil {
			return nil, err
		} else if len(as) == 0 {
			return nil, identity.ErrIdentityNotFound
		}
		return anonymousToIdentityInfo(as[0]), nil
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
func (a *Provider) ListByClaims(claims map[string]string) ([]*identity.Info, error) {
	var all []*identity.Info

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

func (a *Provider) ListByUser(userID string) ([]*identity.Info, error) {
	iis := []*identity.Info{}

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

	// anonymous
	ais, err := a.Anonymous.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range ais {
		iis = append(iis, anonymousToIdentityInfo(i))
	}

	return iis, nil
}

func (a *Provider) New(userID string, typ authn.IdentityType, claims map[string]interface{}) *identity.Info {
	switch typ {
	case authn.IdentityTypeLoginID:
		loginID := extractLoginIDClaims(claims)
		l := a.LoginID.New(userID, loginID)
		return loginIDToIdentityInfo(l)

	case authn.IdentityTypeOAuth:
		providerID, subjectID := extractOAuthClaims(claims)
		var profile, oidcClaims map[string]interface{}
		var ok bool
		if profile, ok = claims[identity.IdentityClaimOAuthProfile].(map[string]interface{}); !ok {
			profile = map[string]interface{}{}
		}
		if oidcClaims, ok = claims[identity.IdentityClaimOAuthClaims].(map[string]interface{}); !ok {
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

func (a *Provider) WithClaims(userID string, ii *identity.Info, claims map[string]interface{}) *identity.Info {
	switch ii.Type {
	case authn.IdentityTypeLoginID:
		oldIden := loginIDFromIdentityInfo(userID, ii)
		newLoginID := extractLoginIDClaims(claims)
		newIden := a.LoginID.WithLoginID(oldIden, newLoginID)
		return loginIDToIdentityInfo(newIden)

	case authn.IdentityTypeOAuth:
		var profile map[string]interface{}
		var ok bool
		if profile, ok = claims[identity.IdentityClaimOAuthProfile].(map[string]interface{}); !ok {
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

func (a *Provider) CreateAll(userID string, is []*identity.Info) error {
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

func (a *Provider) UpdateAll(userID string, is []*identity.Info) error {
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

func (a *Provider) DeleteAll(userID string, is []*identity.Info) error {
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
		case authn.IdentityTypeAnonymous:
			identity := anonymousFromIdentityInfo(userID, i)
			if err := a.Anonymous.Delete(identity); err != nil {
				return err
			}
		default:
			panic("interaction_adaptors: unknown identity type " + i.Type)
		}
	}
	return nil
}

func (a *Provider) Validate(is []*identity.Info) error {
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

func (a *Provider) RelateIdentityToAuthenticator(is identity.Spec, as *authenticator.Spec) *authenticator.Spec {
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
			as.Props[authenticator.AuthenticatorPropOOBOTPChannelType] = string(authn.AuthenticatorOOBChannelEmail)
			as.Props[authenticator.AuthenticatorPropOOBOTPEmail] = loginID.Value
			return as
		case string(metadata.Phone):
			as.Props[authenticator.AuthenticatorPropOOBOTPChannelType] = string(authn.AuthenticatorOOBChannelSMS)
			as.Props[authenticator.AuthenticatorPropOOBOTPPhone] = loginID.Value
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

func (a *Provider) ListCandidates(userID string) (out []identity.Candidate, err error) {
	var loginIDs []*loginid.Identity
	var oauths []*oauth.Identity

	if userID != "" {
		loginIDs, err = a.LoginID.List(userID)
		if err != nil {
			return
		}
		oauths, err = a.OAuth.List(userID)
		if err != nil {
			return
		}
		// No need to consider anonymous identity
	}

	for _, i := range a.Authentication.Identities {
		switch i {
		case string(authn.IdentityTypeOAuth):
			for _, providerConfig := range a.Identity.OAuth.Providers {
				configProviderID := oauth.NewProviderID(providerConfig)
				candidate := identity.NewOAuthCandidate(&providerConfig)
				for _, iden := range oauths {
					if iden.ProviderID.Equal(&configProviderID) {
						candidate[identity.CandidateKeyProviderSubjectID] = string(iden.ProviderSubjectID)
						if email, ok := iden.Claims["email"].(string); ok {
							candidate[identity.CandidateKeyEmail] = email
						}
					}
				}
				out = append(out, candidate)
			}
		case string(authn.IdentityTypeLoginID):
			for _, loginIDKeyConfig := range a.Identity.LoginID.Keys {
				candidate := identity.NewLoginIDCandidate(&loginIDKeyConfig)
				for _, iden := range loginIDs {
					if loginIDKeyConfig.Key == iden.LoginIDKey {
						candidate[identity.CandidateKeyLoginIDValue] = iden.LoginID
						if email, ok := iden.Claims["email"]; ok {
							candidate[identity.CandidateKeyEmail] = email
						}
					}
				}
				out = append(out, candidate)
			}
		}
	}

	return
}

func extractLoginIDClaims(claims map[string]interface{}) loginid.LoginID {
	loginIDKey := ""
	if v, ok := claims[identity.IdentityClaimLoginIDKey]; ok {
		if loginIDKey, ok = v.(string); !ok {
			panic(fmt.Sprintf("interaction_adaptors: expect string login ID key, got %T", claims[identity.IdentityClaimLoginIDKey]))
		}
	}
	loginID, ok := claims[identity.IdentityClaimLoginIDValue].(string)
	if !ok {
		panic(fmt.Sprintf("interaction_adaptors: expect string login ID value, got %T", claims[identity.IdentityClaimLoginIDValue]))
	}

	return loginid.LoginID{Key: loginIDKey, Value: loginID}
}

func extractOAuthClaims(claims map[string]interface{}) (providerID oauth.ProviderID, subjectID string) {
	providerID = extractOAuthProviderClaims(claims)

	subjectID, ok := claims[identity.IdentityClaimOAuthSubjectID].(string)
	if !ok {
		panic(fmt.Sprintf("interaction_adaptors: expect string subject ID claim, got %T", claims[identity.IdentityClaimOAuthSubjectID]))
	}

	return
}

func extractOAuthProviderClaims(claims map[string]interface{}) oauth.ProviderID {
	provider, ok := claims[identity.IdentityClaimOAuthProvider].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("interaction_adaptors: expect map provider claim, got %T", claims[identity.IdentityClaimOAuthProvider]))
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
	if v, ok := claims[identity.IdentityClaimAnonymousKeyID]; ok {
		if keyID, ok = v.(string); !ok {
			panic(fmt.Sprintf("interaction_adaptors: expect string key ID, got %T", claims[identity.IdentityClaimAnonymousKeyID]))
		}
	}
	if v, ok := claims[identity.IdentityClaimAnonymousKey]; ok {
		if key, ok = v.(string); !ok {
			panic(fmt.Sprintf("interaction_adaptors: expect string key, got %T", claims[identity.IdentityClaimAnonymousKey]))
		}
	}
	return
}
