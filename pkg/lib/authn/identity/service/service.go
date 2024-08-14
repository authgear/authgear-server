package service

import (
	"errors"
	"fmt"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api"
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
	GetByUniqueKey(uniqueKey string) (*identity.LoginID, error)
	ListByClaim(name string, value string) ([]*identity.LoginID, error)
	New(userID string, loginID identity.LoginIDSpec, options loginid.CheckerOptions) (*identity.LoginID, error)
	WithValue(iden *identity.LoginID, value string, options loginid.CheckerOptions) (*identity.LoginID, error)
	Create(i *identity.LoginID) error
	Update(i *identity.LoginID) error
	Delete(i *identity.LoginID) error
}

type OAuthIdentityProvider interface {
	Get(userID, id string) (*identity.OAuth, error)
	GetMany(ids []string) ([]*identity.OAuth, error)
	List(userID string) ([]*identity.OAuth, error)
	GetByProviderSubject(providerID oauthrelyingparty.ProviderID, subjectID string) (*identity.OAuth, error)
	GetByUserProvider(userID string, providerID oauthrelyingparty.ProviderID) (*identity.OAuth, error)
	ListByClaim(name string, value string) ([]*identity.OAuth, error)
	New(
		userID string,
		providerID oauthrelyingparty.ProviderID,
		subjectID string,
		profile map[string]interface{},
		claims map[string]interface{},
	) *identity.OAuth
	WithUpdate(iden *identity.OAuth, rawProfile map[string]interface{}, claims map[string]interface{}) *identity.OAuth
	Create(i *identity.OAuth) error
	Update(i *identity.OAuth) error
	Delete(i *identity.OAuth) error
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
	GetByMessage(msg string, signature string) (*identity.SIWE, error)
	List(userID string) ([]*identity.SIWE, error)
	New(userID string, msg string, signature string) (*identity.SIWE, error)
	Create(i *identity.SIWE) error
	Delete(i *identity.SIWE) error
}

