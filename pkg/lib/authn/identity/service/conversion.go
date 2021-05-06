package service

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
)

func loginIDToIdentityInfo(l *loginid.Identity) *identity.Info {
	claims := map[string]interface{}{
		identity.IdentityClaimLoginIDKey:           l.LoginIDKey,
		identity.IdentityClaimLoginIDType:          string(l.LoginIDType),
		identity.IdentityClaimLoginIDValue:         l.LoginID,
		identity.IdentityClaimLoginIDOriginalValue: l.OriginalLoginID,
		identity.IdentityClaimLoginIDUniqueKey:     l.UniqueKey,
	}
	for k, v := range l.Claims {
		claims[k] = v
	}

	return &identity.Info{
		UserID:    l.UserID,
		Labels:    l.Labels,
		ID:        l.ID,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
		Type:      authn.IdentityTypeLoginID,
		Claims:    claims,
	}
}

func loginIDFromIdentityInfo(i *identity.Info) *loginid.Identity {
	l := &loginid.Identity{
		ID:        i.ID,
		Labels:    i.Labels,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		UserID:    i.UserID,
		Claims:    make(map[string]interface{}),
	}
	for k, v := range i.Claims {
		switch k {
		case identity.IdentityClaimLoginIDKey:
			l.LoginIDKey = v.(string)
		case identity.IdentityClaimLoginIDType:
			l.LoginIDType = config.LoginIDKeyType(v.(string))
		case identity.IdentityClaimLoginIDValue:
			l.LoginID = v.(string)
		case identity.IdentityClaimLoginIDOriginalValue:
			l.OriginalLoginID = v.(string)
		case identity.IdentityClaimLoginIDUniqueKey:
			l.UniqueKey = v.(string)
		default:
			l.Claims[k] = v.(string)
		}
	}
	return l
}

func oauthFromIdentityInfo(i *identity.Info) *oauth.Identity {
	o := &oauth.Identity{
		ID:        i.ID,
		Labels:    i.Labels,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		UserID:    i.UserID,
		Claims:    map[string]interface{}{},
	}
	for k, v := range i.Claims {
		switch k {
		case identity.IdentityClaimOAuthProviderKeys:
			o.ProviderID = config.NewProviderID(v.(map[string]interface{}))
		case identity.IdentityClaimOAuthSubjectID:
			o.ProviderSubjectID = v.(string)
		case identity.IdentityClaimOAuthProfile:
			o.UserProfile = v.(map[string]interface{})
		default:
			o.Claims[k] = v
		}
	}
	return o
}

func anonymousToIdentityInfo(a *anonymous.Identity) *identity.Info {
	claims := map[string]interface{}{
		identity.IdentityClaimAnonymousKeyID: a.KeyID,
		identity.IdentityClaimAnonymousKey:   string(a.Key),
	}

	return &identity.Info{
		ID:        a.ID,
		Labels:    a.Labels,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Type:      authn.IdentityTypeAnonymous,
		Claims:    claims,
	}
}

func anonymousFromIdentityInfo(i *identity.Info) *anonymous.Identity {
	a := &anonymous.Identity{
		ID:        i.ID,
		Labels:    i.Labels,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		UserID:    i.UserID,
	}
	for k, v := range i.Claims {
		switch k {
		case identity.IdentityClaimAnonymousKeyID:
			a.KeyID = v.(string)
		case identity.IdentityClaimAnonymousKey:
			a.Key = []byte(v.(string))
		}
	}
	return a
}

func biometricToIdentityInfo(b *biometric.Identity) *identity.Info {
	claims := map[string]interface{}{
		identity.IdentityClaimBiometricKeyID:               b.KeyID,
		identity.IdentityClaimBiometricKey:                 string(b.Key),
		identity.IdentityClaimBiometricDeviceInfo:          b.DeviceInfo,
		identity.IdentityClaimBiometricFormattedDeviceInfo: deviceinfo.Format(b.DeviceInfo),
	}

	return &identity.Info{
		ID:        b.ID,
		Labels:    b.Labels,
		UserID:    b.UserID,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		Type:      authn.IdentityTypeBiometric,
		Claims:    claims,
	}
}

func biometricFromIdentityInfo(i *identity.Info) *biometric.Identity {
	b := &biometric.Identity{
		ID:        i.ID,
		Labels:    i.Labels,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		UserID:    i.UserID,
	}
	for k, v := range i.Claims {
		switch k {
		case identity.IdentityClaimBiometricKeyID:
			b.KeyID = v.(string)
		case identity.IdentityClaimBiometricKey:
			b.Key = []byte(v.(string))
		case identity.IdentityClaimBiometricDeviceInfo:
			b.DeviceInfo = v.(map[string]interface{})
		}
	}
	return b
}
