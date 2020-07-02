package provider

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/bearertoken"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/password"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/recoverycode"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type PasswordAuthenticatorProvider interface {
	Get(userID, id string) (*password.Authenticator, error)
	List(userID string) ([]*password.Authenticator, error)
	New(userID string, password string) (*password.Authenticator, error)
	// WithPassword returns new authenticator pointer if password is changed
	// Otherwise original authenticator will be returned
	WithPassword(userID string, a *password.Authenticator, password string) (*password.Authenticator, error)
	Create(*password.Authenticator) error
	UpdatePassword(*password.Authenticator) error
	Delete(*password.Authenticator) error
	Authenticate(a *password.Authenticator, password string) error
}

type TOTPAuthenticatorProvider interface {
	Get(userID, id string) (*totp.Authenticator, error)
	List(userID string) ([]*totp.Authenticator, error)
	New(userID string, displayName string) *totp.Authenticator
	Create(*totp.Authenticator) error
	Delete(*totp.Authenticator) error
	Authenticate(candidates []*totp.Authenticator, code string) *totp.Authenticator
}

type OOBOTPAuthenticatorProvider interface {
	Get(userID, id string) (*oob.Authenticator, error)
	List(userID string) ([]*oob.Authenticator, error)
	New(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string) *oob.Authenticator
	Create(*oob.Authenticator) error
	Delete(*oob.Authenticator) error
	Authenticate(expectedCode string, code string) error
}

type BearerTokenAuthenticatorProvider interface {
	Get(userID, id string) (*bearertoken.Authenticator, error)
	GetByToken(userID string, token string) (*bearertoken.Authenticator, error)
	List(userID string) ([]*bearertoken.Authenticator, error)
	New(userID string, parentID string) *bearertoken.Authenticator
	Create(*bearertoken.Authenticator) error
	Authenticate(authenticator *bearertoken.Authenticator, token string) error
}

type RecoveryCodeAuthenticatorProvider interface {
	Get(userID, id string) (*recoverycode.Authenticator, error)
	List(userID string) ([]*recoverycode.Authenticator, error)
	Generate(userID string) []*recoverycode.Authenticator
	ReplaceAll(userID string, as []*recoverycode.Authenticator) error
	Authenticate(candidates []*recoverycode.Authenticator, code string) *recoverycode.Authenticator
}

type Provider struct {
	Password     PasswordAuthenticatorProvider
	TOTP         TOTPAuthenticatorProvider
	OOBOTP       OOBOTPAuthenticatorProvider
	BearerToken  BearerTokenAuthenticatorProvider
	RecoveryCode RecoveryCodeAuthenticatorProvider
}

func (a *Provider) Get(userID string, typ authn.AuthenticatorType, id string) (*authenticator.Info, error) {
	switch typ {
	case authn.AuthenticatorTypePassword:
		p, err := a.Password.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return passwordToAuthenticatorInfo(p), nil

	case authn.AuthenticatorTypeTOTP:
		t, err := a.TOTP.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return totpToAuthenticatorInfo(t), nil

	case authn.AuthenticatorTypeOOB:
		o, err := a.OOBOTP.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return oobotpToAuthenticatorInfo(o), nil

	case authn.AuthenticatorTypeBearerToken:
		b, err := a.BearerToken.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return bearerTokenToAuthenticatorInfo(b), nil

	case authn.AuthenticatorTypeRecoveryCode:
		r, err := a.RecoveryCode.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return recoveryCodeToAuthenticatorInfo(r), nil
	}

	panic("interaction_adaptors: unknown authenticator type " + typ)
}

func (a *Provider) List(userID string, typ authn.AuthenticatorType) ([]*authenticator.Info, error) {
	var ais []*authenticator.Info
	switch typ {
	case authn.AuthenticatorTypePassword:
		as, err := a.Password.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, passwordToAuthenticatorInfo(a))
		}

	case authn.AuthenticatorTypeTOTP:
		as, err := a.TOTP.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, totpToAuthenticatorInfo(a))
		}

	case authn.AuthenticatorTypeOOB:
		as, err := a.OOBOTP.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, oobotpToAuthenticatorInfo(a))
		}

	case authn.AuthenticatorTypeBearerToken:
		as, err := a.BearerToken.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, bearerTokenToAuthenticatorInfo(a))
		}

	case authn.AuthenticatorTypeRecoveryCode:
		as, err := a.RecoveryCode.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, recoveryCodeToAuthenticatorInfo(a))
		}

	default:
		panic("interaction_adaptors: unknown authenticator type " + typ)
	}
	return ais, nil
}

