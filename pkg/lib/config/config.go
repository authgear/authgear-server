package config

import (
	"bytes"
	"encoding/json"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/api/model"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
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
		"account_anonymization": { "$ref": "#/$defs/AccountAnonymizationConfig" },
		"account_linking": { "$ref": "#/$defs/AccountLinkingConfig" },
		"forgot_password": { "$ref": "#/$defs/ForgotPasswordConfig" },
		"welcome_message": { "$ref": "#/$defs/WelcomeMessageConfig" },
		"verification": { "$ref": "#/$defs/VerificationConfig" },
		"otp": { "$ref": "#/$defs/OTPLegacyConfig" },
		"web3": { "$ref": "#/$defs/Web3Config" },
		"google_tag_manager": { "$ref": "#/$defs/GoogleTagManagerConfig" },
		"account_migration": { "$ref": "#/$defs/AccountMigrationConfig" },
		"captcha": { "$ref": "#/$defs/CaptchaConfig" },
		"test_mode": { "$ref": "#/$defs/TestModeConfig" },
		"authentication_flow": { "$ref": "#/$defs/AuthenticationFlowConfig" }
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

	Authentication       *AuthenticationConfig       `json:"authentication,omitempty"`
	Session              *SessionConfig              `json:"session,omitempty"`
	OAuth                *OAuthConfig                `json:"oauth,omitempty"`
	Identity             *IdentityConfig             `json:"identity,omitempty"`
	Authenticator        *AuthenticatorConfig        `json:"authenticator,omitempty"`
	UserProfile          *UserProfileConfig          `json:"user_profile,omitempty"`
	AccountDeletion      *AccountDeletionConfig      `json:"account_deletion,omitempty"`
	AccountAnonymization *AccountAnonymizationConfig `json:"account_anonymization,omitempty"`
	AccountLinking       *AccountLinkingConfig       `json:"account_linking,omitempty"`

	ForgotPassword            *ForgotPasswordConfig `json:"forgot_password,omitempty"`
	Deprecated_WelcomeMessage *WelcomeMessageConfig `json:"welcome_message,omitempty"`
	Verification              *VerificationConfig   `json:"verification,omitempty"`
	Deprecated_OTP            *OTPLegacyConfig      `json:"otp,omitempty"`

	Web3 *Web3Config `json:"web3,omitempty"`

	GoogleTagManager *GoogleTagManagerConfig `json:"google_tag_manager,omitempty"`

	AccountMigration *AccountMigrationConfig `json:"account_migration,omitempty"`

	Captcha *CaptchaConfig `json:"captcha,omitempty"`

	TestMode *TestModeConfig `json:"test_mode,omitempty"`

	AuthenticationFlow *AuthenticationFlowConfig `json:"authentication_flow,omitempty"`
}

func (c *AppConfig) SetDefaults() {
	c.Deprecated_WelcomeMessage = nil
	c.Deprecated_OTP = nil
}

func (c *AppConfig) Validate(ctx *validation.Context) {
	// Validation 1: lifetime of refresh token >= lifetime of access token
	c.validateTokenLifetime(ctx)

	// Validation 2: oauth provider
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

	// Validation 8: validate lockout configs
	c.validateLockout(ctx)

	// Validation 9: validate authentication flow
	c.validateAuthenticationFlow(ctx)
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
	// We used to validate that ProviderID is unique.
	// We now relax the validation, only alias is unique.
	oauthProviderAliases := map[string]struct{}{}
	for i, providerConfig := range c.Identity.OAuth.Providers {
		// We used to ensure provider ID is not duplicated.
		// We now expect alias to be unique.
		alias := providerConfig.Alias()
		childCtx := ctx.Child("identity", "oauth", "providers", strconv.Itoa(i))

		if _, ok := oauthProviderAliases[alias]; ok {
			childCtx.EmitErrorMessage("duplicated OAuth provider alias")
			continue
		}
		oauthProviderAliases[alias] = struct{}{}

		// Validate provider config if it is a builin provider.
		// TODO(tung): ValidateProviderConfig should be defined in external package?
		provider := providerConfig.AsProviderConfig().MustGetProvider()
		if builtinProvider, ok := provider.(liboauthrelyingparty.BuiltinProvider); ok {
			builtinProvider.ValidateProviderConfig(childCtx, providerConfig.AsProviderConfig())
		}
	}
}

