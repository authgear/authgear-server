package service

import (
	"context"
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
	New(userID string, loginID identity.LoginIDSpec, options loginid.CheckerOptions) (*identity.LoginID, error)
	WithValue(iden *identity.LoginID, value string, options loginid.CheckerOptions) (*identity.LoginID, error)
	Normalize(typ model.LoginIDKeyType, value string) (normalized string, uniqueKey string, err error)

	Get(ctx context.Context, userID, id string) (*identity.LoginID, error)
	GetMany(ctx context.Context, ids []string) ([]*identity.LoginID, error)
	List(ctx context.Context, userID string) ([]*identity.LoginID, error)
	GetByValue(ctx context.Context, loginIDValue string) ([]*identity.LoginID, error)
	GetByKeyAndValue(ctx context.Context, loginIDKey string, loginIDValue string) (*identity.LoginID, error)
	GetByUniqueKey(ctx context.Context, uniqueKey string) (*identity.LoginID, error)
	ListByClaim(ctx context.Context, name string, value string) ([]*identity.LoginID, error)
	Create(ctx context.Context, i *identity.LoginID) error
	Update(ctx context.Context, i *identity.LoginID) error
	Delete(ctx context.Context, i *identity.LoginID) error
}

type OAuthIdentityProvider interface {
	New(
		userID string,
		providerID oauthrelyingparty.ProviderID,
		subjectID string,
		profile map[string]interface{},
		claims map[string]interface{},
	) *identity.OAuth
	WithUpdate(iden *identity.OAuth, rawProfile map[string]interface{}, claims map[string]interface{}) *identity.OAuth

	Get(ctx context.Context, userID, id string) (*identity.OAuth, error)
	GetMany(ctx context.Context, ids []string) ([]*identity.OAuth, error)
	List(ctx context.Context, userID string) ([]*identity.OAuth, error)
	GetByProviderSubject(ctx context.Context, providerID oauthrelyingparty.ProviderID, subjectID string) (*identity.OAuth, error)
	GetByUserProvider(ctx context.Context, userID string, providerID oauthrelyingparty.ProviderID) (*identity.OAuth, error)
	ListByClaim(ctx context.Context, name string, value string) ([]*identity.OAuth, error)
	Create(ctx context.Context, i *identity.OAuth) error
	Update(ctx context.Context, i *identity.OAuth) error
	Delete(ctx context.Context, i *identity.OAuth) error
}

type AnonymousIdentityProvider interface {
	New(userID string, keyID string, key []byte) *identity.Anonymous

	Get(ctx context.Context, userID, id string) (*identity.Anonymous, error)
	GetMany(ctx context.Context, ids []string) ([]*identity.Anonymous, error)
	GetByKeyID(ctx context.Context, keyID string) (*identity.Anonymous, error)
	List(ctx context.Context, userID string) ([]*identity.Anonymous, error)
	Create(ctx context.Context, i *identity.Anonymous) error
	Delete(ctx context.Context, i *identity.Anonymous) error
}

type BiometricIdentityProvider interface {
	New(userID string, keyID string, key []byte, deviceInfo map[string]interface{}) *identity.Biometric

	Get(ctx context.Context, userID, id string) (*identity.Biometric, error)
	GetMany(ctx context.Context, ids []string) ([]*identity.Biometric, error)
	GetByKeyID(ctx context.Context, keyID string) (*identity.Biometric, error)
	List(ctx context.Context, userID string) ([]*identity.Biometric, error)
	Create(ctx context.Context, i *identity.Biometric) error
	Delete(ctx context.Context, i *identity.Biometric) error
}

type PasskeyIdentityProvider interface {
	New(ctx context.Context, userID string, attestationResponse []byte) (*identity.Passkey, error)
	Get(ctx context.Context, userID, id string) (*identity.Passkey, error)
	GetMany(ctx context.Context, ids []string) ([]*identity.Passkey, error)
	GetBySpec(ctx context.Context, spec *identity.PasskeySpec) (*identity.Passkey, error)
	List(ctx context.Context, userID string) ([]*identity.Passkey, error)
	Create(ctx context.Context, i *identity.Passkey) error
	Delete(ctx context.Context, i *identity.Passkey) error
}

