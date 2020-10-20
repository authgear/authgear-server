package config

import (
	"bytes"
	"encoding/json"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Config struct {
	AppConfig    *AppConfig
	SecretConfig *SecretConfig
}

type AppID string

var _ = Schema.Add("AppConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"id": { "type": "string" },
		"http": { "$ref": "#/$defs/HTTPConfig" },
		"database": { "$ref": "#/$defs/DatabaseConfig" },
		"redis": { "$ref": "#/$defs/RedisConfig" },
		"hook": { "$ref": "#/$defs/HookConfig" },
		"ui": { "$ref": "#/$defs/UIConfig" },
		"localization": { "$ref": "#/$defs/LocalizationConfig" },
		"messaging": { "$ref": "#/$defs/MessagingConfig" },
		"authentication": { "$ref": "#/$defs/AuthenticationConfig" },
		"session": { "$ref": "#/$defs/SessionConfig" },
		"oauth": { "$ref": "#/$defs/OAuthConfig" },
		"identity": { "$ref": "#/$defs/IdentityConfig" },
		"authenticator": { "$ref": "#/$defs/AuthenticatorConfig" },
		"forgot_password": { "$ref": "#/$defs/ForgotPasswordConfig" },
		"welcome_message": { "$ref": "#/$defs/WelcomeMessageConfig" },
		"verification": { "$ref": "#/$defs/VerificationConfig" }
	},
	"required": ["id"]
}
`)

type AppConfig struct {
	ID AppID `json:"id"`

	HTTP     *HTTPConfig     `json:"http,omitempty"`
	Database *DatabaseConfig `json:"database,omitempty"`
	Redis    *RedisConfig    `json:"redis,omitempty"`
	Hook     *HookConfig     `json:"hook,omitempty"`

	UI           *UIConfig           `json:"ui,omitempty"`
	Localization *LocalizationConfig `json:"localization,omitempty"`
	Messaging    *MessagingConfig    `json:"messaging,omitempty"`

	Authentication *AuthenticationConfig `json:"authentication,omitempty"`
	Session        *SessionConfig        `json:"session,omitempty"`
	OAuth          *OAuthConfig          `json:"oauth,omitempty"`
	Identity       *IdentityConfig       `json:"identity,omitempty"`
	Authenticator  *AuthenticatorConfig  `json:"authenticator,omitempty"`

	ForgotPassword *ForgotPasswordConfig `json:"forgot_password,omitempty"`
	WelcomeMessage *WelcomeMessageConfig `json:"welcome_message,omitempty"`
	Verification   *VerificationConfig   `json:"verification,omitempty"`
}

func (c *AppConfig) Validate(ctx *validation.Context) {
	for i, client := range c.OAuth.Clients {
		if client.RefreshTokenLifetime() < client.AccessTokenLifetime() {
			ctx.Child("oauth", "clients", strconv.Itoa(i), "refresh_token_lifetime_seconds").EmitErrorMessage(
				"refresh token lifetime must be greater than or equal to access token lifetime",
			)
		}
	}

	oAuthProviderIDs := map[string]struct{}{}
	oauthProviderAliases := map[string]struct{}{}
	for i, provider := range c.Identity.OAuth.Providers {
		// Ensure provider ID is not duplicated
		id, err := json.Marshal(provider.ProviderID().Claims())
		if err != nil {
			panic("config: cannot marshal provider ID claims: " + err.Error())
		}
		if _, ok := oAuthProviderIDs[string(id)]; ok {
			ctx.Child("identity", "oauth", "providers", strconv.Itoa(i)).
				EmitErrorMessage("duplicated OAuth provider")
			continue
		}
		oAuthProviderIDs[string(id)] = struct{}{}

		// Ensure alias is not duplicated.
		if _, ok := oauthProviderAliases[provider.Alias]; ok {
			ctx.Child("identity", "oauth", "providers", strconv.Itoa(i)).
				EmitErrorMessage("duplicated OAuth provider alias")
			continue
		}
		oauthProviderAliases[provider.Alias] = struct{}{}
	}

	authenticatorTypes := map[authn.AuthenticatorType]struct{}{}
	for i, a := range c.Authentication.PrimaryAuthenticators {
		if _, ok := authenticatorTypes[a]; ok {
			ctx.Child("authentication", "primary_authenticators", strconv.Itoa(i)).
				EmitErrorMessage("duplicated authenticator type")
		}
		authenticatorTypes[a] = struct{}{}
	}
	for i, a := range c.Authentication.SecondaryAuthenticators {
		if _, ok := authenticatorTypes[a]; ok {
			ctx.Child("authentication", "secondary_authenticators", strconv.Itoa(i)).
				EmitErrorMessage("duplicated authenticator type")
		}
		authenticatorTypes[a] = struct{}{}
	}

	for i, it := range c.Authentication.Identities {
		hasPrimaryAuth := true
		switch it {
		case authn.IdentityTypeLoginID:
			_, hasPassword := authenticatorTypes[authn.AuthenticatorTypePassword]
			_, hasOOB := authenticatorTypes[authn.AuthenticatorTypeOOB]
			for _, k := range c.Identity.LoginID.Keys {
				switch k.Type {
				case LoginIDKeyTypeEmail, LoginIDKeyTypePhone:
					if !hasPassword && !hasOOB {
						hasPrimaryAuth = false
					}
				case LoginIDKeyTypeUsername:
					if !hasPassword {
						hasPrimaryAuth = false
					}
				}
			}
		case authn.IdentityTypeOAuth, authn.IdentityTypeAnonymous:
			// Primary authenticator is not needed for these types of identity.
			break
		}

		if !hasPrimaryAuth {
			ctx.Child("authentication", "identities", strconv.Itoa(i)).
				EmitErrorMessage("no usable primary authenticator is enabled")
		}
	}

	countryCallingCodePinnedOK := true
	countryCallingCodeAllowListMap := make(map[string]bool)
	for _, code := range c.UI.CountryCallingCode.AllowList {
		countryCallingCodeAllowListMap[code] = true
	}

	for _, pinnedCode := range c.UI.CountryCallingCode.PinnedList {
		if !countryCallingCodeAllowListMap[pinnedCode] {
			countryCallingCodePinnedOK = false
		}
	}

	if !countryCallingCodePinnedOK {
		ctx.Child("ui", "country_calling_code", "pinned_list").
			EmitErrorMessage("pinned country calling code is unlisted")
	}
}

func Parse(inputYAML []byte) (*AppConfig, error) {
	const validationErrorMessage = "invalid configuration"

	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}

	err = Schema.Validator().ValidateWithMessage(bytes.NewReader(jsonData), validationErrorMessage)
	if err != nil {
		return nil, err
	}

	var config AppConfig
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	PopulateDefaultValues(&config)

	err = validation.ValidateValueWithMessage(&config, validationErrorMessage)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func PopulateDefaultValues(config *AppConfig) {
	setFieldDefaults(config)
}