func (c *AppConfig) validateIdentityPrimaryAuthenticatorLoginID(
	ctx *validation.Context,
	authenticatorTypes map[model.AuthenticatorType]struct{},
	k LoginIDKeyConfig,
	idx int) {
	hasAtLeastOnePrimaryAuthenticator := false
	required := model.IdentityTypeLoginID.PrimaryAuthenticatorTypes(k.Type)
	for _, typ := range required {
		if _, ok := authenticatorTypes[typ]; ok {
			hasAtLeastOnePrimaryAuthenticator = true
		}
	}
	if len(required) > 0 && !hasAtLeastOnePrimaryAuthenticator {
		ctx.Child("authentication", "identities", strconv.Itoa(idx)).
			EmitError(
				"noPrimaryAuthenticator",
				map[string]interface{}{
					"identity_type": model.IdentityTypeLoginID,
					"login_id_type": k.Type,
				},
			)
	}
}

func (c *AppConfig) validateIdentityPrimaryAuthenticatorOthers(
	ctx *validation.Context,
	authenticatorTypes map[model.AuthenticatorType]struct{},
	it model.IdentityType,
	idx int) {
	hasAtLeastOnePrimaryAuthenticator := false
	required := it.PrimaryAuthenticatorTypes("")
	for _, typ := range required {
		if _, ok := authenticatorTypes[typ]; ok {
			hasAtLeastOnePrimaryAuthenticator = true
		}
	}
	if len(required) > 0 && !hasAtLeastOnePrimaryAuthenticator {
		ctx.Child("authentication", "identities", strconv.Itoa(idx)).
			EmitError(
				"noPrimaryAuthenticator",
				map[string]interface{}{
					"identity_type": it,
				},
			)
	}
}

