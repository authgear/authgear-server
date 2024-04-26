package identity

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Info struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Type      model.IdentityType `json:"type"`

	LoginID   *LoginID   `json:"login_id,omitempty"`
	OAuth     *OAuth     `json:"oauth,omitempty"`
	Anonymous *Anonymous `json:"anonymous,omitempty"`
	Biometric *Biometric `json:"biometric,omitempty"`
	Passkey   *Passkey   `json:"passkey,omitempty"`
	SIWE      *SIWE      `json:"siwe,omitempty"`
}

func (i *Info) ToSpec() Spec {
	switch i.Type {
	case model.IdentityTypeLoginID:
		return Spec{
			Type: i.Type,
			LoginID: &LoginIDSpec{
				Key:   i.LoginID.LoginIDKey,
				Type:  i.LoginID.LoginIDType,
				Value: i.LoginID.LoginID,
			},
		}
	case model.IdentityTypeOAuth:
		return Spec{
			Type: i.Type,
			OAuth: &OAuthSpec{
				ProviderID:     i.OAuth.ProviderID,
				SubjectID:      i.OAuth.ProviderSubjectID,
				RawProfile:     i.OAuth.UserProfile,
				StandardClaims: i.OAuth.Claims,
			},
		}
	case model.IdentityTypeAnonymous:
		return Spec{
			Type: i.Type,
			Anonymous: &AnonymousSpec{
				KeyID:              i.Anonymous.KeyID,
				Key:                string(i.Anonymous.Key),
				ExistingUserID:     i.Anonymous.UserID,
				ExistingIdentityID: i.Anonymous.ID,
			},
		}
	case model.IdentityTypeBiometric:
		return Spec{
			Type: i.Type,
			Biometric: &BiometricSpec{
				KeyID:      i.Biometric.KeyID,
				Key:        string(i.Biometric.Key),
				DeviceInfo: i.Biometric.DeviceInfo,
			},
		}
	case model.IdentityTypePasskey:
		return Spec{
			Type: i.Type,
			Passkey: &PasskeySpec{
				AttestationResponse: i.Passkey.AttestationResponse,
			},
		}
	case model.IdentityTypeSIWE:
		return Spec{
			Type: i.Type,
			SIWE: &SIWESpec{
				Message:   i.SIWE.Data.Message,
				Signature: i.SIWE.Data.Signature,
			},
		}
	default:
		panic("identity: unknown identity type: " + i.Type)
	}
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
	case model.IdentityTypePasskey:
		return nil
	case model.IdentityTypeSIWE:
		return nil
	default:
		panic("identity: unknown identity type: " + i.Type)
	}
}

