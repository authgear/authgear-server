package adaptors

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/bearertoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/recoverycode"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type PasswordAuthenticatorProvider interface {
	Get(userID, id string) (*password.Authenticator, error)
	List(userID string) ([]*password.Authenticator, error)
	New(userID string, password string) (*password.Authenticator, error)
	Authenticate(a *password.Authenticator, password string) error
}

type TOTPAuthenticatorProvider interface {
	Get(userID, id string) (*totp.Authenticator, error)
	List(userID string) ([]*totp.Authenticator, error)
	New(userID string, displayName string) *totp.Authenticator
	Authenticate(candidates []*totp.Authenticator, code string) *totp.Authenticator
}

type OOBOTPAuthenticatorProvider interface {
	Get(userID, id string) (*oob.Authenticator, error)
	List(userID string) ([]*oob.Authenticator, error)
	New(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string) *oob.Authenticator
	Authenticate(a *oob.Authenticator, expectedCode string, code string) error
}

type BearerTokenAuthenticatorProvider interface {
	Get(userID, id string) (*bearertoken.Authenticator, error)
	GetByToken(userID string, token string) (*bearertoken.Authenticator, error)
	List(userID string) ([]*bearertoken.Authenticator, error)
	New(userID string, parentID string) *bearertoken.Authenticator
	Authenticate(authenticator *bearertoken.Authenticator, token string) error
}

type RecoveryCodeAuthenticatorProvider interface {
	Get(userID, id string) (*recoverycode.Authenticator, error)
	List(userID string) ([]*recoverycode.Authenticator, error)
	Generate(userID string) []*recoverycode.Authenticator
	Authenticate(candidates []*recoverycode.Authenticator, code string) *recoverycode.Authenticator
}

type AuthenticatorAdaptor struct {
	Password     PasswordAuthenticatorProvider
	TOTP         TOTPAuthenticatorProvider
	OOBOTP       OOBOTPAuthenticatorProvider
	BearerToken  BearerTokenAuthenticatorProvider
	RecoveryCode RecoveryCodeAuthenticatorProvider
}

func (a *AuthenticatorAdaptor) Get(userID string, typ interaction.AuthenticatorType, id string) (*interaction.AuthenticatorInfo, error) {
	switch typ {
	case interaction.AuthenticatorTypePassword:
		p, err := a.Password.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return passwordToAuthenticatorInfo(p), nil

	case interaction.AuthenticatorTypeTOTP:
		t, err := a.TOTP.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return totpToAuthenticatorInfo(t), nil

	case interaction.AuthenticatorTypeOOBOTP:
		o, err := a.OOBOTP.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return oobotpToAuthenticatorInfo(o), nil

	case interaction.AuthenticatorTypeBearerToken:
		b, err := a.BearerToken.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return bearerTokenToAuthenticatorInfo(b), nil

	case interaction.AuthenticatorTypeRecoveryCode:
		r, err := a.RecoveryCode.Get(userID, id)
		if err != nil {
			return nil, err
		}
		return recoveryCodeToAuthenticatorInfo(r), nil
	}

	panic("interaction_adaptors: unknown authenticator type " + typ)
}

func (a *AuthenticatorAdaptor) List(userID string, typ interaction.AuthenticatorType) ([]*interaction.AuthenticatorInfo, error) {
	var ais []*interaction.AuthenticatorInfo
	switch typ {
	case interaction.AuthenticatorTypePassword:
		as, err := a.Password.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, passwordToAuthenticatorInfo(a))
		}

	case interaction.AuthenticatorTypeTOTP:
		as, err := a.TOTP.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, totpToAuthenticatorInfo(a))
		}

	case interaction.AuthenticatorTypeOOBOTP:
		as, err := a.OOBOTP.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, oobotpToAuthenticatorInfo(a))
		}

	case interaction.AuthenticatorTypeBearerToken:
		as, err := a.BearerToken.List(userID)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			ais = append(ais, bearerTokenToAuthenticatorInfo(a))
		}

	case interaction.AuthenticatorTypeRecoveryCode:
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

func (a *AuthenticatorAdaptor) New(userID string, spec interaction.AuthenticatorSpec, secret string) ([]*interaction.AuthenticatorInfo, error) {
	switch spec.Type {
	case interaction.AuthenticatorTypePassword:
		p, err := a.Password.New(userID, secret)
		if err != nil {
			return nil, err
		}
		return []*interaction.AuthenticatorInfo{passwordToAuthenticatorInfo(p)}, nil

	case interaction.AuthenticatorTypeTOTP:
		displayName, _ := spec.Props[interaction.AuthenticatorPropTOTPDisplayName].(string)
		t := a.TOTP.New(userID, displayName)
		return []*interaction.AuthenticatorInfo{totpToAuthenticatorInfo(t)}, nil

	case interaction.AuthenticatorTypeOOBOTP:
		channel := spec.Props[interaction.AuthenticatorPropOOBOTPChannelType].(string)
		var phone, email string
		switch authn.AuthenticatorOOBChannel(channel) {
		case authn.AuthenticatorOOBChannelSMS:
			phone = spec.Props[interaction.AuthenticatorPropOOBOTPPhone].(string)
		case authn.AuthenticatorOOBChannelEmail:
			email = spec.Props[interaction.AuthenticatorPropOOBOTPEmail].(string)
		}
		o := a.OOBOTP.New(userID, authn.AuthenticatorOOBChannel(channel), phone, email)
		return []*interaction.AuthenticatorInfo{oobotpToAuthenticatorInfo(o)}, nil

	case interaction.AuthenticatorTypeBearerToken:
		parentID := spec.Props[interaction.AuthenticatorPropBearerTokenParentID].(string)
		b := a.BearerToken.New(userID, parentID)
		return []*interaction.AuthenticatorInfo{bearerTokenToAuthenticatorInfo(b)}, nil

	case interaction.AuthenticatorTypeRecoveryCode:
		rs := a.RecoveryCode.Generate(userID)
		var ais []*interaction.AuthenticatorInfo
		for _, r := range rs {
			ais = append(ais, recoveryCodeToAuthenticatorInfo(r))
		}
		return ais, nil
	}

	panic("interaction_adaptors: unknown authenticator type " + spec.Type)
}

func (a *AuthenticatorAdaptor) Authenticate(userID string, spec interaction.AuthenticatorSpec, state *map[string]string, secret string) (*interaction.AuthenticatorInfo, error) {
	switch spec.Type {
	case interaction.AuthenticatorTypePassword:
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

	case interaction.AuthenticatorTypeTOTP:
		ts, err := a.TOTP.List(userID)
		if err != nil {
			return nil, err
		}

		t := a.TOTP.Authenticate(ts, secret)
		if t == nil {
			return nil, interaction.ErrInvalidCredentials
		}
		return totpToAuthenticatorInfo(t), nil

	case interaction.AuthenticatorTypeOOBOTP:
		if state == nil {
			return nil, interaction.ErrInvalidCredentials
		}
		id := (*state)[interaction.AuthenticatorStateOOBOTPID]
		code := (*state)[interaction.AuthenticatorStateOOBOTPCode]

		o, err := a.OOBOTP.Get(userID, id)
		if errors.Is(err, authenticator.ErrAuthenticatorNotFound) {
			return nil, interaction.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}

		if a.OOBOTP.Authenticate(o, code, secret) != nil {
			return nil, interaction.ErrInvalidCredentials
		}
		return oobotpToAuthenticatorInfo(o), nil

	case interaction.AuthenticatorTypeBearerToken:
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

	case interaction.AuthenticatorTypeRecoveryCode:
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
