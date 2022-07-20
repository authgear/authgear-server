package identity

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Info struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Type      model.IdentityType     `json:"type"`
	Claims    map[string]interface{} `json:"claims"`
}

func (i *Info) ToSpec() Spec {
	return Spec{Type: i.Type, Claims: i.Claims}
}

func (i *Info) ToRef() *model.IdentityRef {
	return &model.IdentityRef{
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
	case model.IdentityTypeLoginID:
		return nil
	case model.IdentityTypeOAuth:
		return nil
	case model.IdentityTypeAnonymous:
		return nil
	case model.IdentityTypeBiometric:
		return []string{model.AMRXBiometric}
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
	case model.IdentityTypeLoginID:
		displayID, _ := i.Claims[IdentityClaimLoginIDOriginalValue].(string)
		return displayID
	case model.IdentityTypeOAuth:
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
	case model.IdentityTypeAnonymous:
		displayID, _ := i.Claims[IdentityClaimAnonymousKeyID].(string)
		return displayID
	case model.IdentityTypeBiometric:
		displayID, _ := i.Claims[IdentityClaimBiometricKeyID].(string)
		return displayID
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
}

func (i *Info) StandardClaims() map[model.ClaimName]string {
	claims := map[model.ClaimName]string{}
	switch i.Type {
	case model.IdentityTypeLoginID:
		loginIDType, _ := i.Claims[IdentityClaimLoginIDType].(string)
		loginIDValue, _ := i.Claims[IdentityClaimLoginIDOriginalValue].(string)
		switch config.LoginIDKeyType(loginIDType) {
		case config.LoginIDKeyTypeEmail:
			claims[model.ClaimEmail] = loginIDValue
		case config.LoginIDKeyTypePhone:
			claims[model.ClaimPhoneNumber] = loginIDValue
		case config.LoginIDKeyTypeUsername:
			claims[model.ClaimPreferredUsername] = loginIDValue
		}
	case model.IdentityTypeOAuth:
		if email, ok := i.Claims[StandardClaimEmail].(string); ok {
			claims[model.ClaimEmail] = email
		}
	case model.IdentityTypeAnonymous:
		break
	case model.IdentityTypeBiometric:
		break
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
	return claims
}

func (i *Info) PrimaryAuthenticatorTypes() []model.AuthenticatorType {
	switch i.Type {
	case model.IdentityTypeLoginID:
		switch config.LoginIDKeyType(i.Claims[IdentityClaimLoginIDType].(string)) {
		case config.LoginIDKeyTypeUsername:
			return []model.AuthenticatorType{
				model.AuthenticatorTypePassword,
			}
		case config.LoginIDKeyTypeEmail:
			return []model.AuthenticatorType{
				model.AuthenticatorTypePassword,
				model.AuthenticatorTypeOOBEmail,
			}
		case config.LoginIDKeyTypePhone:
			return []model.AuthenticatorType{
				model.AuthenticatorTypePassword,
				model.AuthenticatorTypeOOBSMS,
			}
		default:
			panic(fmt.Sprintf("identity: unexpected login ID type: %s", i.Claims[IdentityClaimLoginIDType]))
		}
	case model.IdentityTypeOAuth:
		return nil
	case model.IdentityTypeAnonymous:
		return nil
	case model.IdentityTypeBiometric:
		return nil
	default:
		panic(fmt.Sprintf("identity: unexpected identity type: %s", i.Type))
	}
}

func (i *Info) ModifyDisabled(c *config.IdentityConfig) bool {
	switch i.Type {
	case model.IdentityTypeLoginID:
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
	case model.IdentityTypeOAuth:
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
	case model.IdentityTypeAnonymous:
		// modify_disabled is only applicable to login_id and oauth.
		// So we return false here.
		return false
	case model.IdentityTypeBiometric:
		// modify_disabled is only applicable to login_id and oauth.
		// So we return false here.
		return false
	default:
		panic(fmt.Sprintf("identity: unexpected identity type: %s", i.Type))
	}
}
