package service

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
)

type PasswordAuthenticatorProvider interface {
	Get(userID, id string) (*password.Authenticator, error)
	GetMany(ids []string) ([]*password.Authenticator, error)
	List(userID string) ([]*password.Authenticator, error)
	New(userID string, password string, isDefault bool, kind string) (*password.Authenticator, error)
	// WithPassword returns new authenticator pointer if password is changed
	// Otherwise original authenticator will be returned
	WithPassword(a *password.Authenticator, password string) (*password.Authenticator, error)
	Create(*password.Authenticator) error
	UpdatePassword(*password.Authenticator) error
	Delete(*password.Authenticator) error
	Authenticate(a *password.Authenticator, password string) error
}

type TOTPAuthenticatorProvider interface {
	Get(userID, id string) (*totp.Authenticator, error)
	GetMany(ids []string) ([]*totp.Authenticator, error)
	List(userID string) ([]*totp.Authenticator, error)
	New(userID string, displayName string, isDefault bool, kind string) *totp.Authenticator
	Create(*totp.Authenticator) error
	Delete(*totp.Authenticator) error
	Authenticate(a *totp.Authenticator, code string) error
}

type OOBOTPAuthenticatorProvider interface {
	Get(userID, id string) (*oob.Authenticator, error)
	GetMany(ids []string) ([]*oob.Authenticator, error)
	List(userID string) ([]*oob.Authenticator, error)
	New(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string, isDefault bool, kind string) *oob.Authenticator
	Create(*oob.Authenticator) error
	Delete(*oob.Authenticator) error
	Authenticate(secret string, channel authn.AuthenticatorOOBChannel, code string) error
}

type Service struct {
	Store    *Store
	Password PasswordAuthenticatorProvider
	TOTP     TOTPAuthenticatorProvider
	OOBOTP   OOBOTPAuthenticatorProvider
}

func (s *Service) Get(userID string, typ authn.AuthenticatorType, id string) (*authenticator.Info, error) {
	switch typ {
	case authn.AuthenticatorTypePassword:
		p, err := s.Password.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return passwordToAuthenticatorInfo(p), nil

	case authn.AuthenticatorTypeTOTP:
		t, err := s.TOTP.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return totpToAuthenticatorInfo(t), nil

	case authn.AuthenticatorTypeOOB:
		o, err := s.OOBOTP.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return oobotpToAuthenticatorInfo(o), nil
	}

	panic("authenticator: unknown authenticator type " + typ)
}

func (s *Service) GetMany(refs []*authenticator.Ref) ([]*authenticator.Info, error) {
	var passwordIDs, totpIDs, oobIDs []string
	for _, ref := range refs {
		switch ref.Type {
		case authn.AuthenticatorTypePassword:
			passwordIDs = append(passwordIDs, ref.ID)
		case authn.AuthenticatorTypeTOTP:
			totpIDs = append(totpIDs, ref.ID)
		case authn.AuthenticatorTypeOOB:
			oobIDs = append(oobIDs, ref.ID)
		default:
			panic("authenticator: unknown authenticator type " + ref.Type)
		}
	}

	var infos []*authenticator.Info

	p, err := s.Password.GetMany(passwordIDs)
	if err != nil {
		return nil, err
	}
	for _, a := range p {
		infos = append(infos, passwordToAuthenticatorInfo(a))
	}

	t, err := s.TOTP.GetMany(totpIDs)
	if err != nil {
		return nil, err
	}
	for _, a := range t {
		infos = append(infos, totpToAuthenticatorInfo(a))
	}

	o, err := s.OOBOTP.GetMany(oobIDs)
	if err != nil {
		return nil, err
	}
	for _, a := range o {
		infos = append(infos, oobotpToAuthenticatorInfo(a))
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
			ais = append(ais, passwordToAuthenticatorInfo(a))
		}
	}
	{
		as, err := s.TOTP.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, totpToAuthenticatorInfo(a))
		}
	}
	{
		as, err := s.OOBOTP.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, oobotpToAuthenticatorInfo(a))
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

func (s *Service) ListRefsByUsers(userIDs []string) ([]*authenticator.Ref, error) {
	return s.Store.ListRefsByUsers(userIDs)
}

