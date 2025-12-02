package config

import (
	"bytes"
	"context"
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
		"search": { "$ref": "#/$defs/SearchConfig" },
		"authentication": { "$ref": "#/$defs/AuthenticationConfig" },
		"session": { "$ref": "#/$defs/SessionConfig" },
		"oauth": { "$ref": "#/$defs/OAuthConfig" },
		"saml": { "$ref": "#/$defs/SAMLConfig" },
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
		"bot_protection": { "$ref": "#/$defs/BotProtectionConfig" },
		"protection": { "$ref": "#/$defs/ProtectionConfig" },
		"test_mode": { "$ref": "#/$defs/TestModeConfig" },
		"authentication_flow": { "$ref": "#/$defs/AuthenticationFlowConfig" },
		"external_jwt": { "$ref": "#/$defs/ExternalJWTConfig" }
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
	Search       *SearchConfig       `json:"search,omitempty"`

	Authentication       *AuthenticationConfig       `json:"authentication,omitempty"`
	Session              *SessionConfig              `json:"session,omitempty"`
	OAuth                *OAuthConfig                `json:"oauth,omitempty"`
	SAML                 *SAMLConfig                 `json:"saml,omitempty"`
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

	Deprecated_Web3 *Deprecated_Web3Config `json:"web3,omitempty"`

	GoogleTagManager *GoogleTagManagerConfig `json:"google_tag_manager,omitempty"`

	AccountMigration *AccountMigrationConfig `json:"account_migration,omitempty"`

	Captcha       *CaptchaConfig       `json:"captcha,omitempty"`
	BotProtection *BotProtectionConfig `json:"bot_protection,omitempty"`
	Protection    *ProtectionConfig    `json:"protection,omitempty"`

	TestMode *TestModeConfig `json:"test_mode,omitempty"`

	AuthenticationFlow *AuthenticationFlowConfig `json:"authentication_flow,omitempty"`

	ExternalJWT *ExternalJWTConfig `json:"external_jwt,omitempty"`
}

var _ validation.Validator = (*AppConfig)(nil)

func (c *AppConfig) SetDefaults() {
	c.Deprecated_WelcomeMessage = nil
	c.Deprecated_OTP = nil
}

func (c *AppConfig) Validate(ctx context.Context, validationCtx *validation.Context) {
	// Validation 1: lifetime of refresh token >= lifetime of access token
	c.validateTokenLifetime(validationCtx)

	// Validation 2: oauth provider
	c.validateOAuthProvider(ctx, validationCtx)

	// Validation 3: identity must have usable primary authenticator.
	c.validateIdentityPrimaryAuthenticator(validationCtx)

	// Validation 4: secondary authenticator must be available if MFA is not disabled.
	c.validateSecondaryAuthenticator(validationCtx)

	// Validation 5: pinned phone number country must be in allowlist.
	c.validatePhoneInputCountry(validationCtx)

	// Validation 6: fallback language must be in the list of supported language.
	c.validateFallbackLanguage(validationCtx)

	// Validation 7: validate custom attribute
	c.validateCustomAttribute(validationCtx)

	// Validation 8: validate lockout configs
	c.validateLockout(validationCtx)

	// Validation 9: validate authentication flow
	c.validateAuthenticationFlow(validationCtx)

	// Validation 10: validate saml configs
	c.validateSAML(validationCtx)
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

func (c *AppConfig) validateOAuthProvider(ctx context.Context, validationCtx *validation.Context) {
	// We used to validate that ProviderID is unique.
	// We now relax the validation, only alias is unique.
	oauthProviderAliases := map[string]struct{}{}
	for i, providerConfig := range c.Identity.OAuth.Providers {
		// We used to ensure provider ID is not duplicated.
		// We now expect alias to be unique.
		alias := providerConfig.Alias()
		childCtx := validationCtx.Child("identity", "oauth", "providers", strconv.Itoa(i))

		if _, ok := oauthProviderAliases[alias]; ok {
			childCtx.EmitErrorMessage("duplicated OAuth provider alias")
			continue
		}
		oauthProviderAliases[alias] = struct{}{}

		// Validate provider config
		provider := providerConfig.AsProviderConfig().MustGetProvider()
		schema := OAuthSSOProviderConfigSchemaBuilder(validation.SchemaBuilder(provider.GetJSONSchema())).ToSimpleSchema()
		childCtx.AddError(schema.Validator().ValidateValue(ctx, providerConfig))
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
			// NOTE(DEV-3131): We removed the check for login ID, because in custom authflow
			// it is possible to use username to identify, and use oobotp to authenticate.
			// Therefore we decided to allow enabling login ID independently.
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

func (c *AppConfig) validateSAML(ctx *validation.Context) {
	if len(c.SAML.ServiceProviders) > 0 {
		if c.SAML.Signing.KeyID == "" {
			// Signing key must be configured if at least one service provider exist
			ctx.Child("saml", "signing", "key_id").EmitError("minLength", map[string]interface{}{
				"expected": 1,
				"actual":   0,
			})
		}

		for idx, sp := range c.SAML.ServiceProviders {
			c.validateSAMLServiceProvider(ctx, idx, sp)
		}
	}
}

func (c *AppConfig) validateSAMLServiceProvider(ctx *validation.Context, idx int, sp *SAMLServiceProviderConfig) {
	if sp.ClientID != "" {
		found := false
		for _, oauthClient := range c.OAuth.Clients {
			if sp.ClientID == oauthClient.ClientID {
				found = true
				break
			}
		}
		if !found {
			ctx.Child("saml", "service_providers", strconv.Itoa(idx), "client_id").
				EmitErrorMessage("client_id does not exist in /oauth/clients")
		}
	}

	for mappingIdx, mapping := range sp.Attributes.Mappings {
		found := false
		for _, definition := range sp.Attributes.Definitions {
			if definition.Name == mapping.To.SAMLAttribute {
				found = true
				break
			}
		}
		if !found {
			ctx.Child(
				"saml",
				"service_providers",
				strconv.Itoa(idx),
				"mappings",
				strconv.Itoa(mappingIdx),
				"to",
				"saml_attribute",
			).
				EmitErrorMessage("saml_attribute does match any defined attribute name in definitions")
		}
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

func Parse(ctx context.Context, inputYAML []byte) (*AppConfig, error) {
	const validationErrorMessage = "invalid configuration"

	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}

	err = Schema.Validator().ValidateWithMessage(ctx, bytes.NewReader(jsonData), validationErrorMessage)
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

	err = validation.ValidateValueWithMessage(ctx, &config, validationErrorMessage)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func PopulateDefaultValues(config *AppConfig) {
	SetFieldDefaults(config)
}