type LDAPIdentityProvider interface {
	Get(userID, id string) (*identity.LDAP, error)
	GetMany(ids []string) ([]*identity.LDAP, error)
	List(userID string) ([]*identity.LDAP, error)
	GetByServerUserID(serverName string, userIDAttributeName string, userIDAttributeValue []byte) (*identity.LDAP, error)
	ListByClaim(name string, value string) ([]*identity.LDAP, error)
	New(
		userID string,
		serverName string,
		userIDAttributeName string,
		userIDAttributeValue []byte,
		claims map[string]interface{},
		rawEntryJSON map[string]interface{},
	) *identity.LDAP
	WithUpdate(iden *identity.LDAP, claims map[string]interface{}, rawEntryJSON map[string]interface{}) *identity.LDAP
	Create(i *identity.LDAP) error
	Update(i *identity.LDAP) error
	Delete(i *identity.LDAP) error
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
	LDAP                  LDAPIdentityProvider
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
	case model.IdentityTypeLDAP:
		s, err := s.LDAP.Get(ref.UserID, id)
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

	var loginIDs, oauthIDs, anonymousIDs, biometricIDs, passkeyIDs, siweIDs, ldapIDs []string
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
		case model.IdentityTypeLDAP:
			ldapIDs = append(ldapIDs, ref.ID)
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

	ldapIdentities, err := s.LDAP.GetMany(ldapIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range ldapIdentities {
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
			return nil, api.ErrIdentityNotFound
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
			return nil, api.ErrIdentityNotFound
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
		signature := spec.SIWE.Signature
		e, err := s.SIWE.GetByMessage(message, signature)
		if err != nil {
			return nil, err
		}
		return e.ToInfo(), nil
	case model.IdentityTypeLDAP:
		serverName := spec.LDAP.ServerName
		userIDAttributeName := spec.LDAP.UserIDAttributeName
		userIDAttributeValue := spec.LDAP.UserIDAttributeValue
		l, err := s.LDAP.GetByServerUserID(serverName, userIDAttributeName, userIDAttributeValue)
		if err != nil {
			return nil, err
		}
		return l.ToInfo(), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

// SearchBySpec does not return api.ErrIdentityNotFound.
func (s *Service) SearchBySpec(spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error) {
	exactMatch, err = s.getBySpec(spec)
	// The simplest case is the exact match case.
	if err == nil {
		return
	}

	// Any error other than api.ErrIdentityNotFound
	if !errors.Is(err, api.ErrIdentityNotFound) {
		return
	}

	// Do not consider api.ErrIdentityNotFound as error.
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
	case model.IdentityTypeLDAP:
		if spec.LDAP.Claims != nil {
			claimsToSearch = spec.LDAP.Claims
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

			var ldaps []*identity.LDAP
			ldaps, err = s.LDAP.ListByClaim(name, str)
			if err != nil {
				return
			}

			for _, l := range ldaps {
				otherMatches = append(otherMatches, l.ToInfo())
			}
		}
	}

	return
}

// nolint:gocognit
// This method is actually simple
func (s *Service) ListByUserIDs(userIDs []string) (map[string][]*identity.Info, error) {
	refs, err := s.Store.ListRefsByUsers(userIDs, nil)
	if err != nil {
		return nil, err
	}
	refsByType := map[model.IdentityType]([]*model.IdentityRef){}
	for _, ref := range refs {
		arr := refsByType[ref.Type]
		arr = append(arr, ref)
		refsByType[ref.Type] = arr
	}

	extractIDs := func(idRefs []*model.IdentityRef) []string {
		ids := []string{}
		for _, idRef := range idRefs {
			ids = append(ids, idRef.ID)
		}
		return ids
	}

	infos := []*identity.Info{}

	// login id
	if loginIDRefs, ok := refsByType[model.IdentityTypeLoginID]; ok && len(loginIDRefs) > 0 {
		loginIDs, err := s.LoginID.GetMany(extractIDs(loginIDRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range loginIDs {
			infos = append(infos, i.ToInfo())
		}
	}

	// oauth
	if oauthRefs, ok := refsByType[model.IdentityTypeOAuth]; ok && len(oauthRefs) > 0 {
		oauthIdens, err := s.OAuth.GetMany(extractIDs(oauthRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range oauthIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// anonymous
	if anonymousRefs, ok := refsByType[model.IdentityTypeAnonymous]; ok && len(anonymousRefs) > 0 {
		anonymousIdens, err := s.Anonymous.GetMany(extractIDs(anonymousRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range anonymousIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// biometric
	if biometricRefs, ok := refsByType[model.IdentityTypeBiometric]; ok && len(biometricRefs) > 0 {
		biometricIdens, err := s.Biometric.GetMany(extractIDs(biometricRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range biometricIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// passkey
	if passkeyRefs, ok := refsByType[model.IdentityTypePasskey]; ok && len(passkeyRefs) > 0 {
		passkeyIdens, err := s.Passkey.GetMany(extractIDs(passkeyRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range passkeyIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// siwe
	if siweRefs, ok := refsByType[model.IdentityTypeSIWE]; ok && len(siweRefs) > 0 {
		siweIdens, err := s.SIWE.GetMany(extractIDs(siweRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range siweIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// ldap
	if ldapRefs, ok := refsByType[model.IdentityTypeLDAP]; ok && len(ldapRefs) > 0 {
		ldapIdens, err := s.LDAP.GetMany(extractIDs(ldapRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range ldapIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	infosByUserID := map[string][]*identity.Info{}
	for _, info := range infos {
		arr := infosByUserID[info.UserID]
		arr = append(arr, info)
		infosByUserID[info.UserID] = arr
	}

	return infosByUserID, nil
}

func (s *Service) ListByUser(userID string) ([]*identity.Info, error) {
	infosByUserID, err := s.ListByUserIDs([]string{userID})
	if err != nil {
		return nil, err
	}

	infos, ok := infosByUserID[userID]

	if !ok || len(infos) == 0 {
		return []*identity.Info{}, nil
	}

	return infos, nil

}

func (s *Service) Count(userID string) (uint64, error) {
	return s.Store.Count(userID)
}

func (s *Service) ListRefsByUsers(userIDs []string, identityType *model.IdentityType) ([]*model.IdentityRef, error) {
	return s.Store.ListRefsByUsers(userIDs, identityType)
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

	// ldaps
	ldapIdentities, err := s.LDAP.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range ldapIdentities {
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
	case model.IdentityTypeLDAP:
		serverName := spec.LDAP.ServerName
		userIDAttributeName := spec.LDAP.UserIDAttributeName
		userIDAttributeValue := spec.LDAP.UserIDAttributeValue
		claims := spec.LDAP.Claims
		rawEntryJSON := spec.LDAP.RawEntryJSON
		l := s.LDAP.New(userID, serverName, userIDAttributeName, userIDAttributeValue, claims, rawEntryJSON)
		return l.ToInfo(), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

func (s *Service) Create(info *identity.Info) error {
	// DEV-1613: In https://github.com/authgear/authgear-server/pull/4462
	// We add checking of duplicated identity in Create().
	// The way we check duplicate is by turning a identity.Info into a identity.Spec
	// and then call getBySpec.
	// However, this is not compatible with anonymous identity spec.
	// The anonymous identity spec have different behavior based on which fields are present.
	// But calling ToSpec() will generate a anonymous identity spec with all fields set.
	// A anonymous identity spec with all fields set being passed to getBySpec() will confuse getBySpec() to panic.
	if info.Type != model.IdentityTypeAnonymous {
		incoming := info.ToSpec()
		exactMatch, err := s.getBySpec(&incoming)
		if errors.Is(err, api.ErrIdentityNotFound) {
			// nolint: ineffassign
			err = nil
		} else if err != nil {
			return err
		} else {
			existing := exactMatch.ToSpec()
			err = identity.NewErrDuplicatedIdentity(&incoming, &existing)
			return err
		}
	}

	// DEV-1664: For OAuth Identity, we additionally disallow
	// a user to have more than one identity of the same provider.
	if info.Type == model.IdentityTypeOAuth {
		sameProvider, err := s.OAuth.GetByUserProvider(info.UserID, info.OAuth.ProviderID)
		// Other errors
		if errors.Is(err, api.ErrIdentityNotFound) {
			// nolint: ineffassign
			err = nil
		} else if err != nil {
			return err
		} else {
			incoming := info.ToSpec()
			existing := sameProvider.ToInfo().ToSpec()
			err = identity.NewErrDuplicatedIdentity(&incoming, &existing)
			return err
		}
	}

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
	case model.IdentityTypeLDAP:
		i := info.LDAP
		if err := s.LDAP.Create(i); err != nil {
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
	case model.IdentityTypeLDAP:
		i := s.LDAP.WithUpdate(info.LDAP, spec.LDAP.Claims, spec.LDAP.RawEntryJSON)
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
	case model.IdentityTypeLDAP:
		i := info.LDAP
		if err := s.LDAP.Update(i); err != nil {
			return err
		}
		*info = *i.ToInfo()
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
	case model.IdentityTypeLDAP:
		i := info.LDAP
		if err := s.LDAP.Delete(i); err != nil {
			return err
		}
	default:
		panic("identity: unknown identity type " + info.Type)
	}

	return nil
}

func (s *Service) CheckDuplicated(info *identity.Info) (dupeIdentity *identity.Info, err error) {
	// There are two ways to check duplicate.
	// 1. Check duplicate by considering standard attributes.
	// 2. Check duplicate by considering type-specific unique key.
	// Only LoginID and OAuth has identity aware standard attributes and unique key.

	// 1. Check duplicate by considering standard attributes.
	claims := info.IdentityAwareStandardClaims()
	for name, value := range claims {
		var loginIDs []*identity.LoginID
		loginIDs, err = s.LoginID.ListByClaim(string(name), value)
		if err != nil {
			return nil, err
		}

		for _, i := range loginIDs {
			if i.UserID == info.UserID {
				continue
			}
			dupeIdentity = i.ToInfo()

			incoming := info.ToSpec()
			existing := dupeIdentity.ToSpec()
			err = identity.NewErrDuplicatedIdentity(&incoming, &existing)
			return
		}

		var oauths []*identity.OAuth
		oauths, err = s.OAuth.ListByClaim(string(name), value)
		if err != nil {
			return nil, err
		}

		for _, i := range oauths {
			if i.UserID == info.UserID {
				continue
			}
			dupeIdentity = i.ToInfo()

			incoming := info.ToSpec()
			existing := dupeIdentity.ToSpec()
			err = identity.NewErrDuplicatedIdentity(&incoming, &existing)
			return
		}

		var ldapIdentities []*identity.LDAP
		ldapIdentities, err = s.LDAP.ListByClaim(string(name), value)
		if err != nil {
			return nil, err
		}

		for _, i := range ldapIdentities {
			if i.UserID == info.UserID {
				continue
			}
			dupeIdentity = i.ToInfo()

			incoming := info.ToSpec()
			existing := dupeIdentity.ToSpec()
			err = identity.NewErrDuplicatedIdentity(&incoming, &existing)
			return
		}
	}

	// 2. Check duplicate by considering type-specific unique key.
	return s.CheckDuplicatedByUniqueKey(info)
}

func (s *Service) CheckDuplicatedByUniqueKey(info *identity.Info) (dupeIdentity *identity.Info, err error) {
	// Check duplicate by considering type-specific unique key.
	switch info.Type {
	case model.IdentityTypeLoginID:
		var i *identity.LoginID
		i, err = s.LoginID.GetByUniqueKey(info.LoginID.UniqueKey)
		if err != nil {
			if !errors.Is(err, api.ErrIdentityNotFound) {
				return
			}
			err = nil
		} else if i.UserID != info.UserID {
			dupeIdentity = i.ToInfo()

			incoming := info.ToSpec()
			existing := dupeIdentity.ToSpec()
			err = identity.NewErrDuplicatedIdentity(&incoming, &existing)
		}
	case model.IdentityTypeOAuth:
		var o *identity.OAuth
		o, err = s.OAuth.GetByProviderSubject(info.OAuth.ProviderID, info.OAuth.ProviderSubjectID)
		if err != nil {
			if !errors.Is(err, api.ErrIdentityNotFound) {
				return
			}
			err = nil
		} else if o.UserID != info.UserID {
			dupeIdentity = o.ToInfo()

			incoming := info.ToSpec()
			existing := dupeIdentity.ToSpec()
			err = identity.NewErrDuplicatedIdentity(&incoming, &existing)
		}
	case model.IdentityTypeLDAP:
		var l *identity.LDAP
		l, err = s.LDAP.GetByServerUserID(info.LDAP.ServerName, info.LDAP.UserIDAttributeName, info.LDAP.UserIDAttributeValue)
		if err != nil {
			if !errors.Is(err, api.ErrIdentityNotFound) {
				return
			}
			err = nil
		} else if l.UserID != info.UserID {
			dupeIdentity = l.ToInfo()

			incoming := info.ToSpec()
			existing := dupeIdentity.ToSpec()
			err = identity.NewErrDuplicatedIdentity(&incoming, &existing)
		}
	}

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
			out = append(out, s.listOAuthCandidates(oauths)...)
		case model.IdentityTypeLoginID:
			out = append(out, s.listLoginIDCandidates(loginIDs)...)
		case model.IdentityTypeSIWE:
			out = append(out, s.listSIWECandidates(siwes)...)
		case model.IdentityTypeLDAP:
			// TODO(DEV-1671): Support LDAP in settings page
			break
		}

	}

	return
}

func (s *Service) listOAuthCandidates(oauths []*identity.OAuth) []identity.Candidate {
	out := []identity.Candidate{}
	for _, providerConfig := range s.Identity.OAuth.Providers {
		pc := providerConfig
		if identity.IsOAuthSSOProviderTypeDisabled(pc.AsProviderConfig(), s.IdentityFeatureConfig.OAuth.Providers) {
			continue
		}
		configProviderID := pc.AsProviderConfig().ProviderID()
		candidate := identity.NewOAuthCandidate(pc)
		matched := false
		for _, iden := range oauths {
			if iden.ProviderID.Equal(configProviderID) {
				matched = true
				candidate[identity.CandidateKeyIdentityID] = iden.ID
				candidate[identity.CandidateKeyProviderSubjectID] = string(iden.ProviderSubjectID)
				candidate[identity.CandidateKeyDisplayID] = iden.ToInfo().DisplayID()
			}
		}
		canAppend := true
		if providerConfig.DeleteDisabled() && !matched {
			canAppend = false
		}
		if canAppend {
			out = append(out, candidate)
		}
	}
	return out
}

func (s *Service) listLoginIDCandidates(loginIDs []*identity.LoginID) []identity.Candidate {
	out := []identity.Candidate{}
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
		if *loginIDKeyConfig.DeleteDisabled && *loginIDKeyConfig.UpdateDisabled && !matched {
			canAppend = false
		}
		if canAppend {
			out = append(out, candidate)
		}
	}
	return out
}

func (s *Service) listSIWECandidates(siwes []*identity.SIWE) []identity.Candidate {
	out := []identity.Candidate{}
	for _, iden := range siwes {
		candidate := identity.NewSIWECandidate()
		candidate[identity.CandidateKeyDisplayID] = iden.Address.String()
		candidate[identity.CandidateKeyIdentityID] = iden.ID

		out = append(out, candidate)

	}
	return out
}
