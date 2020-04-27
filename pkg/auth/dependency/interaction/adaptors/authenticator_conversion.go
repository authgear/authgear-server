package adaptors

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/bearertoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/recoverycode"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func passwordToAuthenticatorInfo(p *password.Authenticator) *interaction.AuthenticatorInfo {
	return &interaction.AuthenticatorInfo{
		Type:          interaction.AuthenticatorTypePassword,
		ID:            p.ID,
		Secret:        string(p.PasswordHash),
		Props:         map[string]interface{}{},
		Authenticator: p,
	}
}

func passwordFromAuthenticatorInfo(userID string, a *interaction.AuthenticatorInfo) *password.Authenticator {
	return &password.Authenticator{
		ID:           a.ID,
		UserID:       userID,
		PasswordHash: []byte(a.Secret),
	}
}

func totpToAuthenticatorInfo(t *totp.Authenticator) *interaction.AuthenticatorInfo {
	return &interaction.AuthenticatorInfo{
		Type:   interaction.AuthenticatorTypeTOTP,
		ID:     t.ID,
		Secret: t.Secret,
		Props: map[string]interface{}{
			interaction.AuthenticatorPropTOTPDisplayName: t.DisplayName,
		},
		Authenticator: t,
	}
}

func totpFromAuthenticatorInfo(userID string, a *interaction.AuthenticatorInfo) *totp.Authenticator {
	return &totp.Authenticator{
		ID:          a.ID,
		UserID:      userID,
		Secret:      a.Secret,
		DisplayName: a.Props[interaction.AuthenticatorPropTOTPDisplayName].(string),
	}
}

func oobotpToAuthenticatorInfo(o *oob.Authenticator) *interaction.AuthenticatorInfo {
	return &interaction.AuthenticatorInfo{
		Type:   interaction.AuthenticatorTypeOOBOTP,
		ID:     o.ID,
		Secret: "",
		Props: map[string]interface{}{
			interaction.AuthenticatorPropOOBOTPChannelType: o.Channel,
			interaction.AuthenticatorPropOOBOTPEmail:       o.Email,
			interaction.AuthenticatorPropOOBOTPPhone:       o.Phone,
		},
		Authenticator: o,
	}
}

func oobotpFromAuthenticatorInfo(userID string, a *interaction.AuthenticatorInfo) *oob.Authenticator {
	return &oob.Authenticator{
		ID:      a.ID,
		UserID:  userID,
		Channel: authn.AuthenticatorOOBChannel(a.Props[interaction.AuthenticatorPropOOBOTPChannelType].(string)),
		Phone:   a.Props[interaction.AuthenticatorPropOOBOTPPhone].(string),
		Email:   a.Props[interaction.AuthenticatorPropOOBOTPEmail].(string),
	}
}

func bearerTokenToAuthenticatorInfo(b *bearertoken.Authenticator) *interaction.AuthenticatorInfo {
	return &interaction.AuthenticatorInfo{
		Type:   interaction.AuthenticatorTypeBearerToken,
		ID:     b.ID,
		Secret: b.Token,
		Props: map[string]interface{}{
			interaction.AuthenticatorPropBearerTokenParentID: b.ParentID,
		},
		Authenticator: b,
	}
}

func bearerTokenFromAuthenticatorInfo(userID string, a *interaction.AuthenticatorInfo) *bearertoken.Authenticator {
	return &bearertoken.Authenticator{
		ID:       a.ID,
		UserID:   userID,
		ParentID: a.Props[interaction.AuthenticatorPropBearerTokenParentID].(string),
		Token:    a.Secret,
	}
}

func recoveryCodeToAuthenticatorInfo(r *recoverycode.Authenticator) *interaction.AuthenticatorInfo {
	return &interaction.AuthenticatorInfo{
		Type:          interaction.AuthenticatorTypeBearerToken,
		ID:            r.ID,
		Secret:        r.Code,
		Props:         map[string]interface{}{},
		Authenticator: r,
	}
}

func recoveryCodeFromAuthenticatorInfo(userID string, a *interaction.AuthenticatorInfo) *recoverycode.Authenticator {
	return &recoverycode.Authenticator{
		ID:     a.ID,
		UserID: userID,
		Code:   a.Secret,
	}
}
