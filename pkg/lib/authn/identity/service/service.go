package service

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package service

type LoginIDIdentityProvider interface {
	Get(userID, id string) (*loginid.Identity, error)
	GetMany(ids []string) ([]*loginid.Identity, error)
	List(userID string) ([]*loginid.Identity, error)
	GetByValue(loginIDValue string) ([]*loginid.Identity, error)
	ListByClaim(name string, value string) ([]*loginid.Identity, error)
	New(userID string, loginID loginid.Spec) (*loginid.Identity, error)
	WithValue(iden *loginid.Identity, value string) (*loginid.Identity, error)
	Create(i *loginid.Identity) error
	Update(i *loginid.Identity) error
	Delete(i *loginid.Identity) error
	CheckDuplicated(uniqueKey string, standardClaims map[string]string, userID string) (*loginid.Identity, error)
}

type OAuthIdentityProvider interface {
	Get(userID, id string) (*oauth.Identity, error)
	GetMany(ids []string) ([]*oauth.Identity, error)
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
	CheckDuplicated(standardClaims map[string]string, userID string) (*oauth.Identity, error)
}

type AnonymousIdentityProvider interface {
	Get(userID, id string) (*anonymous.Identity, error)
	GetMany(ids []string) ([]*anonymous.Identity, error)
	GetByKeyID(keyID string) (*anonymous.Identity, error)
	List(userID string) ([]*anonymous.Identity, error)
	ListByClaim(name string, value string) ([]*anonymous.Identity, error)
	New(userID string, keyID string, key []byte) *anonymous.Identity
	Create(i *anonymous.Identity) error
	Delete(i *anonymous.Identity) error
}

type AuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	Delete(authenticatorInfo *authenticator.Info) error
}