func (a *Provider) ListByIdentity(userID string, ii *identity.Info) (ais []*authenticator.Info, err error) {
	// This function takes IdentityInfo instead of IdentitySpec because
	// The login ID value in IdentityInfo is normalized.
	switch ii.Type {
	case authn.IdentityTypeOAuth:
		// OAuth Identity does not have associated authenticators.
		return
	case authn.IdentityTypeLoginID:
		// Login ID Identity has password, TOTP and OOB OTP.
		// Note that we only return OOB OTP associated with the login ID.
		var pas []*password.Authenticator
		pas, err = a.Password.List(userID)
		if err != nil {
			return
		}
		for _, pa := range pas {
			ais = append(ais, passwordToAuthenticatorInfo(pa))
		}

		var tas []*totp.Authenticator
		tas, err = a.TOTP.List(userID)
		if err != nil {
			return
		}
		for _, ta := range tas {
			ais = append(ais, totpToAuthenticatorInfo(ta))
		}

		loginID := ii.Claims[identity.IdentityClaimLoginIDValue]
		var oas []*oob.Authenticator
		oas, err = a.OOBOTP.List(userID)
		if err != nil {
			return
		}
		for _, oa := range oas {
			if oa.Email == loginID || oa.Phone == loginID {
				ais = append(ais, oobotpToAuthenticatorInfo(oa))
			}
		}
	case authn.IdentityTypeAnonymous:
		// Anonymous Identity does not have associated authenticators.
		return
	default:
		panic("interaction_adaptors: unknown identity type " + ii.Type)
	}

	return
}

func (a *Provider) New(userID string, spec authenticator.Spec, secret string) ([]*authenticator.Info, error) {
	switch spec.Type {
	case authn.AuthenticatorTypePassword:
		p, err := a.Password.New(userID, secret)
		if err != nil {
			return nil, err
		}
		return []*authenticator.Info{passwordToAuthenticatorInfo(p)}, nil

	case authn.AuthenticatorTypeTOTP:
		displayName, _ := spec.Props[authenticator.AuthenticatorPropTOTPDisplayName].(string)
		t := a.TOTP.New(userID, displayName)
		return []*authenticator.Info{totpToAuthenticatorInfo(t)}, nil

	case authn.AuthenticatorTypeOOB:
		channel := spec.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string)
		var phone, email string
		switch authn.AuthenticatorOOBChannel(channel) {
		case authn.AuthenticatorOOBChannelSMS:
			phone = spec.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string)
		case authn.AuthenticatorOOBChannelEmail:
			email = spec.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string)
		}
		o := a.OOBOTP.New(userID, authn.AuthenticatorOOBChannel(channel), phone, email)
		return []*authenticator.Info{oobotpToAuthenticatorInfo(o)}, nil

	case authn.AuthenticatorTypeBearerToken:
		parentID := spec.Props[authenticator.AuthenticatorPropBearerTokenParentID].(string)
		b := a.BearerToken.New(userID, parentID)
		return []*authenticator.Info{bearerTokenToAuthenticatorInfo(b)}, nil

	case authn.AuthenticatorTypeRecoveryCode:
		rs := a.RecoveryCode.Generate(userID)
		var ais []*authenticator.Info
		for _, r := range rs {
			ais = append(ais, recoveryCodeToAuthenticatorInfo(r))
		}
		return ais, nil
	}

	panic("interaction_adaptors: unknown authenticator type " + spec.Type)
}

func (a *Provider) WithSecret(userID string, ai *authenticator.Info, secret string) (bool, *authenticator.Info, error) {
	changed := false
	switch ai.Type {
	case authn.AuthenticatorTypePassword:
		authen := passwordFromAuthenticatorInfo(userID, ai)
		newAuth, err := a.Password.WithPassword(userID, authen, secret)
		if err != nil {
			return false, nil, err
		}
		changed = (newAuth != authen)
		return changed, passwordToAuthenticatorInfo(newAuth), nil
	}

	panic("interaction_adaptors: update authenticator is not supported for type " + ai.Type)
}

