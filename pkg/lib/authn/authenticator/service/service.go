package service

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type PasswordAuthenticatorProvider interface {
	Get(userID, id string) (*authenticator.Password, error)
	GetMany(ids []string) ([]*authenticator.Password, error)
	List(userID string) ([]*authenticator.Password, error)
	New(id string, userID string, password string, isDefault bool, kind string) (*authenticator.Password, error)
	// WithPassword returns new authenticator pointer if password is changed
	// Otherwise original authenticator will be returned
	WithPassword(a *authenticator.Password, password string) (*authenticator.Password, error)
	Create(*authenticator.Password) error
	UpdatePassword(*authenticator.Password) error
	Delete(*authenticator.Password) error
	Authenticate(a *authenticator.Password, password string) (requireUpdate bool, err error)
}

type PasskeyAuthenticatorProvider interface {
	Get(userID, id string) (*authenticator.Passkey, error)
	GetMany(ids []string) ([]*authenticator.Passkey, error)
	List(userID string) ([]*authenticator.Passkey, error)
	New(
		id string,
		userID string,
		attestationResponse []byte,
		isDefault bool,
		kind string,
	) (*authenticator.Passkey, error)
	Create(*authenticator.Passkey) error
	Update(*authenticator.Passkey) error
	Delete(*authenticator.Passkey) error
	Authenticate(a *authenticator.Passkey, assertionResponse []byte) (requireUpdate bool, err error)
}

type TOTPAuthenticatorProvider interface {
	Get(userID, id string) (*authenticator.TOTP, error)
	GetMany(ids []string) ([]*authenticator.TOTP, error)
	List(userID string) ([]*authenticator.TOTP, error)
	New(id string, userID string, displayName string, isDefault bool, kind string) *authenticator.TOTP
	Create(*authenticator.TOTP) error
	Delete(*authenticator.TOTP) error
	Authenticate(a *authenticator.TOTP, code string) error
}

type OOBOTPAuthenticatorProvider interface {
	Get(userID, id string) (*authenticator.OOBOTP, error)
	GetMany(ids []string) ([]*authenticator.OOBOTP, error)
	List(userID string) ([]*authenticator.OOBOTP, error)
	New(id string, userID string, oobAuthenticatorType model.AuthenticatorType, target string, isDefault bool, kind string) *authenticator.OOBOTP
	Create(*authenticator.OOBOTP) error
	Delete(*authenticator.OOBOTP) error
}

type OTPCodeService interface {
	VerifyCode(target string, code string) error
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

type Service struct {
	Store          *Store
	Password       PasswordAuthenticatorProvider
	Passkey        PasskeyAuthenticatorProvider
	TOTP           TOTPAuthenticatorProvider
	OOBOTP         OOBOTPAuthenticatorProvider
	OTPCodeService OTPCodeService
	RateLimiter    RateLimiter
}

func (s *Service) Get(id string) (*authenticator.Info, error) {
	ref, err := s.Store.GetRefByID(id)
	if err != nil {
		return nil, err
	}
	switch ref.Type {
	case model.AuthenticatorTypePassword:
		p, err := s.Password.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil

	case model.AuthenticatorTypePasskey:
		p, err := s.Passkey.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil

	case model.AuthenticatorTypeTOTP:
		t, err := s.TOTP.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return t.ToInfo(), nil

	// FIXME(oob): handle getting different channel
	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		o, err := s.OOBOTP.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return o.ToInfo(), nil
	}

	panic("authenticator: unknown authenticator type " + ref.Type)
}

func (s *Service) GetMany(ids []string) ([]*authenticator.Info, error) {
	refs, err := s.Store.ListRefsByIDs(ids)
	if err != nil {
		return nil, err
	}

	var passwordIDs, passkeyIDs, totpIDs, oobIDs []string
	for _, ref := range refs {
		switch ref.Type {
		case model.AuthenticatorTypePassword:
			passwordIDs = append(passwordIDs, ref.ID)
		case model.AuthenticatorTypePasskey:
			passkeyIDs = append(passkeyIDs, ref.ID)
		case model.AuthenticatorTypeTOTP:
			totpIDs = append(totpIDs, ref.ID)
		case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
			oobIDs = append(oobIDs, ref.ID)
		default:
			panic("authenticator: unknown authenticator type " + ref.Type)
		}
	}

	var infos []*authenticator.Info

	{
		p, err := s.Password.GetMany(passwordIDs)
		if err != nil {
			return nil, err
		}
		for _, a := range p {
			infos = append(infos, a.ToInfo())
		}
	}
	{
		passkeys, err := s.Passkey.GetMany(passkeyIDs)
		if err != nil {
			return nil, err
		}
		for _, a := range passkeys {
			infos = append(infos, a.ToInfo())
		}
	}

	{
		t, err := s.TOTP.GetMany(totpIDs)
		if err != nil {
			return nil, err
		}
		for _, a := range t {
			infos = append(infos, a.ToInfo())
		}
	}

	{
		o, err := s.OOBOTP.GetMany(oobIDs)
		if err != nil {
			return nil, err
		}
		for _, a := range o {
			infos = append(infos, a.ToInfo())
		}
	}

	return infos, nil
}

func (s *Service) List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	var ais []*authenticator.Info
	{
		as, err := s.Password.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, a.ToInfo())
		}
	}
	{
		as, err := s.Passkey.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, a.ToInfo())
		}
	}
	{
		as, err := s.TOTP.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, a.ToInfo())
		}
	}
	{
		as, err := s.OOBOTP.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, a.ToInfo())
		}
	}

	var filtered []*authenticator.Info
	for _, a := range ais {
		keep := true
		for _, f := range filters {
			if !f.Keep(a) {
				keep = false
				break
			}
		}
		if keep {
			filtered = append(filtered, a)
		}
	}

	return filtered, nil
}

