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
	AppConfig     *AppConfig
	SecretConfig  *SecretConfig
	FeatureConfig *FeatureConfig
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
		"user_profile": { "$ref": "#/$defs/UserProfileConfig" },
		"forgot_password": { "$ref": "#/$defs/ForgotPasswordConfig" },
		"welcome_message": { "$ref": "#/$defs/WelcomeMessageConfig" },
		"verification": { "$ref": "#/$defs/VerificationConfig" }
	},
	"required": ["id", "http"]
}
`)

type AppConfig struct {
	ID AppID `json:"id"`

	HTTP     *HTTPConfig     `json:"http"`
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
	UserProfile    *UserProfileConfig    `json:"user_profile,omitempty"`

	ForgotPassword *ForgotPasswordConfig `json:"forgot_password,omitempty"`
	WelcomeMessage *WelcomeMessageConfig `json:"welcome_message,omitempty"`
	Verification   *VerificationConfig   `json:"verification,omitempty"`
}

func (c *AppConfig) Validate(ctx *validation.Context) {
	for i, client := range c.OAuth.Clients {
		if client.RefreshTokenLifetime < client.AccessTokenLifetime {
			ctx.Child("oauth", "clients", strconv.Itoa(i), "refresh_token_lifetime_seconds").EmitErrorMessage(
				"refresh token lifetime must be greater than or equal to access token lifetime",
			)
		}
	}

	oAuthProviderIDs := map[string]struct{}{}
	oauthProviderAliases := map[string]struct{}{}
	for i, provider := range c.Identity.OAuth.Providers {
		// Ensure provider ID is not duplicated
		// Except WeChat provider with different app type
		providerID := map[string]interface{}{}
		for k, v := range provider.ProviderID().Claims() {
			providerID[k] = v
		}
		if provider.Type == OAuthSSOProviderTypeWechat {
			providerID["app_type"] = provider.AppType
		}
		id, err := json.Marshal(providerID)
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
	for _, a := range c.Authentication.PrimaryAuthenticators {
		authenticatorTypes[a] = struct{}{}
	}

	for i, it := range c.Authentication.Identities {
		hasPrimaryAuth := true
		var loginIDKeyType LoginIDKeyType
		switch it {
		case authn.IdentityTypeLoginID:
			_, hasPassword := authenticatorTypes[authn.AuthenticatorTypePassword]
			_, hasOOBEmail := authenticatorTypes[authn.AuthenticatorTypeOOBEmail]
			_, hasOOBSMS := authenticatorTypes[authn.AuthenticatorTypeOOBSMS]
			for _, k := range c.Identity.LoginID.Keys {
				switch k.Type {
				case LoginIDKeyTypeEmail:
					if !hasPassword && !hasOOBEmail {
						hasPrimaryAuth = false
						loginIDKeyType = k.Type
					}
				case LoginIDKeyTypePhone:
					if !hasPassword && !hasOOBSMS {
						hasPrimaryAuth = false
						loginIDKeyType = k.Type
					}
				case LoginIDKeyTypeUsername:
					if !hasPassword {
						hasPrimaryAuth = false
						loginIDKeyType = k.Type
					}
				}
			}
		case authn.IdentityTypeOAuth, authn.IdentityTypeAnonymous:
			// Primary authenticator is not needed for these types of identity.
			break
		}

		if !hasPrimaryAuth {
			ctx.Child("authentication", "identities", strconv.Itoa(i)).
				EmitError(
					"noPrimaryAuthenticator",
					map[string]interface{}{"login_id_type": loginIDKeyType},
				)
		}
	}

	phoneInputPinnedOK := true
	phoneInputAllowListMap := make(map[string]bool)
	for _, alpha2 := range c.UI.PhoneInput.AllowList {
		phoneInputAllowListMap[alpha2] = true
	}

	for _, alpha2 := range c.UI.PhoneInput.PinnedList {
		if !phoneInputAllowListMap[alpha2] {
			phoneInputPinnedOK = false
		}
	}

	if !phoneInputPinnedOK {
		ctx.Child("ui", "phone_input", "pinned_list").
			EmitErrorMessage("pinned country code is unlisted")
	}

	supportedLanguagesSet := make(map[string]struct{})
	for _, tag := range c.Localization.SupportedLanguages {
		supportedLanguagesSet[tag] = struct{}{}
	}
	_, fallbackLanguageOK := supportedLanguagesSet[*c.Localization.FallbackLanguage]
	if !fallbackLanguageOK {
		ctx.Child("localization", "supported_languages").EmitErrorMessage("supported_languages must contain fallback_language")
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
