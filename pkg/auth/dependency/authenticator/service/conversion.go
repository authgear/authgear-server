package service

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/password"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func passwordToAuthenticatorInfo(p *password.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:   authn.AuthenticatorTypePassword,
		ID:     p.ID,
		UserID: p.UserID,
		Secret: string(p.PasswordHash),
		Props:  map[string]interface{}{},
	}
}

func passwordFromAuthenticatorInfo(a *authenticator.Info) *password.Authenticator {
	return &password.Authenticator{
		ID:           a.ID,
		UserID:       a.UserID,
		PasswordHash: []byte(a.Secret),
	}
}

func totpToAuthenticatorInfo(t *totp.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:   authn.AuthenticatorTypeTOTP,
		ID:     t.ID,
		UserID: t.UserID,
		Secret: t.Secret,
		Props: map[string]interface{}{
			authenticator.AuthenticatorPropTOTPDisplayName: t.DisplayName,
		},
	}
}

func totpFromAuthenticatorInfo(a *authenticator.Info) *totp.Authenticator {
	return &totp.Authenticator{
		ID:          a.ID,
		UserID:      a.UserID,
		Secret:      a.Secret,
		DisplayName: a.Props[authenticator.AuthenticatorPropTOTPDisplayName].(string),
	}
}

func oobotpToAuthenticatorInfo(o *oob.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:   authn.AuthenticatorTypeOOB,
		ID:     o.ID,
		UserID: o.UserID,
		Secret: "",
		Props: map[string]interface{}{
			authenticator.AuthenticatorPropOOBOTPID:          o.ID,
			authenticator.AuthenticatorPropOOBOTPChannelType: string(o.Channel),
			authenticator.AuthenticatorPropOOBOTPEmail:       o.Email,
			authenticator.AuthenticatorPropOOBOTPPhone:       o.Phone,
		},
	}
}

func oobotpFromAuthenticatorInfo(a *authenticator.Info) *oob.Authenticator {
	return &oob.Authenticator{
		ID:      a.ID,
		UserID:  a.UserID,
		Channel: authn.AuthenticatorOOBChannel(a.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string)),
		Phone:   a.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string),
		Email:   a.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string),
	}
}
