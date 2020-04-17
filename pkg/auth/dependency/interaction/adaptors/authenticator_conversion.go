package adaptors

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/bearertoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/recoverycode"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
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

func bearerTokenToAuthenticatorInfo(b *bearertoken.Authenticator) *interaction.AuthenticatorInfo {
	return &interaction.AuthenticatorInfo{
		Type:          interaction.AuthenticatorTypeBearerToken,
		ID:            b.ID,
		Secret:        b.Token,
		Props:         map[string]interface{}{},
		Authenticator: b,
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