type SIWEIdentityProvider interface {
	New(ctx context.Context, userID string, msg string, signature string) (*identity.SIWE, error)

	Get(ctx context.Context, userID, id string) (*identity.SIWE, error)
	GetMany(ctx context.Context, ids []string) ([]*identity.SIWE, error)
	GetByMessage(ctx context.Context, msg string, signature string) (*identity.SIWE, error)
	List(ctx context.Context, userID string) ([]*identity.SIWE, error)
	Create(ctx context.Context, i *identity.SIWE) error
	Delete(ctx context.Context, i *identity.SIWE) error
}

type LDAPIdentityProvider interface {
	New(
		userID string,
		serverName string,
		loginUserName *string,
		userIDAttributeName string,
		userIDAttributeValue []byte,
		claims map[string]interface{},
		rawEntryJSON map[string]interface{},
	) *identity.LDAP
	WithUpdate(iden *identity.LDAP, loginUserName *string, claims map[string]interface{}, rawEntryJSON map[string]interface{}) *identity.LDAP

	Get(ctx context.Context, userID, id string) (*identity.LDAP, error)
	GetMany(ctx context.Context, ids []string) ([]*identity.LDAP, error)
	List(ctx context.Context, userID string) ([]*identity.LDAP, error)
	GetByServerUserID(ctx context.Context, serverName string, userIDAttributeName string, userIDAttributeValue []byte) (*identity.LDAP, error)
	ListByClaim(ctx context.Context, name string, value string) ([]*identity.LDAP, error)
	Create(ctx context.Context, i *identity.LDAP) error
	Update(ctx context.Context, i *identity.LDAP) error
	Delete(ctx context.Context, i *identity.LDAP) error
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

func (s *Service) Get(ctx context.Context, id string) (*identity.Info, error) {
	ref, err := s.Store.GetRefByID(ctx, id)
	if err != nil {
		return nil, err
	}
	switch ref.Type {
	case model.IdentityTypeLoginID:
		l, err := s.LoginID.Get(ctx, ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return l.ToInfo(), nil
	case model.IdentityTypeOAuth:
		o, err := s.OAuth.Get(ctx, ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return o.ToInfo(), nil
	case model.IdentityTypeAnonymous:
		a, err := s.Anonymous.Get(ctx, ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return a.ToInfo(), nil
	case model.IdentityTypeBiometric:
		b, err := s.Biometric.Get(ctx, ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return b.ToInfo(), nil
	case model.IdentityTypePasskey:
		p, err := s.Passkey.Get(ctx, ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil
	case model.IdentityTypeSIWE:
		s, err := s.SIWE.Get(ctx, ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return s.ToInfo(), nil
	case model.IdentityTypeLDAP:
		s, err := s.LDAP.Get(ctx, ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return s.ToInfo(), nil
	}

	panic("identity: unknown identity type " + ref.Type)
}

func (s *Service) GetMany(ctx context.Context, ids []string) ([]*identity.Info, error) {
	refs, err := s.Store.ListRefsByIDs(ctx, ids)
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

	l, err := s.LoginID.GetMany(ctx, loginIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range l {
		infos = append(infos, i.ToInfo())
	}

	o, err := s.OAuth.GetMany(ctx, oauthIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range o {
		infos = append(infos, i.ToInfo())
	}

	a, err := s.Anonymous.GetMany(ctx, anonymousIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range a {
		infos = append(infos, i.ToInfo())
	}

	b, err := s.Biometric.GetMany(ctx, biometricIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range b {
		infos = append(infos, i.ToInfo())
	}

	p, err := s.Passkey.GetMany(ctx, passkeyIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range p {
		infos = append(infos, i.ToInfo())
	}

	e, err := s.SIWE.GetMany(ctx, siweIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range e {
		infos = append(infos, i.ToInfo())
	}

	ldapIdentities, err := s.LDAP.GetMany(ctx, ldapIDs)
	if err != nil {
		return nil, err
	}
	for _, i := range ldapIdentities {
		infos = append(infos, i.ToInfo())
	}

	return infos, nil
}

func (s *Service) getBySpec(ctx context.Context, spec *identity.Spec) (*identity.Info, error) {
	switch spec.Type {
	case model.IdentityTypeLoginID:
		loginID := spec.LoginID.Value.TrimSpace()
		l, err := s.LoginID.GetByValue(ctx, loginID)
		if err != nil {
			return nil, err
		} else if len(l) != 1 {
			return nil, api.ErrIdentityNotFound
		}
		return l[0].ToInfo(), nil
	case model.IdentityTypeOAuth:
		o, err := s.OAuth.GetByProviderSubject(ctx, spec.OAuth.ProviderID, spec.OAuth.SubjectID)
		if err != nil {
			return nil, err
		}
		return o.ToInfo(), nil
	case model.IdentityTypeAnonymous:
		keyID := spec.Anonymous.KeyID
		if keyID != "" {
			a, err := s.Anonymous.GetByKeyID(ctx, keyID)
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
		a, err := s.Anonymous.Get(ctx, userID, identityID)
		// identity must be found with existing user and identity id
		if err != nil {
			panic(fmt.Errorf("identity: failed to fetch anonymous identity: %s, %s, %w", userID, identityID, err))
		}
		return a.ToInfo(), nil
	case model.IdentityTypeBiometric:
		keyID := spec.Biometric.KeyID
		b, err := s.Biometric.GetByKeyID(ctx, keyID)
		if err != nil {
			return nil, err
		}
		return b.ToInfo(), nil
	case model.IdentityTypePasskey:
		p, err := s.Passkey.GetBySpec(ctx, spec.Passkey)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil
	case model.IdentityTypeSIWE:
		message := spec.SIWE.Message
		signature := spec.SIWE.Signature
		e, err := s.SIWE.GetByMessage(ctx, message, signature)
		if err != nil {
			return nil, err
		}
		return e.ToInfo(), nil
	case model.IdentityTypeLDAP:
		serverName := spec.LDAP.ServerName
		userIDAttributeName := spec.LDAP.UserIDAttributeName
		userIDAttributeValue := spec.LDAP.UserIDAttributeValue
		l, err := s.LDAP.GetByServerUserID(ctx, serverName, userIDAttributeName, userIDAttributeValue)
		if err != nil {
			return nil, err
		}
		return l.ToInfo(), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

// SearchBySpec does not return api.ErrIdentityNotFound.
func (s *Service) SearchBySpec(ctx context.Context, spec *identity.Spec) (exactMatch *identity.Info, otherMatches []*identity.Info, err error) {
	exactMatch, err = s.getBySpec(ctx, spec)
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
			loginIDs, err = s.LoginID.ListByClaim(ctx, name, str)
			if err != nil {
				return
			}

			for _, loginID := range loginIDs {
				otherMatches = append(otherMatches, loginID.ToInfo())
			}

			var oauths []*identity.OAuth
			oauths, err = s.OAuth.ListByClaim(ctx, name, str)
			if err != nil {
				return
			}

			for _, o := range oauths {
				otherMatches = append(otherMatches, o.ToInfo())
			}

			var ldaps []*identity.LDAP
			ldaps, err = s.LDAP.ListByClaim(ctx, name, str)
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
func (s *Service) ListByUserIDs(ctx context.Context, userIDs []string) (map[string][]*identity.Info, error) {
	refs, err := s.Store.ListRefsByUsers(ctx, userIDs, nil)
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
		loginIDs, err := s.LoginID.GetMany(ctx, extractIDs(loginIDRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range loginIDs {
			infos = append(infos, i.ToInfo())
		}
	}

	// oauth
	if oauthRefs, ok := refsByType[model.IdentityTypeOAuth]; ok && len(oauthRefs) > 0 {
		oauthIdens, err := s.OAuth.GetMany(ctx, extractIDs(oauthRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range oauthIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// anonymous
	if anonymousRefs, ok := refsByType[model.IdentityTypeAnonymous]; ok && len(anonymousRefs) > 0 {
		anonymousIdens, err := s.Anonymous.GetMany(ctx, extractIDs(anonymousRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range anonymousIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// biometric
	if biometricRefs, ok := refsByType[model.IdentityTypeBiometric]; ok && len(biometricRefs) > 0 {
		biometricIdens, err := s.Biometric.GetMany(ctx, extractIDs(biometricRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range biometricIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// passkey
	if passkeyRefs, ok := refsByType[model.IdentityTypePasskey]; ok && len(passkeyRefs) > 0 {
		passkeyIdens, err := s.Passkey.GetMany(ctx, extractIDs(passkeyRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range passkeyIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// siwe
	if siweRefs, ok := refsByType[model.IdentityTypeSIWE]; ok && len(siweRefs) > 0 {
		siweIdens, err := s.SIWE.GetMany(ctx, extractIDs(siweRefs))
		if err != nil {
			return nil, err
		}
		for _, i := range siweIdens {
			infos = append(infos, i.ToInfo())
		}
	}

	// ldap
	if ldapRefs, ok := refsByType[model.IdentityTypeLDAP]; ok && len(ldapRefs) > 0 {
		ldapIdens, err := s.LDAP.GetMany(ctx, extractIDs(ldapRefs))
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

func (s *Service) ListByUser(ctx context.Context, userID string) ([]*identity.Info, error) {
	infosByUserID, err := s.ListByUserIDs(ctx, []string{userID})
	if err != nil {
		return nil, err
	}

	infos, ok := infosByUserID[userID]

	if !ok || len(infos) == 0 {
		return []*identity.Info{}, nil
	}

	return infos, nil

}

func (s *Service) ListIdentitiesThatHaveStandardAttributes(ctx context.Context, userID string) ([]*identity.Info, error) {
	userIDs := []string{userID}

	extractIDs := func(idRefs []*model.IdentityRef) []string {
		ids := []string{}
		for _, idRef := range idRefs {
			ids = append(ids, idRef.ID)
		}
		return ids
	}

	infos := []*identity.Info{}

	{
		typeLoginID := model.IdentityTypeLoginID
		loginIDRefs, err := s.Store.ListRefsByUsers(ctx, userIDs, &typeLoginID)
		if err != nil {
			return nil, err
		}

		if len(loginIDRefs) > 0 {
			loginIDs, err := s.LoginID.GetMany(ctx, extractIDs(loginIDRefs))
			if err != nil {
				return nil, err
			}
			for _, i := range loginIDs {
				infos = append(infos, i.ToInfo())
			}
		}
	}

	{
		typeOAuth := model.IdentityTypeOAuth
		oauthRefs, err := s.Store.ListRefsByUsers(ctx, userIDs, &typeOAuth)
		if err != nil {
			return nil, err
		}

		if len(oauthRefs) > 0 {
			oauths, err := s.OAuth.GetMany(ctx, extractIDs(oauthRefs))
			if err != nil {
				return nil, err
			}
			for _, i := range oauths {
				infos = append(infos, i.ToInfo())
			}
		}
	}

	{
		typeLDAP := model.IdentityTypeLDAP
		ldapRefs, err := s.Store.ListRefsByUsers(ctx, userIDs, &typeLDAP)
		if err != nil {
			return nil, err
		}

		if len(ldapRefs) > 0 {
			ldaps, err := s.LDAP.GetMany(ctx, extractIDs(ldapRefs))
			if err != nil {
				return nil, err
			}
			for _, i := range ldaps {
				infos = append(infos, i.ToInfo())
			}
		}
	}

	return infos, nil
}

func (s *Service) Count(ctx context.Context, userID string) (uint64, error) {
	return s.Store.Count(ctx, userID)
}

func (s *Service) ListRefsByUsers(ctx context.Context, userIDs []string, identityType *model.IdentityType) ([]*model.IdentityRef, error) {
	return s.Store.ListRefsByUsers(ctx, userIDs, identityType)
}

func (s *Service) ListByClaim(ctx context.Context, name string, value string) ([]*identity.Info, error) {
	var infos []*identity.Info

	// login id
	lis, err := s.LoginID.ListByClaim(ctx, name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range lis {
		infos = append(infos, i.ToInfo())
	}

	// oauth
	ois, err := s.OAuth.ListByClaim(ctx, name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range ois {
		infos = append(infos, i.ToInfo())
	}

	// ldaps
	ldapIdentities, err := s.LDAP.ListByClaim(ctx, name, value)
	if err != nil {
		return nil, err
	}
	for _, i := range ldapIdentities {
		infos = append(infos, i.ToInfo())
	}

	return infos, nil
}

func (s *Service) New(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
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
		p, err := s.Passkey.New(ctx, userID, attestationResponse)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil
	case model.IdentityTypeSIWE:
		message := spec.SIWE.Message
		signature := spec.SIWE.Signature
		e, err := s.SIWE.New(ctx, userID, message, signature)
		if err != nil {
			return nil, err
		}
		return e.ToInfo(), nil
	case model.IdentityTypeLDAP:
		serverName := spec.LDAP.ServerName
		loginUserName := spec.LDAP.LastLoginUserName
		userIDAttributeName := spec.LDAP.UserIDAttributeName
		userIDAttributeValue := spec.LDAP.UserIDAttributeValue
		claims := spec.LDAP.Claims
		rawEntryJSON := spec.LDAP.RawEntryJSON
		l := s.LDAP.New(userID, serverName, loginUserName, userIDAttributeName, userIDAttributeValue, claims, rawEntryJSON)
		return l.ToInfo(), nil
	}

	panic("identity: unknown identity type " + spec.Type)
}

func (s *Service) Create(ctx context.Context, info *identity.Info) error {
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
		exactMatch, err := s.getBySpec(ctx, &incoming)
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
		sameProvider, err := s.OAuth.GetByUserProvider(ctx, info.UserID, info.OAuth.ProviderID)
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
		if err := s.LoginID.Create(ctx, i); err != nil {
			return err
		}
		*info = *i.ToInfo()

	case model.IdentityTypeOAuth:
		i := info.OAuth
		if err := s.OAuth.Create(ctx, i); err != nil {
			return err
		}
		*info = *i.ToInfo()

	case model.IdentityTypeAnonymous:
		i := info.Anonymous
		if err := s.Anonymous.Create(ctx, i); err != nil {
			return err
		}
		*info = *i.ToInfo()

	case model.IdentityTypeBiometric:
		i := info.Biometric
		if err := s.Biometric.Create(ctx, i); err != nil {
			return err
		}
		*info = *i.ToInfo()
	case model.IdentityTypePasskey:
		i := info.Passkey
		if err := s.Passkey.Create(ctx, i); err != nil {
			return err
		}
		*info = *i.ToInfo()
	case model.IdentityTypeSIWE:
		i := info.SIWE
		if err := s.SIWE.Create(ctx, i); err != nil {
			return err
		}
		*info = *i.ToInfo()
	case model.IdentityTypeLDAP:
		i := info.LDAP
		if err := s.LDAP.Create(ctx, i); err != nil {
			return err
		}
		*info = *i.ToInfo()
	default:
		panic("identity: unknown identity type " + info.Type)
	}
	return nil
}

func (s *Service) UpdateWithSpec(ctx context.Context, info *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error) {
	switch info.Type {
	case model.IdentityTypeLoginID:
		i, err := s.LoginID.WithValue(info.LoginID, spec.LoginID.Value.TrimSpace(), loginid.CheckerOptions{
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
		i := s.LDAP.WithUpdate(info.LDAP, spec.LDAP.LastLoginUserName, spec.LDAP.Claims, spec.LDAP.RawEntryJSON)
		return i.ToInfo(), nil
	default:
		panic("identity: cannot update identity type " + info.Type)
	}
}

func (s *Service) Update(ctx context.Context, info *identity.Info) error {
	switch info.Type {
	case model.IdentityTypeLoginID:
		i := info.LoginID
		if err := s.LoginID.Update(ctx, i); err != nil {
			return err
		}
		*info = *i.ToInfo()

	case model.IdentityTypeOAuth:
		i := info.OAuth
		if err := s.OAuth.Update(ctx, i); err != nil {
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
		if err := s.LDAP.Update(ctx, i); err != nil {
			return err
		}
		*info = *i.ToInfo()
	default:
		panic("identity: unknown identity type " + info.Type)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, info *identity.Info) error {
	switch info.Type {
	case model.IdentityTypeLoginID:
		i := info.LoginID
		if err := s.LoginID.Delete(ctx, i); err != nil {
			return err
		}
	case model.IdentityTypeOAuth:
		i := info.OAuth
		if err := s.OAuth.Delete(ctx, i); err != nil {
			return err
		}
	case model.IdentityTypeAnonymous:
		i := info.Anonymous
		if err := s.Anonymous.Delete(ctx, i); err != nil {
			return err
		}
	case model.IdentityTypeBiometric:
		i := info.Biometric
		if err := s.Biometric.Delete(ctx, i); err != nil {
			return err
		}
	case model.IdentityTypePasskey:
		i := info.Passkey
		if err := s.Passkey.Delete(ctx, i); err != nil {
			return err
		}
	case model.IdentityTypeSIWE:
		i := info.SIWE
		if err := s.SIWE.Delete(ctx, i); err != nil {
			return err
		}
	case model.IdentityTypeLDAP:
		i := info.LDAP
		if err := s.LDAP.Delete(ctx, i); err != nil {
			return err
		}
	default:
		panic("identity: unknown identity type " + info.Type)
	}

	return nil
}

func (s *Service) CheckDuplicated(ctx context.Context, info *identity.Info) (dupeIdentity *identity.Info, err error) {
	// There are two ways to check duplicate.
	// 1. Check duplicate by considering standard attributes.
	// 2. Check duplicate by considering type-specific unique key.
	// Only LoginID and OAuth has identity aware standard attributes and unique key.

	// 1. Check duplicate by considering standard attributes.
	claims := info.IdentityAwareStandardClaims()
	for name, value := range claims {
		var loginIDs []*identity.LoginID
		loginIDs, err = s.LoginID.ListByClaim(ctx, string(name), value)
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
		oauths, err = s.OAuth.ListByClaim(ctx, string(name), value)
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
		ldapIdentities, err = s.LDAP.ListByClaim(ctx, string(name), value)
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
	return s.CheckDuplicatedByUniqueKey(ctx, info)
}

func (s *Service) CheckDuplicatedByUniqueKey(ctx context.Context, info *identity.Info) (dupeIdentity *identity.Info, err error) {
	// Check duplicate by considering type-specific unique key.
	switch info.Type {
	case model.IdentityTypeLoginID:
		var i *identity.LoginID
		i, err = s.LoginID.GetByUniqueKey(ctx, info.LoginID.UniqueKey)
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
		o, err = s.OAuth.GetByProviderSubject(ctx, info.OAuth.ProviderID, info.OAuth.ProviderSubjectID)
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
		l, err = s.LDAP.GetByServerUserID(ctx, info.LDAP.ServerName, info.LDAP.UserIDAttributeName, info.LDAP.UserIDAttributeValue)
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

func (s *Service) ListCandidates(ctx context.Context, userID string) (out []identity.Candidate, err error) {
	var loginIDs []*identity.LoginID
	var oauths []*identity.OAuth
	var siwes []*identity.SIWE

	if userID != "" {
		loginIDs, err = s.LoginID.List(ctx, userID)
		if err != nil {
			return
		}
		oauths, err = s.OAuth.List(ctx, userID)
		if err != nil {
			return
		}
		// No need to consider anonymous identity
		// No need to consider biometric identity
		// No need to consider passkey identity

		siwes, err = s.SIWE.List(ctx, userID)
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
		if providerConfig.CreateDisabled() && !matched {
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
		if *loginIDKeyConfig.CreateDisabled && *loginIDKeyConfig.UpdateDisabled && !matched {
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

func (s *Service) Normalize(ctx context.Context, typ model.LoginIDKeyType, value string) (normalized string, uniqueKey string, err error) {
	return s.LoginID.Normalize(typ, value)
}

func (s *Service) AdminAPIGetByLoginIDKeyAndLoginIDValue(ctx context.Context, loginIDKey string, loginIDValue string) (*identity.Info, error) {
	loginID, err := s.LoginID.GetByKeyAndValue(ctx, loginIDKey, loginIDValue)
	if err != nil {
		return nil, err
	}
	return loginID.ToInfo(), nil
}

func (s *Service) AdminAPIGetByOAuthAliasAndSubject(ctx context.Context, alias string, subjectID string) (*identity.Info, error) {
	cfg, ok := s.Identity.OAuth.GetProviderConfig(alias)
	if !ok {
		return nil, api.ErrGetUsersInvalidArgument.New("invalid OAuth provider alias")
	}

	oauth, err := s.OAuth.GetByProviderSubject(ctx, cfg.ProviderID(), subjectID)
	if err != nil {
		return nil, err
	}

	return oauth.ToInfo(), nil
}
