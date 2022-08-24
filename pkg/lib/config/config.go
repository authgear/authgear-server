package config

import (
	"bytes"
	"encoding/json"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/model"
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
		"account_deletion": { "$ref": "#/$defs/AccountDeletionConfig" },
		"forgot_password": { "$ref": "#/$defs/ForgotPasswordConfig" },
		"welcome_message": { "$ref": "#/$defs/WelcomeMessageConfig" },
		"verification": { "$ref": "#/$defs/VerificationConfig" },
		"web3": { "$ref": "#/$defs/Web3Config" },
		"google_tag_manager": { "$ref": "#/$defs/GoogleTagManagerConfig" }
	},
	"required": ["id", "http"]
}
`)

type AppConfig struct {
	ID AppID `json:"id"`

	HTTP *HTTPConfig `json:"http"`
	Hook *HookConfig `json:"hook,omitempty"`

	UI           *UIConfig           `json:"ui,omitempty"`
	Localization *LocalizationConfig `json:"localization,omitempty"`
	Messaging    *MessagingConfig    `json:"messaging,omitempty"`

	Authentication  *AuthenticationConfig  `json:"authentication,omitempty"`
	Session         *SessionConfig         `json:"session,omitempty"`
	OAuth           *OAuthConfig           `json:"oauth,omitempty"`
	Identity        *IdentityConfig        `json:"identity,omitempty"`
	Authenticator   *AuthenticatorConfig   `json:"authenticator,omitempty"`
	UserProfile     *UserProfileConfig     `json:"user_profile,omitempty"`
	AccountDeletion *AccountDeletionConfig `json:"account_deletion,omitempty"`

	ForgotPassword *ForgotPasswordConfig `json:"forgot_password,omitempty"`
	WelcomeMessage *WelcomeMessageConfig `json:"welcome_message,omitempty"`
	Verification   *VerificationConfig   `json:"verification,omitempty"`

	Web3 *Web3Config `json:"web3,omitempty"`

	GoogleTagManager *GoogleTagManagerConfig `json:"google_tag_manager,omitempty"`
}

func (c *AppConfig) Validate(ctx *validation.Context) {
	// Validation 1: lifetime of refresh token >= lifetime of access token
	c.validateTokenLifetime(ctx)

	// Validation 2: OAuth provider cannot duplicate.
	c.validateOAuthProvider(ctx)

	// Validation 3: identity must have usable primary authenticator.
	c.validateIdentityPrimaryAuthenticator(ctx)

	// Validation 4: secondary authenticator must be available if MFA is not disabled.
	c.validateSecondaryAuthenticator(ctx)

	// Validation 5: pinned phone number country must be in allowlist.
	c.validatePhoneInputCountry(ctx)

	// Validation 6: fallback language must be in the list of supported language.
	c.validateFallbackLanguage(ctx)

	// Validation 7: validate custom attribute
	c.validateCustomAttribute(ctx)
}

func (c *AppConfig) validateTokenLifetime(ctx *validation.Context) {
	for i, client := range c.OAuth.Clients {
		if client.RefreshTokenLifetime < client.AccessTokenLifetime {
			ctx.Child("oauth", "clients", strconv.Itoa(i), "refresh_token_lifetime_seconds").EmitErrorMessage(
				"refresh token lifetime must be greater than or equal to access token lifetime",
			)
		}
	}
}

func (c *AppConfig) validateOAuthProvider(ctx *validation.Context) {
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
}

func (c *AppConfig) validateIdentityPrimaryAuthenticator(ctx *validation.Context) {
	authenticatorTypes := map[model.AuthenticatorType]struct{}{}
	for _, a := range *c.Authentication.PrimaryAuthenticators {
		authenticatorTypes[a] = struct{}{}
	}

	for i, it := range c.Authentication.Identities {
		if it == model.IdentityTypeLoginID {
			for _, k := range c.Identity.LoginID.Keys {
				hasAtLeastOnePrimaryAuthenticator := false
				required := it.PrimaryAuthenticatorTypes(k.Type)
				for _, typ := range required {
					if _, ok := authenticatorTypes[typ]; ok {
						hasAtLeastOnePrimaryAuthenticator = true
					}
				}
				if len(required) > 0 && !hasAtLeastOnePrimaryAuthenticator {
					ctx.Child("authentication", "identities", strconv.Itoa(i)).
						EmitError(
							"noPrimaryAuthenticator",
							map[string]interface{}{
								"identity_type": it,
								"login_id_type": k.Type,
							},
						)
				}
			}
		} else {
			hasAtLeastOnePrimaryAuthenticator := false
			required := it.PrimaryAuthenticatorTypes("")
			for _, typ := range required {
				if _, ok := authenticatorTypes[typ]; ok {
					hasAtLeastOnePrimaryAuthenticator = true
				}
			}
			if len(required) > 0 && !hasAtLeastOnePrimaryAuthenticator {
				ctx.Child("authentication", "identities", strconv.Itoa(i)).
					EmitError(
						"noPrimaryAuthenticator",
						map[string]interface{}{
							"identity_type": it,
						},
					)
			}
		}
	}
}

func (c *AppConfig) validateSecondaryAuthenticator(ctx *validation.Context) {
	if !c.Authentication.SecondaryAuthenticationMode.IsDisabled() {
		if len(*c.Authentication.SecondaryAuthenticators) <= 0 {
			ctx.Child("authentication", "secondary_authentication_mode").
				EmitError(
					"noSecondaryAuthenticator",
					map[string]interface{}{"secondary_authentication_mode": c.Authentication.SecondaryAuthenticationMode})
		}
	}
}

func (c *AppConfig) validatePhoneInputCountry(ctx *validation.Context) {
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
}

func (c *AppConfig) validateFallbackLanguage(ctx *validation.Context) {
	supportedLanguagesSet := make(map[string]struct{})
	for _, tag := range c.Localization.SupportedLanguages {
		supportedLanguagesSet[tag] = struct{}{}
	}
	_, fallbackLanguageOK := supportedLanguagesSet[*c.Localization.FallbackLanguage]
	if !fallbackLanguageOK {
		ctx.Child("localization", "supported_languages").EmitErrorMessage("supported_languages must contain fallback_language")
	}
}

func (c *AppConfig) validateCustomAttribute(ctx *validation.Context) {
	customAttributeIDs := map[string]struct{}{}
	customAttributePointers := map[string]struct{}{}
	for i, customAttributeConfig := range c.UserProfile.CustomAttributes.Attributes {
		if _, ok := customAttributeIDs[customAttributeConfig.ID]; ok {
			ctx.Child(
				"user_profile",
				"custom_attributes",
				"attributes",
				strconv.Itoa(i),
				"id",
			).EmitError("duplicated", nil)
		}
		if _, ok := customAttributePointers[customAttributeConfig.Pointer]; ok {
			ctx.Child(
				"user_profile",
				"custom_attributes",
				"attributes",
				strconv.Itoa(i),
				"pointer",
			).EmitError("duplicated", nil)
		}
		customAttributeIDs[customAttributeConfig.ID] = struct{}{}
		customAttributePointers[customAttributeConfig.Pointer] = struct{}{}

		// ensure the minimum config is smaller than the maximum config
		if customAttributeConfig.Type == CustomAttributeTypeNumber ||
			customAttributeConfig.Type == CustomAttributeTypeInteger {
			if customAttributeConfig.Maximum != nil &&
				customAttributeConfig.Minimum != nil &&
				*customAttributeConfig.Minimum > *customAttributeConfig.Maximum {
				ctx.Child(
					"user_profile",
					"custom_attributes",
					"attributes",
					strconv.Itoa(i),
					"minimum",
				).EmitError("maximum", map[string]interface{}{
					"maximum": *customAttributeConfig.Maximum,
					"actual":  *customAttributeConfig.Minimum,
				})
			}
		}
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
