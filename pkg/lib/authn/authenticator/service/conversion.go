package service

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
)

func passwordToAuthenticatorInfo(p *password.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:      authn.AuthenticatorTypePassword,
		Labels:    p.Labels,
		ID:        p.ID,
		UserID:    p.UserID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		Secret:    string(p.PasswordHash),
		Claims:    map[string]interface{}{},
		IsDefault: p.IsDefault,
		Kind:      authenticator.Kind(p.Kind),
	}
}

func passwordFromAuthenticatorInfo(a *authenticator.Info) *password.Authenticator {
	return &password.Authenticator{
		ID:           a.ID,
		Labels:       a.Labels,
		UserID:       a.UserID,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
		PasswordHash: []byte(a.Secret),
		IsDefault:    a.IsDefault,
		Kind:         string(a.Kind),
	}
}

func totpToAuthenticatorInfo(t *totp.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:      authn.AuthenticatorTypeTOTP,
		Labels:    t.Labels,
		ID:        t.ID,
		UserID:    t.UserID,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		Secret:    t.Secret,
		Claims: map[string]interface{}{
			authenticator.AuthenticatorClaimTOTPDisplayName: t.DisplayName,
		},
		IsDefault: t.IsDefault,
		Kind:      authenticator.Kind(t.Kind),
	}
}

func totpFromAuthenticatorInfo(a *authenticator.Info) *totp.Authenticator {
	return &totp.Authenticator{
		ID:          a.ID,
		Labels:      a.Labels,
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		Secret:      a.Secret,
		DisplayName: a.Claims[authenticator.AuthenticatorClaimTOTPDisplayName].(string),
		IsDefault:   a.IsDefault,
		Kind:        string(a.Kind),
	}
}

func oobotpToAuthenticatorInfo(o *oob.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:      authn.AuthenticatorTypeOOB,
		ID:        o.ID,
		Labels:    o.Labels,
		UserID:    o.UserID,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		Secret:    "",
		Claims: map[string]interface{}{
			authenticator.AuthenticatorClaimOOBOTPChannelType: string(o.Channel),
			authenticator.AuthenticatorClaimOOBOTPEmail:       o.Email,
			authenticator.AuthenticatorClaimOOBOTPPhone:       o.Phone,
		},
		IsDefault: o.IsDefault,
		Kind:      authenticator.Kind(o.Kind),
	}
}

func oobotpFromAuthenticatorInfo(a *authenticator.Info) *oob.Authenticator {
	return &oob.Authenticator{
		ID:        a.ID,
		Labels:    a.Labels,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Channel:   authn.AuthenticatorOOBChannel(a.Claims[authenticator.AuthenticatorClaimOOBOTPChannelType].(string)),
		Phone:     a.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string),
		Email:     a.Claims[authenticator.AuthenticatorClaimOOBOTPEmail].(string),
		IsDefault: a.IsDefault,
		Kind:      string(a.Kind),
	}
}
