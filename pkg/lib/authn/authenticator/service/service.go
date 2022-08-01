package service

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

type PasswordAuthenticatorProvider interface {
	Get(userID, id string) (*password.Authenticator, error)
	GetMany(ids []string) ([]*password.Authenticator, error)
	List(userID string) ([]*password.Authenticator, error)
	New(id string, userID string, password string, isDefault bool, kind string) (*password.Authenticator, error)
	// WithPassword returns new authenticator pointer if password is changed
	// Otherwise original authenticator will be returned
	WithPassword(a *password.Authenticator, password string) (*password.Authenticator, error)
	Create(*password.Authenticator) error
	UpdatePassword(*password.Authenticator) error
	Delete(*password.Authenticator) error
	Authenticate(a *password.Authenticator, password string) (requireUpdate bool, err error)
}

type TOTPAuthenticatorProvider interface {
	Get(userID, id string) (*totp.Authenticator, error)
	GetMany(ids []string) ([]*totp.Authenticator, error)
	List(userID string) ([]*totp.Authenticator, error)
	New(id string, userID string, displayName string, isDefault bool, kind string) *totp.Authenticator
	Create(*totp.Authenticator) error
	Delete(*totp.Authenticator) error
	Authenticate(a *totp.Authenticator, code string) error
}

type OOBOTPAuthenticatorProvider interface {
	Get(userID, id string) (*oob.Authenticator, error)
	GetMany(ids []string) ([]*oob.Authenticator, error)
	List(userID string) ([]*oob.Authenticator, error)
	New(id string, userID string, oobAuthenticatorType model.AuthenticatorType, target string, isDefault bool, kind string) *oob.Authenticator
	Create(*oob.Authenticator) error
	Delete(*oob.Authenticator) error
	VerifyCode(authenticatorID string, code string) (*oob.Code, error)
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

type Service struct {
	Store       *Store
	Password    PasswordAuthenticatorProvider
	TOTP        TOTPAuthenticatorProvider
	OOBOTP      OOBOTPAuthenticatorProvider
	RateLimiter RateLimiter
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
		return passwordToAuthenticatorInfo(p), nil

	case model.AuthenticatorTypeTOTP:
		t, err := s.TOTP.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return totpToAuthenticatorInfo(t), nil

	// FIXME(oob): handle getting different channel
	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		o, err := s.OOBOTP.Get(ref.UserID, id)
		if err != nil {
			return nil, err
		}
		return oobotpToAuthenticatorInfo(o), nil
	}

	panic("authenticator: unknown authenticator type " + ref.Type)
}

