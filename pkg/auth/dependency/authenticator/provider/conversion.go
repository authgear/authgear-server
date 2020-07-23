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
		Type:   authn.AuthenticatorTypePassword,
		ID:     p.ID,
		UserID: p.UserID,
		Secret: string(p.PasswordHash),
		Props:  map[string]interface{}{},
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
		UserID: t.UserID,
		Secret: t.Secret,
		Props: map[string]interface{}{
			authenticator.AuthenticatorPropTOTPDisplayName: t.DisplayName,
		},
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
		UserID: o.UserID,
		Secret: "",
		Props: map[string]interface{}{
			authenticator.AuthenticatorPropOOBOTPID:          o.ID,
			authenticator.AuthenticatorPropOOBOTPChannelType: string(o.Channel),
			authenticator.AuthenticatorPropOOBOTPEmail:       o.Email,
			authenticator.AuthenticatorPropOOBOTPPhone:       o.Phone,
			authenticator.AuthenticatorPropOOBOTPIdentityID:  o.IdentityID,
		},
	}
}

func oobotpFromAuthenticatorInfo(userID string, a *authenticator.Info) *oob.Authenticator {
	return &oob.Authenticator{
		ID:         a.ID,
		UserID:     userID,
		Channel:    authn.AuthenticatorOOBChannel(a.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string)),
		Phone:      a.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string),
		Email:      a.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string),
		IdentityID: a.Props[authenticator.AuthenticatorPropOOBOTPIdentityID].(*string),
	}
}

func bearerTokenToAuthenticatorInfo(b *bearertoken.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:   authn.AuthenticatorTypeBearerToken,
		ID:     b.ID,
		UserID: b.UserID,
		Secret: b.Token,
		Props: map[string]interface{}{
			authenticator.AuthenticatorPropBearerTokenParentID: b.ParentID,
		},
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
		Type:   authn.AuthenticatorTypeBearerToken,
		ID:     r.ID,
		UserID: r.UserID,
		Secret: r.Code,
		Props:  map[string]interface{}{},
	}
}

func recoveryCodeFromAuthenticatorInfo(userID string, a *authenticator.Info) *recoverycode.Authenticator {
	return &recoverycode.Authenticator{
		ID:     a.ID,
		UserID: userID,
		Code:   a.Secret,
	}
}
