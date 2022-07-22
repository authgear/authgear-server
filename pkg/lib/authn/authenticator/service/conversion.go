package service

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
)

func passwordToAuthenticatorInfo(p *password.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:      model.AuthenticatorTypePassword,
		ID:        p.ID,
		UserID:    p.UserID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		Claims: map[string]interface{}{
			authenticator.AuthenticatorClaimPasswordPasswordHash: p.PasswordHash,
		},
		IsDefault: p.IsDefault,
		Kind:      authenticator.Kind(p.Kind),
	}
}

func passwordFromAuthenticatorInfo(a *authenticator.Info) *password.Authenticator {
	return &password.Authenticator{
		ID:           a.ID,
		UserID:       a.UserID,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
		PasswordHash: a.Claims[authenticator.AuthenticatorClaimPasswordPasswordHash].([]byte),
		IsDefault:    a.IsDefault,
		Kind:         string(a.Kind),
	}
}

func totpToAuthenticatorInfo(t *totp.Authenticator) *authenticator.Info {
	return &authenticator.Info{
		Type:      model.AuthenticatorTypeTOTP,
		ID:        t.ID,
		UserID:    t.UserID,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		Claims: map[string]interface{}{
			authenticator.AuthenticatorClaimTOTPDisplayName: t.DisplayName,
			authenticator.AuthenticatorClaimTOTPSecret:      t.Secret,
		},
		IsDefault: t.IsDefault,
		Kind:      authenticator.Kind(t.Kind),
	}
}

func totpFromAuthenticatorInfo(a *authenticator.Info) *totp.Authenticator {
	return &totp.Authenticator{
		ID:          a.ID,
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		Secret:      a.Claims[authenticator.AuthenticatorClaimTOTPSecret].(string),
		DisplayName: a.Claims[authenticator.AuthenticatorClaimTOTPDisplayName].(string),
		IsDefault:   a.IsDefault,
		Kind:        string(a.Kind),
	}
}

func oobotpToAuthenticatorInfo(o *oob.Authenticator) *authenticator.Info {
	info := &authenticator.Info{
		Type:      o.OOBAuthenticatorType,
		ID:        o.ID,
		UserID:    o.UserID,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		Claims:    map[string]interface{}{},
		IsDefault: o.IsDefault,
		Kind:      authenticator.Kind(o.Kind),
	}
	switch o.OOBAuthenticatorType {
	case model.AuthenticatorTypeOOBSMS:
		info.Claims[authenticator.AuthenticatorClaimOOBOTPPhone] = o.Phone
	case model.AuthenticatorTypeOOBEmail:
		info.Claims[authenticator.AuthenticatorClaimOOBOTPEmail] = o.Email
	default:
		panic("authenticator: incompatible authenticator type for oob: " + o.OOBAuthenticatorType)
	}
	return info
}

func oobotpFromAuthenticatorInfo(a *authenticator.Info) *oob.Authenticator {
	phone, _ := a.Claims[authenticator.AuthenticatorClaimOOBOTPPhone].(string)
	email, _ := a.Claims[authenticator.AuthenticatorClaimOOBOTPEmail].(string)
	return &oob.Authenticator{
		ID:                   a.ID,
		UserID:               a.UserID,
		CreatedAt:            a.CreatedAt,
		UpdatedAt:            a.UpdatedAt,
		OOBAuthenticatorType: a.Type,
		Phone:                phone,
		Email:                email,
		IsDefault:            a.IsDefault,
		Kind:                 string(a.Kind),
	}
}
