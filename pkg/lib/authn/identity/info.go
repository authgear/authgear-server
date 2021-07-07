package identity

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type Info struct {
	ID        string                 `json:"id"`
	Labels    map[string]interface{} `json:"labels"`
	UserID    string                 `json:"user_id"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Type      authn.IdentityType     `json:"type"`
	Claims    map[string]interface{} `json:"claims"`
}

func (i *Info) ToSpec() Spec {
	return Spec{Type: i.Type, Claims: i.Claims}
}

func (i *Info) ToRef() *Ref {
	return &Ref{
		Meta: model.Meta{
			ID:        i.ID,
			CreatedAt: i.CreatedAt,
			UpdatedAt: i.UpdatedAt,
		},
		UserID: i.UserID,
		Type:   i.Type,
	}
}

func (i *Info) GetMeta() model.Meta {
	return model.Meta{
		ID:        i.ID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}

func (i *Info) AMR() []string {
	switch i.Type {
	case authn.IdentityTypeLoginID:
		return nil
	case authn.IdentityTypeOAuth:
		return nil
	case authn.IdentityTypeAnonymous:
		return nil
	case authn.IdentityTypeBiometric:
		return []string{authn.AMRXBiometric}
	default:
		panic("identity: unknown identity type: " + i.Type)
	}
}

func (i *Info) ToModel() model.Identity {
	claims := make(map[string]interface{})
	for key, value := range i.Claims {
		switch key {
		// It contains client_id, tenant or team_id, which should not
		// be exposed to clients.
		case IdentityClaimOAuthProviderKeys:
			continue

		// It contains OIDC standard claims, which is already exposed
		// as top-level claims.
		case IdentityClaimOAuthClaims:
			continue

		// It is a implementation details of login ID normalization,
		// so it should not be used by clients.
		case IdentityClaimLoginIDUniqueKey:
			continue

		// It is not useful to clients, since key ID should be
		// sufficient to identify a key.
		case IdentityClaimAnonymousKey:
			continue

		}
		claims[key] = value
	}

	return model.Identity{
		Meta:   i.GetMeta(),
		Type:   string(i.Type),
		Claims: claims,
	}
}

// DisplayID returns a string that is suitable for the owner to identify the identity.
// If it is a Login ID identity, the original login ID value is returned.
// If it is a OAuth identity, email, phone_number or preferred_username is returned.
// If it is a anonymous identity, the kid is returned.
// If it is a biometric identity, the kid is returned.
func (i *Info) DisplayID() string {
	switch i.Type {
	case authn.IdentityTypeLoginID:
		displayID, _ := i.Claims[IdentityClaimLoginIDOriginalValue].(string)
		return displayID
	case authn.IdentityTypeOAuth:
		if email, ok := i.Claims[StandardClaimEmail].(string); ok {
			return email
		}
		if phoneNumber, ok := i.Claims[StandardClaimPhoneNumber].(string); ok {
			return phoneNumber
		}
		if preferredUsername, ok := i.Claims[StandardClaimPreferredUsername].(string); ok {
			return preferredUsername
		}
		return ""
	case authn.IdentityTypeAnonymous:
		displayID, _ := i.Claims[IdentityClaimAnonymousKeyID].(string)
		return displayID
	case authn.IdentityTypeBiometric:
		displayID, _ := i.Claims[IdentityClaimBiometricKeyID].(string)
		return displayID
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
}

func (i *Info) StandardClaims() map[authn.ClaimName]string {
	claims := map[authn.ClaimName]string{}
	switch i.Type {
	case authn.IdentityTypeLoginID:
		loginIDType, _ := i.Claims[IdentityClaimLoginIDType].(string)
		loginIDValue, _ := i.Claims[IdentityClaimLoginIDOriginalValue].(string)
		switch config.LoginIDKeyType(loginIDType) {
		case config.LoginIDKeyTypeEmail:
			claims[authn.ClaimEmail] = loginIDValue
		case config.LoginIDKeyTypePhone:
			claims[authn.ClaimPhoneNumber] = loginIDValue
		case config.LoginIDKeyTypeUsername:
			claims[authn.ClaimPreferredUsername] = loginIDValue
		}
	case authn.IdentityTypeOAuth:
		if email, ok := i.Claims[StandardClaimEmail].(string); ok {
			claims[authn.ClaimEmail] = email
		}
	case authn.IdentityTypeAnonymous:
		break
	case authn.IdentityTypeBiometric:
		break
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
	return claims
}

func (i *Info) PrimaryAuthenticatorTypes() []authn.AuthenticatorType {
	switch i.Type {
	case authn.IdentityTypeLoginID:
		switch config.LoginIDKeyType(i.Claims[IdentityClaimLoginIDType].(string)) {
		case config.LoginIDKeyTypeUsername:
			return []authn.AuthenticatorType{
				authn.AuthenticatorTypePassword,
			}
		case config.LoginIDKeyTypeEmail:
			return []authn.AuthenticatorType{
				authn.AuthenticatorTypePassword,
				authn.AuthenticatorTypeOOBEmail,
			}
		case config.LoginIDKeyTypePhone:
			return []authn.AuthenticatorType{
				authn.AuthenticatorTypePassword,
				authn.AuthenticatorTypeOOBSMS,
			}
		default:
			panic(fmt.Sprintf("identity: unexpected login ID type: %s", i.Claims[IdentityClaimLoginIDType]))
		}
	case authn.IdentityTypeOAuth:
		return nil
	case authn.IdentityTypeAnonymous:
		return nil
	case authn.IdentityTypeBiometric:
		return nil
	default:
		panic(fmt.Sprintf("identity: unexpected identity type: %s", i.Type))
	}
}

func (i *Info) CanHaveMFA() bool {
	switch i.Type {
	case authn.IdentityTypeLoginID:
		return true
	case authn.IdentityTypeOAuth:
		return false
	case authn.IdentityTypeAnonymous:
		return false
	case authn.IdentityTypeBiometric:
		return false
	default:
		panic(fmt.Sprintf("identity: unexpected identity type: %s", i.Type))
	}
}

func (i *Info) ModifyDisabled(c *config.IdentityConfig) bool {
	switch i.Type {
	case authn.IdentityTypeLoginID:
		loginIDKey := i.Claims[IdentityClaimLoginIDKey].(string)
		var keyConfig *config.LoginIDKeyConfig
		for _, kc := range c.LoginID.Keys {
			if kc.Key == loginIDKey {
				kcc := kc
				keyConfig = &kcc
			}
		}
		if keyConfig == nil {
			return true
		}
		return *keyConfig.ModifyDisabled
	case authn.IdentityTypeOAuth:
		alias := i.Claims[IdentityClaimOAuthProviderAlias].(string)
		var providerConfig *config.OAuthSSOProviderConfig
		for _, pc := range c.OAuth.Providers {
			if pc.Alias == alias {
				pcc := pc
				providerConfig = &pcc
			}
		}
		if providerConfig == nil {
			return true
		}
		return *providerConfig.ModifyDisabled
	case authn.IdentityTypeAnonymous:
		// modify_disabled is only applicable to login_id and oauth.
		// So we return false here.
		return false
	case authn.IdentityTypeBiometric:
		// modify_disabled is only applicable to login_id and oauth.
		// So we return false here.
		return false
	default:
		panic(fmt.Sprintf("identity: unexpected identity type: %s", i.Type))
	}
}

func (i *Info) FillDetails(err error) error {
	details := errorutil.Details{
		"IdentityType": apierrors.APIErrorDetail.Value(i.Type),
	}
	switch i.Type {
	case authn.IdentityTypeLoginID:
		details["LoginIDType"] = apierrors.APIErrorDetail.Value(i.Claims[IdentityClaimLoginIDType])
	}
	return errorutil.WithDetails(err, details)
}