func (s *Service) New(spec *authenticator.Spec, secret string) (*authenticator.Info, error) {
	switch spec.Type {
	case authn.AuthenticatorTypePassword:
		p, err := s.Password.New(spec.UserID, secret, spec.IsDefault, string(spec.Kind))
		if err != nil {
			return nil, err
		}
		return passwordToAuthenticatorInfo(p), nil

	case authn.AuthenticatorTypeTOTP:
		displayName := spec.Claims[authenticator.AuthenticatorClaimTOTPDisplayName].(string)
		t := s.TOTP.New(spec.UserID, displayName, spec.IsDefault, string(spec.Kind))
		return totpToAuthenticatorInfo(t), nil

	case authn.AuthenticatorTypeOOB:
		channel := spec.Claims[authenticator.AuthenticatorClaimOOBOTPChannelType].(string)
		var phone, email string
		switch authn.AuthenticatorOOBChannel(channel) {
		case authn.AuthenticatorOOBChannelSMS:
			phone = spec.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string)
		case authn.AuthenticatorOOBChannelEmail:
			email = spec.Claims[authenticator.AuthenticatorClaimOOBOTPEmail].(string)
		}
		o := s.OOBOTP.New(spec.UserID, authn.AuthenticatorOOBChannel(channel), phone, email, spec.IsDefault, string(spec.Kind))
		return oobotpToAuthenticatorInfo(o), nil
	}

	panic("authenticator: unknown authenticator type " + spec.Type)
}

func (s *Service) WithSecret(ai *authenticator.Info, secret string) (bool, *authenticator.Info, error) {
	changed := false
	switch ai.Type {
	case authn.AuthenticatorTypePassword:
		a := passwordFromAuthenticatorInfo(ai)
		newAuth, err := s.Password.WithPassword(a, secret)
		if err != nil {
			return false, nil, err
		}
		changed = (newAuth != a)
		return changed, passwordToAuthenticatorInfo(newAuth), nil
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
			err = authenticator.ErrAuthenticatorAlreadyExists
			return err
		}
	}

	switch info.Type {
	case authn.AuthenticatorTypePassword:
		a := passwordFromAuthenticatorInfo(info)
		if err := s.Password.Create(a); err != nil {
			return err
		}

	case authn.AuthenticatorTypeTOTP:
		a := totpFromAuthenticatorInfo(info)
		if err := s.TOTP.Create(a); err != nil {
			return err
		}

	case authn.AuthenticatorTypeOOB:
		a := oobotpFromAuthenticatorInfo(info)
		if err := s.OOBOTP.Create(a); err != nil {
			return err
		}

	default:
		panic("authenticator: unknown authenticator type " + info.Type)
	}

	return nil
}

func (s *Service) Update(info *authenticator.Info) error {
	switch info.Type {
	case authn.AuthenticatorTypePassword:
		a := passwordFromAuthenticatorInfo(info)
		if err := s.Password.UpdatePassword(a); err != nil {
			return err
		}
	default:
		panic("authenticator: unknown authenticator type for update" + info.Type)
	}

	return nil
}

func (s *Service) Delete(info *authenticator.Info) error {
	switch info.Type {
	case authn.AuthenticatorTypePassword:
		a := passwordFromAuthenticatorInfo(info)
		if err := s.Password.Delete(a); err != nil {
			return err
		}

	case authn.AuthenticatorTypeTOTP:
		a := totpFromAuthenticatorInfo(info)
		if err := s.TOTP.Delete(a); err != nil {
			return err
		}

	case authn.AuthenticatorTypeOOB:
		a := oobotpFromAuthenticatorInfo(info)
		if err := s.OOBOTP.Delete(a); err != nil {
			return err
		}
	default:
		panic("authenticator: delete authenticator is not supported yet for type " + info.Type)
	}

	return nil
}

func (s *Service) VerifySecret(info *authenticator.Info, state map[string]string, secret string) error {
	switch info.Type {
	case authn.AuthenticatorTypePassword:
		a := passwordFromAuthenticatorInfo(info)
		if s.Password.Authenticate(a, secret) != nil {
			return authenticator.ErrInvalidCredentials
		}
		return nil

	case authn.AuthenticatorTypeTOTP:
		a := totpFromAuthenticatorInfo(info)
		if s.TOTP.Authenticate(a, secret) != nil {
			return authenticator.ErrInvalidCredentials
		}
		return nil

	case authn.AuthenticatorTypeOOB:
		a := oobotpFromAuthenticatorInfo(info)
		otpSecret := state[authenticator.AuthenticatorStateOOBOTPSecret]
		if s.OOBOTP.Authenticate(otpSecret, a.Channel, secret) != nil {
			return authenticator.ErrInvalidCredentials
		}
		return nil
	}

	panic("authenticator: unhandled authenticator type " + info.Type)
}