func (s *Service) Count(userID string) (uint64, error) {
	return s.Store.Count(userID)
}

func (s *Service) ListRefsByUsers(userIDs []string, authenticatorType *model.AuthenticatorType, authenticatorKind *authenticator.Kind) ([]*authenticator.Ref, error) {
	return s.Store.ListRefsByUsers(userIDs, authenticatorType, authenticatorKind)
}

func (s *Service) New(spec *authenticator.Spec) (*authenticator.Info, error) {
	return s.NewWithAuthenticatorID("", spec)
}

func (s *Service) NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error) {
	switch spec.Type {
	case model.AuthenticatorTypePassword:
		plainPassword := spec.Password.PlainPassword
		p, err := s.Password.New(authenticatorID, spec.UserID, plainPassword, spec.IsDefault, string(spec.Kind))
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil

	case model.AuthenticatorTypePasskey:
		attestationResponse := spec.Passkey.AttestationResponse

		p, err := s.Passkey.New(
			authenticatorID,
			spec.UserID,
			attestationResponse,
			spec.IsDefault,
			string(spec.Kind),
		)
		if err != nil {
			return nil, err
		}
		return p.ToInfo(), nil

	case model.AuthenticatorTypeTOTP:
		displayName := spec.TOTP.DisplayName
		t := s.TOTP.New(authenticatorID, spec.UserID, displayName, spec.IsDefault, string(spec.Kind))
		return t.ToInfo(), nil

	case model.AuthenticatorTypeOOBEmail:
		email := spec.OOBOTP.Email
		o := s.OOBOTP.New(authenticatorID, spec.UserID, model.AuthenticatorTypeOOBEmail, email, spec.IsDefault, string(spec.Kind))
		return o.ToInfo(), nil

	case model.AuthenticatorTypeOOBSMS:
		phone := spec.OOBOTP.Phone
		o := s.OOBOTP.New(authenticatorID, spec.UserID, model.AuthenticatorTypeOOBSMS, phone, spec.IsDefault, string(spec.Kind))
		return o.ToInfo(), nil

	}

	panic("authenticator: unknown authenticator type " + spec.Type)
}

func (s *Service) WithSpec(ai *authenticator.Info, spec *authenticator.Spec) (bool, *authenticator.Info, error) {
	changed := false
	switch ai.Type {
	case model.AuthenticatorTypePassword:
		a := ai.Password
		plainPassword := spec.Password.PlainPassword
		newAuth, err := s.Password.WithPassword(a, plainPassword)
		if err != nil {
			return false, nil, err
		}
		changed = (newAuth != a)
		return changed, newAuth.ToInfo(), nil
	}

	panic("authenticator: update authenticator is not supported for type " + ai.Type)
}