func (i *Info) ToModel() model.Identity {
	claims := make(map[string]interface{})
	switch i.Type {
	case model.IdentityTypeLoginID:
		for k, v := range i.LoginID.Claims {
			claims[k] = v
		}
		claims[IdentityClaimLoginIDType] = i.LoginID.LoginIDType
		claims[IdentityClaimLoginIDKey] = i.LoginID.LoginIDKey
		claims[IdentityClaimLoginIDOriginalValue] = i.LoginID.OriginalLoginID
		claims[IdentityClaimLoginIDValue] = i.LoginID.LoginID

	case model.IdentityTypeOAuth:
		for k, v := range i.OAuth.Claims {
			claims[k] = v
		}
		claims[IdentityClaimOAuthProviderType] = i.OAuth.ProviderID.Type
		claims[IdentityClaimOAuthSubjectID] = i.OAuth.ProviderSubjectID
		claims[IdentityClaimOAuthProfile] = i.OAuth.UserProfile
		claims[IdentityClaimOAuthProviderAlias] = i.OAuth.ProviderAlias

	case model.IdentityTypeAnonymous:
		claims[IdentityClaimAnonymousKeyID] = i.Anonymous.KeyID

	case model.IdentityTypeBiometric:
		claims[IdentityClaimBiometricKeyID] = i.Biometric.KeyID
		claims[IdentityClaimBiometricDeviceInfo] = i.Biometric.DeviceInfo
		claims[IdentityClaimBiometricFormattedDeviceInfo] = i.Biometric.FormattedDeviceInfo()

	case model.IdentityTypePasskey:
		claims[IdentityClaimPasskeyCredentialID] = i.Passkey.CredentialID
		claims[IdentityClaimPasskeyDisplayName] = i.Passkey.CreationOptions.PublicKey.User.DisplayName

	case model.IdentityTypeSIWE:
		claims[IdentityClaimSIWEAddress] = i.SIWE.Address
		claims[IdentityClaimSIWEChainID] = i.SIWE.ChainID

	default:
		panic("identity: unknown identity type: " + i.Type)
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
// If it is a passkey identity, the name is returned.
// If it is a SIWE identity, EIP681 of the address and chainID is returned
func (i *Info) DisplayID() string {
	switch i.Type {
	case model.IdentityTypeLoginID:
		return i.LoginID.OriginalLoginID
	case model.IdentityTypeOAuth:
		if email, ok := i.OAuth.Claims[StandardClaimEmail].(string); ok {
			return email
		}
		if phoneNumber, ok := i.OAuth.Claims[StandardClaimPhoneNumber].(string); ok {
			return phoneNumber
		}
		if preferredUsername, ok := i.OAuth.Claims[StandardClaimPreferredUsername].(string); ok {
			return preferredUsername
		}
		return ""
	case model.IdentityTypeAnonymous:
		return i.Anonymous.KeyID
	case model.IdentityTypeBiometric:
		return i.Biometric.KeyID
	case model.IdentityTypePasskey:
		return i.Passkey.CreationOptions.PublicKey.User.DisplayName
	case model.IdentityTypeSIWE:
		eip681, err := i.SIWE.ToERC681()
		if err != nil {
			panic(fmt.Errorf("identity: failed to parse SIWE identity: %w", err))
		}
		return eip681.URL().String()
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
}

func (i *Info) IdentityAwareStandardClaims() map[model.ClaimName]string {
	claims := map[model.ClaimName]string{}
	switch i.Type {
	case model.IdentityTypeLoginID:
		loginIDType := i.LoginID.LoginIDType
		loginIDValue := i.LoginID.LoginID
		if claimName, ok := model.GetLoginIDKeyTypeClaim(loginIDType); ok {
			claims[claimName] = loginIDValue
		}
	case model.IdentityTypeOAuth:
		if email, ok := i.OAuth.Claims[StandardClaimEmail].(string); ok {
			claims[model.ClaimEmail] = email
		}
	case model.IdentityTypeAnonymous:
		break
	case model.IdentityTypeBiometric:
		break
	case model.IdentityTypePasskey:
		break
	case model.IdentityTypeSIWE:
		break
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
	return claims
}

func (i *Info) AllStandardClaims() map[string]interface{} {
	claims := make(map[string]interface{})
	switch i.Type {
	case model.IdentityTypeLoginID:
		return i.LoginID.Claims
	case model.IdentityTypeOAuth:
		return i.OAuth.Claims
	case model.IdentityTypeAnonymous:
		break
	case model.IdentityTypeBiometric:
		break
	case model.IdentityTypePasskey:
		break
	case model.IdentityTypeSIWE:
		break
	default:
		panic(fmt.Errorf("identity: unexpected identity type %v", i.Type))
	}
	return claims
}

func (i *Info) PrimaryAuthenticatorTypes() []model.AuthenticatorType {
	var loginIDKeyType model.LoginIDKeyType
	if i.Type == model.IdentityTypeLoginID {
		loginIDKeyType = i.LoginID.LoginIDType
	}
	return i.Type.PrimaryAuthenticatorTypes(loginIDKeyType)
}

func (i *Info) ModifyDisabled(c *config.IdentityConfig) bool {
	switch i.Type {
	case model.IdentityTypeLoginID:
		loginIDKey := i.LoginID.LoginIDKey
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
		alias := i.OAuth.ProviderAlias
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
	case model.IdentityTypePasskey:
		// modify_disabled is only applicable to login_id and oauth.
		// So we return false here.
		return false
	case model.IdentityTypeSIWE:
		// modify_disabled is only applicable to login_id and oauth.
		// So we return false here.
		return false
	default:
		panic(fmt.Sprintf("identity: unexpected identity type: %s", i.Type))
	}
}

func (i *Info) UpdateUserID(newUserID string) *Info {
	i.UserID = newUserID
	switch i.Type {
	case model.IdentityTypeLoginID:
		i.LoginID.UserID = newUserID
	case model.IdentityTypeOAuth:
		i.OAuth.UserID = newUserID
	case model.IdentityTypeBiometric:
		i.Biometric.UserID = newUserID
	case model.IdentityTypePasskey:
		i.Passkey.UserID = newUserID
	case model.IdentityTypeSIWE:
		fallthrough
	case model.IdentityTypeAnonymous:
		fallthrough
	default:
		panic(fmt.Errorf("identity: identity type %v does not support updating user ID", i.Type))
	}
	return i
}
