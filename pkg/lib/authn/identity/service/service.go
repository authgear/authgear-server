package service

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package service

type LoginIDIdentityProvider interface {
	Get(userID, id string) (*identity.LoginID, error)
	GetMany(ids []string) ([]*identity.LoginID, error)
	List(userID string) ([]*identity.LoginID, error)
	GetByValue(loginIDValue string) ([]*identity.LoginID, error)
	ListByClaim(name string, value string) ([]*identity.LoginID, error)
	New(userID string, loginID identity.LoginIDSpec, options loginid.CheckerOptions) (*identity.LoginID, error)
	WithValue(iden *identity.LoginID, value string, options loginid.CheckerOptions) (*identity.LoginID, error)
	Create(i *identity.LoginID) error
	Update(i *identity.LoginID) error
	Delete(i *identity.LoginID) error
	CheckDuplicated(uniqueKey string, standardClaims map[model.ClaimName]string, userID string) (*identity.LoginID, error)
}

type OAuthIdentityProvider interface {
	Get(userID, id string) (*identity.OAuth, error)
	GetMany(ids []string) ([]*identity.OAuth, error)
	List(userID string) ([]*identity.OAuth, error)
	GetByProviderSubject(provider config.ProviderID, subjectID string) (*identity.OAuth, error)
	GetByUserProvider(userID string, provider config.ProviderID) (*identity.OAuth, error)
	ListByClaim(name string, value string) ([]*identity.OAuth, error)
	New(
		userID string,
		provider config.ProviderID,
		subjectID string,
		profile map[string]interface{},
		claims map[string]interface{},
	) *identity.OAuth
	WithUpdate(iden *identity.OAuth, rawProfile map[string]interface{}, claims map[string]interface{}) *identity.OAuth
	Create(i *identity.OAuth) error
	Update(i *identity.OAuth) error
	Delete(i *identity.OAuth) error
	CheckDuplicated(standardClaims map[model.ClaimName]string, userID string) (*identity.OAuth, error)
}

type AnonymousIdentityProvider interface {
	Get(userID, id string) (*identity.Anonymous, error)
	GetMany(ids []string) ([]*identity.Anonymous, error)
	GetByKeyID(keyID string) (*identity.Anonymous, error)
	List(userID string) ([]*identity.Anonymous, error)
	New(userID string, keyID string, key []byte) *identity.Anonymous
	Create(i *identity.Anonymous) error
	Delete(i *identity.Anonymous) error
}

type BiometricIdentityProvider interface {
	Get(userID, id string) (*identity.Biometric, error)
	GetMany(ids []string) ([]*identity.Biometric, error)
	GetByKeyID(keyID string) (*identity.Biometric, error)
	List(userID string) ([]*identity.Biometric, error)
	New(userID string, keyID string, key []byte, deviceInfo map[string]interface{}) *identity.Biometric
	Create(i *identity.Biometric) error
	Delete(i *identity.Biometric) error
}

type PasskeyIdentityProvider interface {
	Get(userID, id string) (*identity.Passkey, error)
	GetMany(ids []string) ([]*identity.Passkey, error)
	GetByAssertionResponse(assertionResponse []byte) (*identity.Passkey, error)
	List(userID string) ([]*identity.Passkey, error)
	New(userID string, attestationResponse []byte) (*identity.Passkey, error)
	Create(i *identity.Passkey) error
	Delete(i *identity.Passkey) error
}

type SIWEIdentityProvider interface {
	Get(userID, id string) (*identity.SIWE, error)
	GetMany(ids []string) ([]*identity.SIWE, error)
	GetByMessage(msg string) (*identity.SIWE, error)
	List(userID string) ([]*identity.SIWE, error)
	New(userID string, msg string, signature string) (*identity.SIWE, error)
	Create(i *identity.SIWE) error
	Delete(i *identity.SIWE) error
}

type Service struct {
	Authentication        *config.AuthenticationConfig
	Identity              *config.IdentityConfig
	IdentityFeatureConfig *config.IdentityFeatureConfig
	Store                 *Store
	LoginID               LoginIDIdentityProvider
	OAuth                 OAuthIdentityProvider
	Anonymous             AnonymousIdentityProvider
	Biometric             BiometricIdentityProvider
	Passkey               PasskeyIdentityProvider
	SIWE                  SIWEIdentityProvider
}