func (s *Service) Create(info *authenticator.Info) error {
	ais, err := s.List(info.UserID)
	if err != nil {
		return err
	}

	for _, a := range ais {
		if info.Equal(a) {
			err = authenticator.NewErrDuplicatedAuthenticator(info.Type)
			return err
		}
	}

	switch info.Type {
	case model.AuthenticatorTypePassword:
		a := info.Password
		if err := s.Password.Create(a); err != nil {
			return err
		}
		*info = *a.ToInfo()
	case model.AuthenticatorTypePasskey:
		a := info.Passkey
		if err := s.Passkey.Create(a); err != nil {
			return err
		}
		*info = *a.ToInfo()
	case model.AuthenticatorTypeTOTP:
		a := info.TOTP
		if err := s.TOTP.Create(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		a := info.OOBOTP
		if err := s.OOBOTP.Create(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	default:
		panic("authenticator: unknown authenticator type " + info.Type)
	}

	return nil
}

func (s *Service) Update(info *authenticator.Info) error {
	switch info.Type {
	case model.AuthenticatorTypePassword:
		a := info.Password
		if err := s.Password.UpdatePassword(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	case model.AuthenticatorTypePasskey:
		a := info.Passkey
		if err := s.Passkey.Update(a); err != nil {
			return err
		}
		*info = *a.ToInfo()
	default:
		panic("authenticator: unknown authenticator type for update" + info.Type)
	}

	return nil
}

func (s *Service) Delete(info *authenticator.Info) error {
	switch info.Type {
	case model.AuthenticatorTypePassword:
		a := info.Password
		if err := s.Password.Delete(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	case model.AuthenticatorTypePasskey:
		a := info.Passkey
		if err := s.Passkey.Delete(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	case model.AuthenticatorTypeTOTP:
		a := info.TOTP
		if err := s.TOTP.Delete(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		a := info.OOBOTP
		if err := s.OOBOTP.Delete(a); err != nil {
			return err
		}
		*info = *a.ToInfo()

	default:
		panic("authenticator: delete authenticator is not supported yet for type " + info.Type)
	}

	return nil
}

func (s *Service) VerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec) (requireUpdate bool, err error) {
	err = s.RateLimiter.TakeToken(AntiBruteForceAuthenticateBucket(info.UserID, info.Type))
	if err != nil {
		return
	}

	switch info.Type {
	case model.AuthenticatorTypePassword:
		plainPassword := spec.Password.PlainPassword
		a := info.Password
		requireUpdate, err = s.Password.Authenticate(a, plainPassword)
		if err != nil {
			err = authenticator.ErrInvalidCredentials
			return
		}
		*info = *a.ToInfo()
		return
	case model.AuthenticatorTypePasskey:
		assertionResponse := spec.Passkey.AssertionResponse
		a := info.Passkey
		requireUpdate, err = s.Passkey.Authenticate(a, assertionResponse)
		if err != nil {
			err = authenticator.ErrInvalidCredentials
			return
		}
		*info = *a.ToInfo()

		return
	case model.AuthenticatorTypeTOTP:
		code := spec.TOTP.Code
		a := info.TOTP
		if s.TOTP.Authenticate(a, code) != nil {
			err = authenticator.ErrInvalidCredentials
			return
		}
		// Do not update info because by definition TOTP does not update itself during verification.

		return
	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		code := spec.OOBOTP.Code
		a := info.OOBOTP
		err = s.OTPCodeService.VerifyCode(a.ToTarget(), code)
		if errors.Is(err, otp.ErrInvalidCode) {
			err = authenticator.ErrInvalidCredentials
			return
		} else if err != nil {
			return
		}
		// Do not update info because by definition OOBOTP does not update itself during verification.

		return
	}

	panic("authenticator: unhandled authenticator type " + info.Type)
}

func (s *Service) RemoveOrphans(identities []*identity.Info) error {
	if len(identities) == 0 {
		return nil
	}

	authenticators, err := s.List(identities[0].UserID)
	if err != nil {
		return err
	}

	for _, a := range authenticators {
		if a.IsIndependent() {
			continue
		}

		orphaned := true
		for _, i := range identities {
			if a.IsDependentOf(i) {
				orphaned = false
			}
		}

		if orphaned {
			err = s.Delete(a)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
