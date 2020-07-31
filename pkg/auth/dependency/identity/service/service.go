package service

import (
	"fmt"
	"reflect"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/oauth"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package service

type LoginIDIdentityProvider interface {
	Get(userID, id string) (*loginid.Identity, error)
	List(userID string) ([]*loginid.Identity, error)
	GetByLoginID(loginID loginid.LoginID) ([]*loginid.Identity, error)
	ListByClaim(name string, value string) ([]*loginid.Identity, error)
	New(userID string, loginID loginid.LoginID) (*loginid.Identity, error)
	WithLoginID(iden *loginid.Identity, loginID loginid.LoginID) (*loginid.Identity, error)
	Create(i *loginid.Identity) error
	Update(i *loginid.Identity) error
	Delete(i *loginid.Identity) error
	Validate(loginIDs []loginid.LoginID) error
	Normalize(loginID loginid.LoginID) (*loginid.LoginID, *config.LoginIDKeyConfig, string, error)
	CheckDuplicated(uniqueKey string, standardClaims map[string]string, userID string) error
}

type OAuthIdentityProvider interface {
	Get(userID, id string) (*oauth.Identity, error)
	List(userID string) ([]*oauth.Identity, error)
	GetByProviderSubject(provider config.ProviderID, subjectID string) (*oauth.Identity, error)
	GetByUserProvider(userID string, provider config.ProviderID) (*oauth.Identity, error)
	ListByClaim(name string, value string) ([]*oauth.Identity, error)
	New(
		userID string,
		provider config.ProviderID,
		subjectID string,
		profile map[string]interface{},
		claims map[string]interface{},
	) *oauth.Identity
	Create(i *oauth.Identity) error
	Update(i *oauth.Identity) error
	Delete(i *oauth.Identity) error
	CheckDuplicated(standardClaims map[string]string, userID string) error
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

type Service struct {
	Authentication *config.AuthenticationConfig
	Identity       *config.IdentityConfig
	LoginID        LoginIDIdentityProvider
	OAuth          OAuthIdentityProvider
	Anonymous      AnonymousIdentityProvider
}

func (s *Service) Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error) {
	switch typ {
	case authn.IdentityTypeLoginID:
		l, err := s.LoginID.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return loginIDToIdentityInfo(l), nil

	case authn.IdentityTypeOAuth:
		o, err := s.OAuth.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return s.toIdentityInfo(o), nil

	case authn.IdentityTypeAnonymous:
		a, err := s.Anonymous.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return anonymousToIdentityInfo(a), nil
	}

	panic("identity: unknown identity type " + typ)
}

// GetBySpec return user ID and information about the identity that matches the provided spec.
func (s *Service) GetBySpec(spec *identity.Spec) (*identity.Info, error) {
	switch spec.Type {
	case authn.IdentityTypeLoginID:
		loginID := extractLoginIDClaims(spec.Claims)
		l, err := s.LoginID.GetByLoginID(loginID)
		if err != nil {
			return nil, err
		} else if len(l) != 1 {
			return nil, identity.ErrIdentityNotFound
		}
		return loginIDToIdentityInfo(l[0]), nil

	case authn.IdentityTypeOAuth:
		providerID, subjectID := extractOAuthClaims(spec.Claims)
		o, err := s.OAuth.GetByProviderSubject(providerID, subjectID)
		if err != nil {
			return nil, err
		}
		return s.toIdentityInfo(o), nil

	case authn.IdentityTypeAnonymous:
		keyID, _ := extractAnonymousClaims(spec.Claims)
		a, err := s.Anonymous.GetByKeyID(keyID)
		if err != nil {
			return nil, err
		}
		return anonymousToIdentityInfo(a), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

func (s *Service) ListByUser(userID string) ([]*identity.Info, error) {
	var infos []*identity.Info

	// login id
	lis, err := s.LoginID.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range lis {
		infos = append(infos, loginIDToIdentityInfo(i))
	}

	// oauth
	ois, err := s.OAuth.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range ois {
		infos = append(infos, s.toIdentityInfo(i))
	}

	// anonymous
	ais, err := s.Anonymous.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range ais {
		infos = append(infos, anonymousToIdentityInfo(i))
	}

	return infos, nil
}

func (s *Service) New(userID string, spec *identity.Spec) (*identity.Info, error) {
	switch spec.Type {
	case authn.IdentityTypeLoginID:
		loginID := extractLoginIDClaims(spec.Claims)
		l, err := s.LoginID.New(userID, loginID)
		if err != nil {
			return nil, err
		}
		return loginIDToIdentityInfo(l), nil
	case authn.IdentityTypeOAuth:
		providerID, subjectID := extractOAuthClaims(spec.Claims)
		var profile, oidcClaims map[string]interface{}
		var ok bool
		if profile, ok = spec.Claims[identity.IdentityClaimOAuthProfile].(map[string]interface{}); !ok {
			profile = map[string]interface{}{}
		}
		if oidcClaims, ok = spec.Claims[identity.IdentityClaimOAuthClaims].(map[string]interface{}); !ok {
			oidcClaims = map[string]interface{}{}
		}

		o := s.OAuth.New(userID, providerID, subjectID, profile, oidcClaims)
		return s.toIdentityInfo(o), nil
	case authn.IdentityTypeAnonymous:
		keyID, key := extractAnonymousClaims(spec.Claims)
		a := s.Anonymous.New(userID, keyID, []byte(key))
		return anonymousToIdentityInfo(a), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

func (s *Service) Create(info *identity.Info) error {
	switch info.Type {
	case authn.IdentityTypeLoginID:
		i := loginIDFromIdentityInfo(info)
		if err := s.LoginID.Create(i); err != nil {
			return err
		}

	case authn.IdentityTypeOAuth:
		i := oauthFromIdentityInfo(info)
		if err := s.OAuth.Create(i); err != nil {
			return err
		}

	case authn.IdentityTypeAnonymous:
		i := anonymousFromIdentityInfo(info)
		if err := s.Anonymous.Create(i); err != nil {
			return err
		}

	default:
		panic("identity: unknown identity type " + info.Type)
	}
	return nil
}

func (s *Service) UpdateWithSpec(info *identity.Info, spec *identity.Spec) (*identity.Info, error) {
	switch info.Type {
	case authn.IdentityTypeLoginID:
		i, err := s.LoginID.WithLoginID(loginIDFromIdentityInfo(info), extractLoginIDClaims(spec.Claims))
		if err != nil {
			return nil, err
		}
		return loginIDToIdentityInfo(i), nil
	default:
		panic("identity: cannot update identity type " + info.Type)
	}

}

func (s *Service) Update(info *identity.Info) error {
	switch info.Type {
	case authn.IdentityTypeLoginID:
		i := loginIDFromIdentityInfo(info)
		if err := s.LoginID.Update(i); err != nil {
			return err
		}
	case authn.IdentityTypeOAuth:
		i := oauthFromIdentityInfo(info)
		if err := s.OAuth.Update(i); err != nil {
			return err
		}
	case authn.IdentityTypeAnonymous:
		panic("identity: update no support for identity type " + info.Type)
	default:
		panic("identity: unknown identity type " + info.Type)
	}
	return nil
}

func (s *Service) Delete(info *identity.Info) error {
	switch info.Type {
	case authn.IdentityTypeLoginID:
		i := loginIDFromIdentityInfo(info)
		if err := s.LoginID.Delete(i); err != nil {
			return err
		}
	case authn.IdentityTypeOAuth:
		i := oauthFromIdentityInfo(info)
		if err := s.OAuth.Delete(i); err != nil {
			return err
		}
	case authn.IdentityTypeAnonymous:
		i := anonymousFromIdentityInfo(info)
		if err := s.Anonymous.Delete(i); err != nil {
			return err
		}
	default:
		panic("identity: unknown identity type " + info.Type)
	}
	return nil
}

func (s *Service) Validate(is []*identity.Info) error {
	var loginIDs []loginid.LoginID
	var oauthProviderIDs []config.ProviderID
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
		if err := s.LoginID.Validate(loginIDs); err != nil {
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

func (s *Service) CheckDuplicated(is *identity.Info) (err error) {
	// extract login id unique key
	loginIDUniqueKey := ""
	if is.Type == authn.IdentityTypeLoginID {
		li := loginIDFromIdentityInfo(is)
		loginIDUniqueKey = li.UniqueKey
	}

	// extract standard claims
	claims := extractStandardClaims(is.Claims)

	err = s.LoginID.CheckDuplicated(loginIDUniqueKey, claims, is.UserID)
	if err != nil {
		return err
	}

	err = s.OAuth.CheckDuplicated(claims, is.UserID)
	if err != nil {
		return err
	}

	// No need to consider anonymous identity

	return
}

func (s *Service) ListCandidates(userID string) (out []identity.Candidate, err error) {
	var loginIDs []*loginid.Identity
	var oauths []*oauth.Identity

	if userID != "" {
		loginIDs, err = s.LoginID.List(userID)
		if err != nil {
			return
		}
		oauths, err = s.OAuth.List(userID)
		if err != nil {
			return
		}
		// No need to consider anonymous identity
	}

	for _, i := range s.Authentication.Identities {
		switch i {
		case authn.IdentityTypeOAuth:
			for _, providerConfig := range s.Identity.OAuth.Providers {
				pc := providerConfig
				configProviderID := pc.ProviderID()
				candidate := identity.NewOAuthCandidate(&pc)
				for _, iden := range oauths {
					if iden.ProviderID.Equal(&configProviderID) {
						candidate[identity.CandidateKeyIdentityID] = iden.ID
						candidate[identity.CandidateKeyProviderSubjectID] = string(iden.ProviderSubjectID)
						candidate[identity.CandidateKeyDisplayID] = s.toIdentityInfo(iden).DisplayID()
					}
				}
				out = append(out, candidate)
			}
		case authn.IdentityTypeLoginID:
			for _, loginIDKeyConfig := range s.Identity.LoginID.Keys {
				lkc := loginIDKeyConfig
				candidate := identity.NewLoginIDCandidate(&lkc)
				for _, iden := range loginIDs {
					if loginIDKeyConfig.Key == iden.LoginIDKey {
						candidate[identity.CandidateKeyIdentityID] = iden.ID
						candidate[identity.CandidateKeyLoginIDValue] = iden.LoginID
						candidate[identity.CandidateKeyDisplayID] = loginIDToIdentityInfo(iden).DisplayID()
					}
				}
				out = append(out, candidate)
			}
		}
	}

	return
}

func (s *Service) toIdentityInfo(o *oauth.Identity) *identity.Info {
	provider := map[string]interface{}{
		"type": o.ProviderID.Type,
	}
	for k, v := range o.ProviderID.Keys {
		provider[k] = v
	}

	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProviderKeys: provider,
		identity.IdentityClaimOAuthProviderType: o.ProviderID.Type,
		identity.IdentityClaimOAuthSubjectID:    o.ProviderSubjectID,
		identity.IdentityClaimOAuthProfile:      o.UserProfile,
	}

	alias := ""
	for _, providerConfig := range s.Identity.OAuth.Providers {
		providerID := providerConfig.ProviderID()
		if providerID.Equal(&o.ProviderID) {
			alias = providerConfig.Alias
		}
	}
	if alias != "" {
		claims[identity.IdentityClaimOAuthProviderAlias] = alias
	}

	for k, v := range o.Claims {
		claims[k] = v
	}

	return &identity.Info{
		UserID:   o.UserID,
		Type:     authn.IdentityTypeOAuth,
		ID:       o.ID,
		Claims:   claims,
		Identity: o,
	}
}

func extractLoginIDClaims(claims map[string]interface{}) loginid.LoginID {
	loginIDKey := ""
	if v, ok := claims[identity.IdentityClaimLoginIDKey]; ok {
		if loginIDKey, ok = v.(string); !ok {
			panic(fmt.Sprintf("identity: expect string login ID key, got %T", claims[identity.IdentityClaimLoginIDKey]))
		}
	}
	loginID, ok := claims[identity.IdentityClaimLoginIDValue].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string login ID value, got %T", claims[identity.IdentityClaimLoginIDValue]))
	}

	return loginid.LoginID{Key: loginIDKey, Value: loginID}
}

func extractOAuthClaims(claims map[string]interface{}) (providerID config.ProviderID, subjectID string) {
	providerID = extractOAuthProviderClaims(claims)

	subjectID, ok := claims[identity.IdentityClaimOAuthSubjectID].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string subject ID claim, got %T", claims[identity.IdentityClaimOAuthSubjectID]))
	}

	return
}

func extractOAuthProviderClaims(claims map[string]interface{}) config.ProviderID {
	provider, ok := claims[identity.IdentityClaimOAuthProviderKeys].(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("identity: expect map provider claim, got %T", claims[identity.IdentityClaimOAuthProviderKeys]))
	}

	providerID := config.ProviderID{Keys: map[string]interface{}{}}
	for k, v := range provider {
		if k == "type" {
			providerID.Type, ok = v.(string)
			if !ok {
				panic(fmt.Sprintf("identity: expect string provider type, got %T", v))
			}
		} else {
			providerID.Keys[k] = v.(string)
		}
	}

	return providerID
}

func extractAnonymousClaims(claims map[string]interface{}) (keyID string, key string) {
	if v, ok := claims[identity.IdentityClaimAnonymousKeyID]; ok {
		if keyID, ok = v.(string); !ok {
			panic(fmt.Sprintf("identity: expect string key ID, got %T", claims[identity.IdentityClaimAnonymousKeyID]))
		}
	}
	if v, ok := claims[identity.IdentityClaimAnonymousKey]; ok {
		if key, ok = v.(string); !ok {
			panic(fmt.Sprintf("identity: expect string key, got %T", claims[identity.IdentityClaimAnonymousKey]))
		}
	}
	return
}

func extractStandardClaims(claims map[string]interface{}) map[string]string {
	standardClaims := map[string]string{}
	email, hasEmail := claims[identity.StandardClaimEmail].(string)
	if hasEmail {
		standardClaims[identity.StandardClaimEmail] = email
	}

	return standardClaims
}