func (c *AppConfig) validateIdentityPrimaryAuthenticator(ctx *validation.Context) {
	authenticatorTypes := map[model.AuthenticatorType]struct{}{}
	for _, a := range *c.Authentication.PrimaryAuthenticators {
		authenticatorTypes[a] = struct{}{}
	}

	for idx, it := range c.Authentication.Identities {
		if it == model.IdentityTypeLoginID {
			for _, k := range c.Identity.LoginID.Keys {
				c.validateIdentityPrimaryAuthenticatorLoginID(ctx, authenticatorTypes, k, idx)
			}
		} else {
			c.validateIdentityPrimaryAuthenticatorOthers(ctx, authenticatorTypes, it, idx)
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

func (c *AppConfig) validateLockout(ctx *validation.Context) {
	minDuration, isMinDurationValid := c.Authentication.Lockout.MinimumDuration.MaybeDuration()
	maxDuration, isMaxDurationValid := c.Authentication.Lockout.MaximumDuration.MaybeDuration()
	if isMaxDurationValid && isMinDurationValid && minDuration.Seconds() > maxDuration.Seconds() {
		ctx.Child("authentication", "lockout", "minimum_duration").EmitError("maximum", map[string]interface{}{
			"maximum": maxDuration.Seconds(),
			"actual":  minDuration.Seconds(),
		})
	}
}

func (c *AppConfig) validateAuthenticationFlow(ctx *validation.Context) {
	if c.UI.AuthenticationFlow == nil {
		return
	}

	definedFlows := constructDefinedFlows(c.AuthenticationFlow)
	definedGroups := constructDefinedGroups(c.UI.AuthenticationFlow)

	// Ensure defined groups are valid and unique
	validateDefinedGroups(ctx, c.UI.AuthenticationFlow, definedFlows)

	for i, client := range c.OAuth.Clients {
		// Ensure client's group allowlist is valid
		validateGroupAllowlist(ctx, client.AuthenticationFlowAllowlist.Groups, definedGroups, i)

		// Ensure client's flow allowlist is valid
		validateFlowAllowlist(ctx, client.AuthenticationFlowAllowlist.Flows, definedFlows, i)
	}
}

func constructDefinedFlows(flowConfig *AuthenticationFlowConfig) []*AuthenticationFlowAllowlistFlow {
	definedlist := []*AuthenticationFlowAllowlistFlow{}
	definedlist = append(definedlist, flowsToAllowlist(flowConfig.SignupFlows, AuthenticationFlowTypeSignup)...)
	definedlist = append(definedlist, flowsToAllowlist(flowConfig.LoginFlows, AuthenticationFlowTypeLogin)...)
	definedlist = append(definedlist, flowsToAllowlist(flowConfig.PromoteFlows, AuthenticationFlowTypePromote)...)
	definedlist = append(definedlist, flowsToAllowlist(flowConfig.SignupLoginFlows, AuthenticationFlowTypeSignupLogin)...)
	definedlist = append(definedlist, flowsToAllowlist(flowConfig.ReauthFlows, AuthenticationFlowTypeReauth)...)
	definedlist = append(definedlist, flowsToAllowlist(flowConfig.AccountRecoveryFlows, AuthenticationFlowTypeAccountRecovery)...)
	return definedlist
}

func constructDefinedGroups(groupConfig *UIAuthenticationFlowConfig) []string {
	definedlist := []string{}
	if groupConfig.Groups != nil {
		for _, group := range groupConfig.Groups {
			definedlist = append(definedlist, group.Name)
		}
	}
	return definedlist
}

func flowsToAllowlist[TA AuthenticationFlowObjectFlowRoot](definedFlows []TA, flowType AuthenticationFlowType) []*AuthenticationFlowAllowlistFlow {
	allowlist := []*AuthenticationFlowAllowlistFlow{}
	for _, flow := range definedFlows {
		allowlist = append(allowlist, &AuthenticationFlowAllowlistFlow{
			Type: flowType,
			Name: flow.GetName(),
		})
	}
	return allowlist
}

func validateDefinedGroups(ctx *validation.Context, config *UIAuthenticationFlowConfig, definedFlows []*AuthenticationFlowAllowlistFlow) {
	definedGroups := map[string]struct{}{}
	for i, group := range config.Groups {
		// Ensure defined groups are unique
		if _, ok := definedGroups[group.Name]; ok {
			ctx.Child("ui", "authentication_flow", "groups", strconv.Itoa(i)).EmitErrorMessage("duplicated authentication flow group")
			continue
		}
		definedGroups[group.Name] = struct{}{}

		hasLoginFlow := false
		for j, flow := range group.Flows {
			if flow.Type == AuthenticationFlowTypeLogin {
				hasLoginFlow = true
			}

			// Ensure allowed flows are defined
			flowIsDefined := false
			if flow.Name == "default" {
				flowIsDefined = true
			}
			for _, definedFlow := range definedFlows {
				if flow.Type == definedFlow.Type && flow.Name == definedFlow.Name {
					flowIsDefined = true
					break
				}
			}
			if !flowIsDefined {
				ctx.Child("ui", "authentication_flow", "groups", strconv.Itoa(i), "flows", strconv.Itoa(j)).EmitErrorMessage("invalid authentication flow")
			}
		}
		// Require at least one login flow
		if !hasLoginFlow {
			ctx.Child("ui", "authentication_flow", "groups", strconv.Itoa(i)).EmitErrorMessage("authentication flow group must contain one login flow")
		}
	}
}

func validateGroupAllowlist(ctx *validation.Context, allowlist []*AuthenticationFlowAllowlistGroup, definedlist []string, idx int) {
	for i, group := range allowlist {
		groupIsDefined := false
		if group.Name == "default" {
			groupIsDefined = true
		}

		for _, definedGroup := range definedlist {
			if group.Name == definedGroup {
				groupIsDefined = true
				break
			}
		}

		if !groupIsDefined {
			ctx.Child("oauth", "clients", strconv.Itoa(idx), "authentication_flow_allowlist", "groups", strconv.Itoa(i)).EmitErrorMessage("invalid authentication flow group")
		}
	}
}

func validateFlowAllowlist(ctx *validation.Context, allowlist []*AuthenticationFlowAllowlistFlow, definedlist []*AuthenticationFlowAllowlistFlow, idx int) {
	for i, flow := range allowlist {
		flowIsDefined := false
		if flow.Name == "default" {
			flowIsDefined = true
		}

		for _, definedFlow := range definedlist {
			if flow.Type == definedFlow.Type && flow.Name == definedFlow.Name {
				flowIsDefined = true
				break
			}
		}
		if !flowIsDefined {
			ctx.Child("oauth", "clients", strconv.Itoa(idx), "authentication_flow_allowlist", "flows", strconv.Itoa(i)).EmitErrorMessage("invalid authentication flow")
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
	SetFieldDefaults(config)
}