func (a *Provider) CreateAll(userID string, ais []*authenticator.Info) error {
	var recoveryCodes []*recoverycode.Authenticator
	for _, ai := range ais {
		switch ai.Type {
		case authn.AuthenticatorTypePassword:
			authenticator := passwordFromAuthenticatorInfo(userID, ai)
			if err := a.Password.Create(authenticator); err != nil {
				return err
			}

		case authn.AuthenticatorTypeTOTP:
			authenticator := totpFromAuthenticatorInfo(userID, ai)
			if err := a.TOTP.Create(authenticator); err != nil {
				return err
			}

		case authn.AuthenticatorTypeOOB:
			authenticator := oobotpFromAuthenticatorInfo(userID, ai)
			if err := a.OOBOTP.Create(authenticator); err != nil {
				return err
			}

		case authn.AuthenticatorTypeBearerToken:
			authenticator := bearerTokenFromAuthenticatorInfo(userID, ai)
			if err := a.BearerToken.Create(authenticator); err != nil {
				return err
			}

		case authn.AuthenticatorTypeRecoveryCode:
			authenticator := recoveryCodeFromAuthenticatorInfo(userID, ai)
			recoveryCodes = append(recoveryCodes, authenticator)

		default:
			panic("interaction_adaptors: unknown authenticator type " + ai.Type)
		}
	}

	if len(recoveryCodes) > 0 {
		err := a.RecoveryCode.ReplaceAll(userID, recoveryCodes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Provider) UpdateAll(userID string, ais []*authenticator.Info) error {
	var recoveryCodes []*recoverycode.Authenticator
	for _, ai := range ais {
		switch ai.Type {
		case authn.AuthenticatorTypePassword:
			authenticator := passwordFromAuthenticatorInfo(userID, ai)
			if err := a.Password.UpdatePassword(authenticator); err != nil {
				return err
			}
		default:
			panic("interaction_adaptors: unknown authenticator type for update" + ai.Type)
		}
	}

	if len(recoveryCodes) > 0 {
		err := a.RecoveryCode.ReplaceAll(userID, recoveryCodes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Provider) DeleteAll(userID string, ais []*authenticator.Info) error {
	for _, ai := range ais {
		switch ai.Type {
		case authn.AuthenticatorTypePassword:
			authenticator := passwordFromAuthenticatorInfo(userID, ai)
			if err := a.Password.Delete(authenticator); err != nil {
				return err
			}

		case authn.AuthenticatorTypeTOTP:
			authenticator := totpFromAuthenticatorInfo(userID, ai)
			if err := a.TOTP.Delete(authenticator); err != nil {
				return err
			}

		case authn.AuthenticatorTypeOOB:
			authenticator := oobotpFromAuthenticatorInfo(userID, ai)
			if err := a.OOBOTP.Delete(authenticator); err != nil {
				return err
			}
		default:
			panic("interaction_adaptors: delete authenticator is not supported yet for type " + ai.Type)
		}
	}

	return nil
}

func (a *Provider) Authenticate(userID string, spec authenticator.Spec, state *map[string]string, secret string) (*authenticator.Info, error) {
	switch spec.Type {
	case authn.AuthenticatorTypePassword:
		ps, err := a.Password.List(userID)
		if err != nil {
			return nil, err
		}
		if len(ps) != 1 {
			return nil, interaction.ErrInvalidCredentials
		}

		if a.Password.Authenticate(ps[0], secret) != nil {
			return nil, interaction.ErrInvalidCredentials
		}
		return passwordToAuthenticatorInfo(ps[0]), nil

	case authn.AuthenticatorTypeTOTP:
		ts, err := a.TOTP.List(userID)
		if err != nil {
			return nil, err
		}

		t := a.TOTP.Authenticate(ts, secret)
		if t == nil {
			return nil, interaction.ErrInvalidCredentials
		}
		return totpToAuthenticatorInfo(t), nil

	case authn.AuthenticatorTypeOOB:
		if state == nil {
			return nil, interaction.ErrInvalidCredentials
		}
		id := (*state)[authenticator.AuthenticatorStateOOBOTPID]
		code := (*state)[authenticator.AuthenticatorStateOOBOTPCode]

		var o *oob.Authenticator
		// This function can be called by login or signup.
		// In case of login, we must check if the authenticator belongs to the user.
		if id != "" {
			var err error
			o, err = a.OOBOTP.Get(userID, id)
			if errors.Is(err, authenticator.ErrAuthenticatorNotFound) {
				return nil, interaction.ErrInvalidCredentials
			} else if err != nil {
				return nil, err
			}
		}

		if a.OOBOTP.Authenticate(code, secret) != nil {
			return nil, interaction.ErrInvalidCredentials
		}

		if o != nil {
			return oobotpToAuthenticatorInfo(o), nil
		}
		return nil, nil
	case authn.AuthenticatorTypeBearerToken:
		b, err := a.BearerToken.GetByToken(userID, secret)
		if errors.Is(err, authenticator.ErrAuthenticatorNotFound) {
			return nil, interaction.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}

		if a.BearerToken.Authenticate(b, secret) != nil {
			return nil, interaction.ErrInvalidCredentials
		}
		return bearerTokenToAuthenticatorInfo(b), nil

	case authn.AuthenticatorTypeRecoveryCode:
		rs, err := a.RecoveryCode.List(userID)
		if err != nil {
			return nil, err
		}

		r := a.RecoveryCode.Authenticate(rs, secret)
		if r == nil {
			return nil, interaction.ErrInvalidCredentials
		}
		return recoveryCodeToAuthenticatorInfo(r), nil
	}

	panic("interaction_adaptors: unknown authenticator type " + spec.Type)
}

func (a *Provider) VerifySecret(userID string, ai *authenticator.Info, secret string) error {
	switch ai.Type {
	case authn.AuthenticatorTypePassword:
		authen := passwordFromAuthenticatorInfo(userID, ai)
		if a.Password.Authenticate(authen, secret) != nil {
			return interaction.ErrInvalidCredentials
		}
		return nil
	}

	panic("interaction_adaptors: unhandled authenticator type " + ai.Type)
}
