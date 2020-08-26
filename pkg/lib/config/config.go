package config

import (
	"bytes"
	"encoding/json"
	"strconv"

	"sigs.k8s.io/yaml"

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
		"metadata": { "$ref": "#/$defs/AppMetadata" },
		"http": { "$ref": "#/$defs/HTTPConfig" },
		"database": { "$ref": "#/$defs/DatabaseConfig" },
		"redis": { "$ref": "#/$defs/RedisConfig" },
		"hook": { "$ref": "#/$defs/HookConfig" },
		"template": { "$ref": "#/$defs/TemplateConfig" },
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
	ID       AppID       `json:"id"`
	Metadata AppMetadata `json:"metadata,omitempty"`

	HTTP     *HTTPConfig     `json:"http,omitempty"`
	Database *DatabaseConfig `json:"database,omitempty"`
	Redis    *RedisConfig    `json:"redis,omitempty"`
	Hook     *HookConfig     `json:"hook,omitempty"`

	Template     *TemplateConfig     `json:"template,omitempty"`
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

	authenticatorTypes := map[string]struct{}{}
	for i, a := range c.Authentication.PrimaryAuthenticators {
		if _, ok := authenticatorTypes[string(a)]; ok {
			ctx.Child("authentication", "primary_authenticators", strconv.Itoa(i)).
				EmitErrorMessage("duplicated authenticator type")
		}
		authenticatorTypes[string(a)] = struct{}{}
	}
	for i, a := range c.Authentication.SecondaryAuthenticators {
		if _, ok := authenticatorTypes[string(a)]; ok {
			ctx.Child("authentication", "secondary_authenticators", strconv.Itoa(i)).
				EmitErrorMessage("duplicated authenticator type")
		}
		authenticatorTypes[string(a)] = struct{}{}
	}

	countryCallingCodeDefaultOK := false
	for _, code := range c.UI.CountryCallingCode.Values {
		if code == c.UI.CountryCallingCode.Default {
			countryCallingCodeDefaultOK = true
		}
	}
	if !countryCallingCodeDefaultOK {
		ctx.Child("ui", "country_calling_code", "default").
			EmitErrorMessage("default country calling code is unlisted")
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

	setFieldDefaults(&config)

	err = validation.ValidateValueWithMessage(&config, validationErrorMessage)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
