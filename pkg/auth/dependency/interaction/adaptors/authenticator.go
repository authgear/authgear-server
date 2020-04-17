package adaptors

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/bearertoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/recoverycode"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
)

type PasswordAuthenticatorProvider interface {
	Get(userID, id string) (*password.Authenticator, error)
}

type TOTPAuthenticatorProvider interface {
	Get(userID, id string) (*totp.Authenticator, error)
}

type OOBOTPAuthenticatorProvider interface {
	Get(userID, id string) (*oob.Authenticator, error)
}

type BearerTokenAuthenticatorProvider interface {
	Get(userID, id string) (*bearertoken.Authenticator, error)
}

type RecoveryCodeAuthenticatorProvider interface {
	Get(userID, id string) (*recoverycode.Authenticator, error)
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
