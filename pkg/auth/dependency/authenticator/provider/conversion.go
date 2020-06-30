package provider

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/bearertoken"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/password"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/recoverycode"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func passwordToAuthenticatorInfo(p *password.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:          authn.AuthenticatorTypePassword,
		ID:            p.ID,
		Secret:        string(p.PasswordHash),
		Props:         map[string]interface{}{},
		Authenticator: p,
	}
}

func passwordFromAuthenticatorInfo(userID string, a *authenticator.Info) *password.Authenticator {
	return &password.Authenticator{
		ID:           a.ID,
		UserID:       userID,
		PasswordHash: []byte(a.Secret),
	}
}

func totpToAuthenticatorInfo(t *totp.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:   authn.AuthenticatorTypeTOTP,
		ID:     t.ID,
		Secret: t.Secret,
		Props: map[string]interface{}{
			authenticator.AuthenticatorPropTOTPDisplayName: t.DisplayName,
		},
		Authenticator: t,
	}
}

func totpFromAuthenticatorInfo(userID string, a *authenticator.Info) *totp.Authenticator {
	return &totp.Authenticator{
		ID:          a.ID,
		UserID:      userID,
		Secret:      a.Secret,
		DisplayName: a.Props[authenticator.AuthenticatorPropTOTPDisplayName].(string),
	}
}

func oobotpToAuthenticatorInfo(o *oob.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:   authn.AuthenticatorTypeOOB,
		ID:     o.ID,
		Secret: "",
		Props: map[string]interface{}{
			authenticator.AuthenticatorPropOOBOTPID:          o.ID,
			authenticator.AuthenticatorPropOOBOTPChannelType: string(o.Channel),
			authenticator.AuthenticatorPropOOBOTPEmail:       o.Email,
			authenticator.AuthenticatorPropOOBOTPPhone:       o.Phone,
		},
		Authenticator: o,
	}
}

func oobotpFromAuthenticatorInfo(userID string, a *authenticator.Info) *oob.Authenticator {
	return &oob.Authenticator{
		ID:      a.ID,
		UserID:  userID,
		Channel: authn.AuthenticatorOOBChannel(a.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string)),
		Phone:   a.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string),
		Email:   a.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string),
	}
}

func bearerTokenToAuthenticatorInfo(b *bearertoken.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:   authn.AuthenticatorTypeBearerToken,
		ID:     b.ID,
		Secret: b.Token,
		Props: map[string]interface{}{
			authenticator.AuthenticatorPropBearerTokenParentID: b.ParentID,
		},
		Authenticator: b,
	}
}

func bearerTokenFromAuthenticatorInfo(userID string, a *authenticator.Info) *bearertoken.Authenticator {
	return &bearertoken.Authenticator{
		ID:       a.ID,
		UserID:   userID,
		ParentID: a.Props[authenticator.AuthenticatorPropBearerTokenParentID].(string),
		Token:    a.Secret,
	}
}

func recoveryCodeToAuthenticatorInfo(r *recoverycode.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:          authn.AuthenticatorTypeBearerToken,
		ID:            r.ID,
		Secret:        r.Code,
		Props:         map[string]interface{}{},
		Authenticator: r,
	}
}

func recoveryCodeFromAuthenticatorInfo(userID string, a *authenticator.Info) *recoverycode.Authenticator {
	return &recoverycode.Authenticator{
		ID:     a.ID,
		UserID: userID,
		Code:   a.Secret,
	}
}