type Service struct {
	Authentication *config.AuthenticationConfig
	Identity       *config.IdentityConfig
	Store          *Store
	LoginID        LoginIDIdentityProvider
	OAuth          OAuthIdentityProvider
	Anonymous      AnonymousIdentityProvider
	Authenticators AuthenticatorService
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

func (s *Service) GetMany(refs []*identity.Ref) ([]*identity.Info, error) {
	var loginIDs, oauthIDs, anonymousIDs []string
	for _, ref := range refs {
		switch ref.Type {
		case authn.IdentityTypeLoginID:
			loginIDs = append(loginIDs, ref.ID)
		case authn.IdentityTypeOAuth:
			oauthIDs = append(oauthIDs, ref.ID)
		case authn.IdentityTypeAnonymous:
			anonymousIDs = append(anonymousIDs, ref.ID)
		default:
			panic("identity: unknown identity type " + ref.Type)
		}
	}

	var infos []*identity.Info

	l, err := s.LoginID.GetMany(loginIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range l {
		infos = append(infos, loginIDToIdentityInfo(i))
	}

	o, err := s.OAuth.GetMany(oauthIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range o {
		infos = append(infos, s.toIdentityInfo(i))
	}

	a, err := s.Anonymous.GetMany(anonymousIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range a {
		infos = append(infos, anonymousToIdentityInfo(i))
	}

	return infos, nil
}

// GetBySpec return user ID and information about the identity that matches the provided spec.
func (s *Service) GetBySpec(spec *identity.Spec) (*identity.Info, error) {
	switch spec.Type {
	case authn.IdentityTypeLoginID:
		loginID := extractLoginIDValue(spec.Claims)
		l, err := s.LoginID.GetByValue(loginID)
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

func (s *Service) Count(userID string) (uint64, error) {
	return s.Store.Count(userID)
}

func (s *Service) ListRefsByUsers(userIDs []string) ([]*identity.Ref, error) {
	return s.Store.ListRefsByUsers(userIDs)
}

func (s *Service) ListByClaim(name string, value string) ([]*identity.Info, error) {
	var infos []*identity.Info

	// login id
	lis, err := s.LoginID.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range lis {
		infos = append(infos, loginIDToIdentityInfo(i))
	}

	// oauth
	ois, err := s.OAuth.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range ois {
		infos = append(infos, s.toIdentityInfo(i))
	}

	// anonymous
	ais, err := s.Anonymous.ListByClaim(name, value)
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
		loginID := extractLoginIDSpec(spec.Claims)
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
		i, err := s.LoginID.WithValue(loginIDFromIdentityInfo(info), extractLoginIDValue(spec.Claims))
		if err != nil {
			return nil, err
		}
		return loginIDToIdentityInfo(i), nil
	default:
		panic("identity: cannot update identity type " + info.Type)
	}

}

func (s *Service) Update(before, after *identity.Info) error {
	// FIXME(authenticator): delete only orphaned authenticators,
	//                       leave authenticators still having matching identities alone.
	ais, err := s.Authenticators.List(
		before.UserID,
		authenticator.KeepMatchingAuthenticatorOfIdentity(before),
	)
	if err != nil {
		return err
	}
	for _, ai := range ais {
		err := s.Authenticators.Delete(ai)
		if err != nil {
			return err
		}
	}

	switch after.Type {
	case authn.IdentityTypeLoginID:
		i := loginIDFromIdentityInfo(after)
		if err := s.LoginID.Update(i); err != nil {
			return err
		}
	case authn.IdentityTypeOAuth:
		i := oauthFromIdentityInfo(after)
		if err := s.OAuth.Update(i); err != nil {
			return err
		}
	case authn.IdentityTypeAnonymous:
		panic("identity: update no support for identity type " + after.Type)
	default:
		panic("identity: unknown identity type " + after.Type)
	}
	return nil
}

func (s *Service) Delete(info *identity.Info) error {
	// FIXME(authenticator): delete only orphaned authenticators,
	//                       leave authenticators still having matching identities alone.
	ais, err := s.Authenticators.List(
		info.UserID,
		authenticator.KeepMatchingAuthenticatorOfIdentity(info),
	)
	if err != nil {
		return err
	}
	for _, ai := range ais {
		err := s.Authenticators.Delete(ai)
		if err != nil {
			return err
		}
	}

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

func (s *Service) CheckDuplicated(is *identity.Info) (dupeIdentity *identity.Info, err error) {
	// extract login id unique key
	loginIDUniqueKey := ""
	if is.Type == authn.IdentityTypeLoginID {
		li := loginIDFromIdentityInfo(is)
		loginIDUniqueKey = li.UniqueKey
	}

	// extract standard claims
	claims := extractStandardClaims(is.Claims)

	li, err := s.LoginID.CheckDuplicated(loginIDUniqueKey, claims, is.UserID)
	if errors.Is(err, identity.ErrIdentityAlreadyExists) {
		dupeIdentity = loginIDToIdentityInfo(li)
		return
	} else if err != nil {
		return
	}

	oi, err := s.OAuth.CheckDuplicated(claims, is.UserID)
	if errors.Is(err, identity.ErrIdentityAlreadyExists) {
		dupeIdentity = s.toIdentityInfo(oi)
		return
	} else if err != nil {
		return
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
		ID:        o.ID,
		Labels:    o.Labels,
		UserID:    o.UserID,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		Type:      authn.IdentityTypeOAuth,
		Claims:    claims,
	}
}

func extractLoginIDValue(claims map[string]interface{}) string {
	loginID, ok := claims[identity.IdentityClaimLoginIDValue].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string login ID value, got %T", claims[identity.IdentityClaimLoginIDValue]))
	}

	return loginID
}

func extractLoginIDSpec(claims map[string]interface{}) loginid.Spec {
	loginIDKey, ok := claims[identity.IdentityClaimLoginIDKey].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string login ID key, got %T", claims[identity.IdentityClaimLoginIDKey]))
	}

	loginIDType, ok := claims[identity.IdentityClaimLoginIDType].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string login ID type, got %T", claims[identity.IdentityClaimLoginIDType]))
	}

	loginIDValue, ok := claims[identity.IdentityClaimLoginIDValue].(string)
	if !ok {
		panic(fmt.Sprintf("identity: expect string login ID value, got %T", claims[identity.IdentityClaimLoginIDValue]))
	}

	return loginid.Spec{
		Key:   loginIDKey,
		Type:  config.LoginIDKeyType(loginIDType),
		Value: loginIDValue,
	}
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