func (s *Service) GetMany(ids []string) ([]*authenticator.Info, error) {
	refs, err := s.Store.ListRefsByIDs(ids)
	if err != nil {
		return nil, err
	}

	var passwordIDs, totpIDs, oobIDs []string
	for _, ref := range refs {
		switch ref.Type {
		case model.AuthenticatorTypePassword:
			passwordIDs = append(passwordIDs, ref.ID)
		case model.AuthenticatorTypeTOTP:
			totpIDs = append(totpIDs, ref.ID)
		case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
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

func (s *Service) New(spec *authenticator.Spec) (*authenticator.Info, error) {
	return s.NewWithAuthenticatorID("", spec)
}

func (s *Service) NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error) {
	switch spec.Type {
	case model.AuthenticatorTypePassword:
		plainPassword := spec.Claims[authenticator.AuthenticatorClaimPasswordPlainPassword].(string)
		p, err := s.Password.New(authenticatorID, spec.UserID, plainPassword, spec.IsDefault, string(spec.Kind))
		if err != nil {
			return nil, err
		}
		return passwordToAuthenticatorInfo(p), nil

	case model.AuthenticatorTypeTOTP:
		displayName := spec.Claims[authenticator.AuthenticatorClaimTOTPDisplayName].(string)
		t := s.TOTP.New(authenticatorID, spec.UserID, displayName, spec.IsDefault, string(spec.Kind))
		return totpToAuthenticatorInfo(t), nil

	case model.AuthenticatorTypeOOBEmail:
		email := spec.Claims[authenticator.AuthenticatorClaimOOBOTPEmail].(string)
		o := s.OOBOTP.New(authenticatorID, spec.UserID, model.AuthenticatorTypeOOBEmail, email, spec.IsDefault, string(spec.Kind))
		return oobotpToAuthenticatorInfo(o), nil

	case model.AuthenticatorTypeOOBSMS:
		phone := spec.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string)
		o := s.OOBOTP.New(authenticatorID, spec.UserID, model.AuthenticatorTypeOOBSMS, phone, spec.IsDefault, string(spec.Kind))
		return oobotpToAuthenticatorInfo(o), nil
	}

	panic("authenticator: unknown authenticator type " + spec.Type)
}

func (s *Service) WithSpec(ai *authenticator.Info, spec *authenticator.Spec) (bool, *authenticator.Info, error) {
	changed := false
	switch ai.Type {
	case model.AuthenticatorTypePassword:
		a := passwordFromAuthenticatorInfo(ai)
		plainPassword := spec.Claims[authenticator.AuthenticatorClaimPasswordPlainPassword].(string)
		newAuth, err := s.Password.WithPassword(a, plainPassword)
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
			err = authenticator.NewErrDuplicatedAuthenticator(info.Type)
			return err
		}
	}

	switch info.Type {
	case model.AuthenticatorTypePassword:
		a := passwordFromAuthenticatorInfo(info)
		if err := s.Password.Create(a); err != nil {
			return err
		}

	case model.AuthenticatorTypeTOTP:
		a := totpFromAuthenticatorInfo(info)
		if err := s.TOTP.Create(a); err != nil {
			return err
		}

	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
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
	case model.AuthenticatorTypePassword:
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
	case model.AuthenticatorTypePassword:
		a := passwordFromAuthenticatorInfo(info)
		if err := s.Password.Delete(a); err != nil {
			return err
		}

	case model.AuthenticatorTypeTOTP:
		a := totpFromAuthenticatorInfo(info)
		if err := s.TOTP.Delete(a); err != nil {
			return err
		}

	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		a := oobotpFromAuthenticatorInfo(info)
		if err := s.OOBOTP.Delete(a); err != nil {
			return err
		}
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
		plainPassword := spec.Claims[authenticator.AuthenticatorClaimPasswordPlainPassword].(string)
		a := passwordFromAuthenticatorInfo(info)
		requireUpdate, err = s.Password.Authenticate(a, plainPassword)
		if err != nil {
			err = authenticator.ErrInvalidCredentials
			return
		}
		return
	case model.AuthenticatorTypeTOTP:
		code := spec.Claims[authenticator.AuthenticatorClaimTOTPCode].(string)
		a := totpFromAuthenticatorInfo(info)
		if s.TOTP.Authenticate(a, code) != nil {
			err = authenticator.ErrInvalidCredentials
			return
		}
		return
	case model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypeOOBSMS:
		code := spec.Claims[authenticator.AuthenticatorClaimOOBOTPCode].(string)
		a := oobotpFromAuthenticatorInfo(info)
		_, err = s.OOBOTP.VerifyCode(a.ID, code)
		if errors.Is(err, oob.ErrInvalidCode) {
			err = authenticator.ErrInvalidCredentials
			return
		} else if err != nil {
			return
		}
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
		if a.Kind != authenticator.KindPrimary ||
			(a.Type != model.AuthenticatorTypeOOBEmail && a.Type != model.AuthenticatorTypeOOBSMS) {
			continue
		}

		aClaims := a.StandardClaims()

		orphaned := true
		for _, i := range identities {
			// Matching identities with same claim => not orphan
			isMatching := false
			for _, t := range i.PrimaryAuthenticatorTypes() {
				if t == a.Type {
					isMatching = true
					break
				}
			}
			if !isMatching {
				continue
			}

			for k, v := range i.StandardClaims() {
				if aClaims[k] == v {
					orphaned = false
					break
				}
			}
			if !orphaned {
				break
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