func (s *Service) Get(id string) (*identity.Info, error) {
	ref, err := s.Store.GetRefByID(id)
	if err != nil {
		return nil, err
	}
	switch ref.Type {
	case model.IdentityTypeLoginID:
		l, err := s.LoginID.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return l.ToInfo(), nil
	case model.IdentityTypeOAuth:
		o, err := s.OAuth.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return o.ToInfo(), nil
	case model.IdentityTypeAnonymous:
		a, err := s.Anonymous.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return a.ToInfo(), nil
	case model.IdentityTypeBiometric:
		b, err := s.Biometric.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return b.ToInfo(), nil
	case model.IdentityTypePasskey:
		p, err := s.Passkey.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil
	case model.IdentityTypeSIWE:
		s, err := s.SIWE.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return s.ToInfo(), nil
	}

	panic("identity: unknown identity type " + ref.Type)
}

func (s *Service) GetMany(ids []string) ([]*identity.Info, error) {
	refs, err := s.Store.ListRefsByIDs(ids)
	if err != nil {
		return nil, err
	}

	var loginIDs, oauthIDs, anonymousIDs, biometricIDs, passkeyIDs, siweIDs []string
	for _, ref := range refs {
		switch ref.Type {
		case model.IdentityTypeLoginID:
			loginIDs = append(loginIDs, ref.ID)
		case model.IdentityTypeOAuth:
			oauthIDs = append(oauthIDs, ref.ID)
		case model.IdentityTypeAnonymous:
			anonymousIDs = append(anonymousIDs, ref.ID)
		case model.IdentityTypeBiometric:
			biometricIDs = append(biometricIDs, ref.ID)
		case model.IdentityTypePasskey:
			passkeyIDs = append(passkeyIDs, ref.ID)
		case model.IdentityTypeSIWE:
			siweIDs = append(siweIDs, ref.ID)
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
		infos = append(infos, i.ToInfo())
	}

	o, err := s.OAuth.GetMany(oauthIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range o {
		infos = append(infos, i.ToInfo())
	}

	a, err := s.Anonymous.GetMany(anonymousIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range a {
		infos = append(infos, i.ToInfo())
	}

	b, err := s.Biometric.GetMany(biometricIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range b {
		infos = append(infos, i.ToInfo())
	}

	p, err := s.Passkey.GetMany(passkeyIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range p {
		infos = append(infos, i.ToInfo())
	}

	e, err := s.SIWE.GetMany(siweIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range e {
		infos = append(infos, i.ToInfo())
	}

	return infos, nil
}

func (s *Service) getBySpec(spec *identity.Spec) (*identity.Info, error) {
	switch spec.Type {
	case model.IdentityTypeLoginID:
		loginID := spec.LoginID.Value
		l, err := s.LoginID.GetByValue(loginID)
		if err != nil {
			return nil, err
		} else if len(l) != 1 {
			return nil, identity.ErrIdentityNotFound
		}
		return l[0].ToInfo(), nil
	case model.IdentityTypeOAuth:
		o, err := s.OAuth.GetByProviderSubject(spec.OAuth.ProviderID, spec.OAuth.SubjectID)
		if err != nil {
			return nil, err
		}
		return o.ToInfo(), nil
	case model.IdentityTypeAnonymous:
		keyID := spec.Anonymous.KeyID
		if keyID != "" {
			a, err := s.Anonymous.GetByKeyID(keyID)
			if err != nil {
				return nil, err
			}
			return a.ToInfo(), nil
		}
		// when keyID is empty, try to get the identity from user and identity id
		userID := spec.Anonymous.ExistingUserID
		identityID := spec.Anonymous.ExistingIdentityID
		if userID == "" {
			return nil, identity.ErrIdentityNotFound
		}
		a, err := s.Anonymous.Get(userID, identityID)
		// identity must be found with existing user and identity id
		if err != nil {
			panic(fmt.Errorf("identity: failed to fetch anonymous identity: %s, %s, %w", userID, identityID, err))
		}
		return a.ToInfo(), nil
	case model.IdentityTypeBiometric:
		keyID := spec.Biometric.KeyID
		b, err := s.Biometric.GetByKeyID(keyID)
		if err != nil {
			return nil, err
		}
		return b.ToInfo(), nil
	case model.IdentityTypePasskey:
		assertionResponse := spec.Passkey.AssertionResponse
		p, err := s.Passkey.GetByAssertionResponse(assertionResponse)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil
	case model.IdentityTypeSIWE:
		message := spec.SIWE.Message
		e, err := s.SIWE.GetByMessage(message)
		if err != nil {
			return nil, err
		}
		return e.ToInfo(), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

// SearchBySpec does not return identity.ErrIdentityNotFound.
func (s *Service) SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error) {
	exactMatch, err = s.getBySpec(spec)
	// The simplest case is the exact match case.
	if err == nil {
		return
	}

	// Any error other than identity.ErrIdentityNotFound
	if err != nil && !errors.Is(err, identity.ErrIdentityNotFound) {
		return
	}

	// Do not consider identity.ErrIdentityNotFound as error.
	err = nil

	claimsToSearch := make(map[string]interface{})

	// Otherwise we have to search.
	switch spec.Type {
	case model.IdentityTypeLoginID:
		// For login ID, we treat the login ID value as email, phone_number and preferred_username.
		loginID := spec.LoginID.Value
		claimsToSearch[string(model.ClaimEmail)] = loginID
		claimsToSearch[string(model.ClaimPhoneNumber)] = loginID
		claimsToSearch[string(model.ClaimPreferredUsername)] = loginID
	case model.IdentityTypeOAuth:
		if spec.OAuth.StandardClaims != nil {
			claimsToSearch = spec.OAuth.StandardClaims
		}
	default:
		break
	}

	for name, value := range claimsToSearch {
		str, ok := value.(string)
		if !ok {
			continue
		}
		switch name {
		case string(model.ClaimEmail),
			string(model.ClaimPhoneNumber),
			string(model.ClaimPreferredUsername):

			var loginIDs []*identity.LoginID
			loginIDs, err = s.LoginID.ListByClaim(name, str)
			if err != nil {
				return
			}

			for _, loginID := range loginIDs {
				otherMatches = append(otherMatches, loginID.ToInfo())
			}

			var oauths []*identity.OAuth
			oauths, err = s.OAuth.ListByClaim(name, str)
			if err != nil {
				return
			}

			for _, o := range oauths {
				otherMatches = append(otherMatches, o.ToInfo())
			}

		}
	}

	return
}

func (s *Service) ListByUser(userID string) ([]*identity.Info, error) {
	var infos []*identity.Info

	// login id
	lis, err := s.LoginID.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range lis {
		infos = append(infos, i.ToInfo())
	}

	// oauth
	ois, err := s.OAuth.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range ois {
		infos = append(infos, i.ToInfo())
	}

	// anonymous
	ais, err := s.Anonymous.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range ais {
		infos = append(infos, i.ToInfo())
	}

	// biometric
	bis, err := s.Biometric.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range bis {
		infos = append(infos, i.ToInfo())
	}

	// passkey
	pis, err := s.Passkey.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range pis {
		infos = append(infos, i.ToInfo())
	}

	// siwe
	sis, err := s.SIWE.List(userID)
	if err != nil {
		return nil, err
	}
	for _, i := range sis {
		infos = append(infos, i.ToInfo())
	}

	return infos, nil
}

func (s *Service) Count(userID string) (uint64, error) {
	return s.Store.Count(userID)
}

func (s *Service) ListRefsByUsers(userIDs []string) ([]*model.IdentityRef, error) {
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
		infos = append(infos, i.ToInfo())
	}

	// oauth
	ois, err := s.OAuth.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range ois {
		infos = append(infos, i.ToInfo())
	}

	return infos, nil
}

func (s *Service) New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	switch spec.Type {
	case model.IdentityTypeLoginID:
		l, err := s.LoginID.New(userID, *spec.LoginID, loginid.CheckerOptions{
			EmailByPassBlocklistAllowlist: options.LoginIDEmailByPassBlocklistAllowlist,
		})
		if err != nil {
			return nil, err
		}
		return l.ToInfo(), nil
	case model.IdentityTypeOAuth:
		providerID := spec.OAuth.ProviderID
		subjectID := spec.OAuth.SubjectID
		rawProfile := spec.OAuth.RawProfile
		standardClaims := spec.OAuth.StandardClaims
		o := s.OAuth.New(userID, providerID, subjectID, rawProfile, standardClaims)
		return o.ToInfo(), nil
	case model.IdentityTypeAnonymous:
		keyID := spec.Anonymous.KeyID
		key := spec.Anonymous.Key
		a := s.Anonymous.New(userID, keyID, []byte(key))
		return a.ToInfo(), nil
	case model.IdentityTypeBiometric:
		keyID := spec.Biometric.KeyID
		key := spec.Biometric.Key
		deviceInfo := spec.Biometric.DeviceInfo
		b := s.Biometric.New(userID, keyID, []byte(key), deviceInfo)
		return b.ToInfo(), nil
	case model.IdentityTypePasskey:
		attestationResponse := spec.Passkey.AttestationResponse
		p, err := s.Passkey.New(userID, attestationResponse)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil
	case model.IdentityTypeSIWE:
		message := spec.SIWE.Message
		signature := spec.SIWE.Signature
		e, err := s.SIWE.New(userID, message, signature)
		if err != nil {
			return nil, err
		}
		return e.ToInfo(), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

func (s *Service) Create(info *identity.Info) error {
	// TODO(verification): make OAuth verified according to config.
	switch info.Type {
	case model.IdentityTypeLoginID:
		i := info.LoginID
		if err := s.LoginID.Create(i); err != nil {
			return err
		}
		*info = *i.ToInfo()

	case model.IdentityTypeOAuth:
		i := info.OAuth
		if err := s.OAuth.Create(i); err != nil {
			return err
		}
		*info = *i.ToInfo()

	case model.IdentityTypeAnonymous:
		i := info.Anonymous
		if err := s.Anonymous.Create(i); err != nil {
			return err
		}
		*info = *i.ToInfo()

	case model.IdentityTypeBiometric:
		i := info.Biometric
		if err := s.Biometric.Create(i); err != nil {
			return err
		}
		*info = *i.ToInfo()
	case model.IdentityTypePasskey:
		i := info.Passkey
		if err := s.Passkey.Create(i); err != nil {
			return err
		}
		*info = *i.ToInfo()
	case model.IdentityTypeSIWE:
		i := info.SIWE
		if err := s.SIWE.Create(i); err != nil {
			return err
		}
		*info = *i.ToInfo()
	default:
		panic("identity: unknown identity type " + info.Type)
	}
	return nil
}

func (s *Service) UpdateWithSpec(info *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	switch info.Type {
	case model.IdentityTypeLoginID:
		i, err := s.LoginID.WithValue(info.LoginID, spec.LoginID.Value, loginid.CheckerOptions{
			EmailByPassBlocklistAllowlist: options.LoginIDEmailByPassBlocklistAllowlist,
		})
		if err != nil {
			return nil, err
		}
		return i.ToInfo(), nil
	case model.IdentityTypeOAuth:
		rawProfile := spec.OAuth.RawProfile
		standardClaims := spec.OAuth.StandardClaims
		i := s.OAuth.WithUpdate(
			info.OAuth,
			rawProfile,
			standardClaims,
		)
		return i.ToInfo(), nil
	default:
		panic("identity: cannot update identity type " + info.Type)
	}
}

func (s *Service) Update(info *identity.Info) error {
	switch info.Type {
	case model.IdentityTypeLoginID:
		i := info.LoginID
		if err := s.LoginID.Update(i); err != nil {
			return err
		}
		*info = *i.ToInfo()

	case model.IdentityTypeOAuth:
		i := info.OAuth
		if err := s.OAuth.Update(i); err != nil {
			return err
		}
		*info = *i.ToInfo()

	case model.IdentityTypeAnonymous:
		panic("identity: update no support for identity type " + info.Type)
	case model.IdentityTypeBiometric:
		panic("identity: update no support for identity type " + info.Type)
	case model.IdentityTypePasskey:
		panic("identity: update no support for identity type " + info.Type)
	case model.IdentityTypeSIWE:
		panic("identity: update no support for identity type " + info.Type)
	default:
		panic("identity: unknown identity type " + info.Type)
	}

	return nil
}

func (s *Service) Delete(info *identity.Info) error {
	switch info.Type {
	case model.IdentityTypeLoginID:
		i := info.LoginID
		if err := s.LoginID.Delete(i); err != nil {
			return err
		}
	case model.IdentityTypeOAuth:
		i := info.OAuth
		if err := s.OAuth.Delete(i); err != nil {
			return err
		}
	case model.IdentityTypeAnonymous:
		i := info.Anonymous
		if err := s.Anonymous.Delete(i); err != nil {
			return err
		}
	case model.IdentityTypeBiometric:
		i := info.Biometric
		if err := s.Biometric.Delete(i); err != nil {
			return err
		}
	case model.IdentityTypePasskey:
		i := info.Passkey
		if err := s.Passkey.Delete(i); err != nil {
			return err
		}
	case model.IdentityTypeSIWE:
		i := info.SIWE
		if err := s.SIWE.Delete(i); err != nil {
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
	if is.Type == model.IdentityTypeLoginID {
		li := is.LoginID
		loginIDUniqueKey = li.UniqueKey
	}

	// extract standard claims
	claims := is.IdentityAwareStandardClaims()

	li, err := s.LoginID.CheckDuplicated(loginIDUniqueKey, claims, is.UserID)
	if errors.Is(err, identity.ErrIdentityAlreadyExists) {
		dupeIdentity = li.ToInfo()
		return
	} else if err != nil {
		return
	}

	oi, err := s.OAuth.CheckDuplicated(claims, is.UserID)
	if errors.Is(err, identity.ErrIdentityAlreadyExists) {
		dupeIdentity = oi.ToInfo()
		return
	} else if err != nil {
		return
	}

	// No need to consider anonymous identity

	// No need to consider biometric identity

	// No need to consider passkey identity

	// No need to consider SIWE identity

	return
}

func (s *Service) ListCandidates(userID string) (out []identity.Candidate, err error) {
	var loginIDs []*identity.LoginID
	var oauths []*identity.OAuth
	var siwes []*identity.SIWE

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
		// No need to consider biometric identity
		// No need to consider passkey identity

		siwes, err = s.SIWE.List(userID)
		if err != nil {
			return
		}
	}

	for _, i := range s.Authentication.Identities {
		switch i {
		case model.IdentityTypeOAuth:
			for _, providerConfig := range s.Identity.OAuth.Providers {
				pc := providerConfig
				if identity.IsOAuthSSOProviderTypeDisabled(pc.Type, s.IdentityFeatureConfig.OAuth.Providers) {
					continue
				}
				configProviderID := pc.ProviderID()
				candidate := identity.NewOAuthCandidate(&pc)
				matched := false
				for _, iden := range oauths {
					if iden.ProviderID.Equal(&configProviderID) {
						matched = true
						candidate[identity.CandidateKeyIdentityID] = iden.ID
						candidate[identity.CandidateKeyProviderSubjectID] = string(iden.ProviderSubjectID)
						candidate[identity.CandidateKeyDisplayID] = iden.ToInfo().DisplayID()
					}
				}
				canAppend := true
				if *providerConfig.ModifyDisabled && !matched {
					canAppend = false
				}
				if canAppend {
					out = append(out, candidate)
				}
			}
		case model.IdentityTypeLoginID:
			for _, loginIDKeyConfig := range s.Identity.LoginID.Keys {
				lkc := loginIDKeyConfig
				candidate := identity.NewLoginIDCandidate(&lkc)
				matched := false
				for _, iden := range loginIDs {
					if loginIDKeyConfig.Key == iden.LoginIDKey {
						matched = true
						candidate[identity.CandidateKeyIdentityID] = iden.ID
						candidate[identity.CandidateKeyLoginIDValue] = iden.LoginID
						candidate[identity.CandidateKeyDisplayID] = iden.ToInfo().DisplayID()
					}
				}
				canAppend := true
				if *loginIDKeyConfig.ModifyDisabled && !matched {
					canAppend = false
				}
				if canAppend {
					out = append(out, candidate)
				}
			}
		case model.IdentityTypeSIWE:

			for _, iden := range siwes {
				candidate := identity.NewSIWECandidate()
				candidate[identity.CandidateKeyDisplayID] = iden.Address
				candidate[identity.CandidateKeyIdentityID] = iden.ID

				out = append(out, candidate)

			}
		}

	}

	return
}
