package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var _ = SecretConfigSchema.Add("SecretConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"secrets": {
			"type": "array",
			"items": { "$ref": "#/$defs/SecretItem" }
		}
	},
	"required": ["secrets"]
}
`)

type SecretConfig struct {
	Secrets []SecretItem `json:"secrets,omitempty"`
}

const validationErrorMessage = "invalid secrets"

// ParsePartialSecret unmarshals inputYAML into a full SecretConfig,
// without performing validation.
func ParsePartialSecret(ctx context.Context, inputYAML []byte) (*SecretConfig, error) {

	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}

	var config SecretConfig
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return ParseSecretData(ctx, config)
}

func ParseSecret(ctx context.Context, inputYAML []byte) (*SecretConfig, error) {

	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}

	err = SecretConfigSchema.Validator().ValidateWithMessage(
		ctx,
		bytes.NewReader(jsonData),
		validationErrorMessage,
	)
	if err != nil {
		return nil, err
	}

	var config SecretConfig
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return ParseSecretData(ctx, config)
}

func ParseSecretData(ctx context.Context, config SecretConfig) (*SecretConfig, error) {
	vctx := &validation.Context{}
	for i := range config.Secrets {
		config.Secrets[i].parse(ctx, vctx.Child("secrets", strconv.Itoa(i)))
	}
	if err := vctx.Error(validationErrorMessage); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *SecretConfig) Overlay(layers ...*SecretConfig) *SecretConfig {
	items := make(map[SecretKey]SecretItem)
	for _, item := range c.Secrets {
		items[item.Key] = item
	}
	for _, layer := range layers {
		for _, item := range layer.Secrets {
			items[item.Key] = item
		}
	}

	merged := &SecretConfig{}
	for _, item := range items {
		merged.Secrets = append(merged.Secrets, item)
	}
	sort.Slice(merged.Secrets, func(i, j int) bool {
		return merged.Secrets[i].Key < merged.Secrets[j].Key
	})

	return merged
}

func (c *SecretConfig) Lookup(key SecretKey) (int, *SecretItem, bool) {
	for index, item := range c.Secrets {
		if item.Key == key {
			return index, &item, true
		}
	}
	return -1, nil, false
}

func (c *SecretConfig) LookupData(key SecretKey) SecretItemData {
	if _, item, ok := c.Lookup(key); ok {
		return item.Data
	}
	return nil
}

func (c *SecretConfig) LookupDataWithIndex(key SecretKey) (int, SecretItemData, bool) {
	if index, item, ok := c.Lookup(key); ok {
		return index, item.Data, true
	}
	return -1, nil, false
}

func (c *SecretConfig) validateRequire(ctx *validation.Context, key SecretKey, item string) {
	if _, _, ok := c.Lookup(key); !ok {
		ctx.EmitErrorMessage(fmt.Sprintf("%s (secret '%s') is required", item, key))
	}
}

func (c *SecretConfig) validateOAuthProviders(ctx *validation.Context, appConfig *AppConfig) {
	c.validateRequire(ctx, OAuthSSOProviderCredentialsKey, "OAuth SSO provider client credentials")
	secretIndex, data, _ := c.LookupDataWithIndex(OAuthSSOProviderCredentialsKey)
	oauth, ok := data.(*OAuthSSOProviderCredentials)
	if ok {
		for _, p := range appConfig.Identity.OAuth.Providers {
			providerAlias := p.Alias()
			var matchedItem *OAuthSSOProviderCredentialsItem = nil
			var matchedItemIndex int = -1
			for index := range oauth.Items {
				item := oauth.Items[index]
				if providerAlias == item.Alias {
					matchedItem = &item
					matchedItemIndex = index
					break
				}
			}
			if matchedItem == nil {
				ctx.EmitErrorMessage(fmt.Sprintf("OAuth SSO provider client credentials for '%s' is required", providerAlias))
			} else {
				if matchedItem.ClientSecret == "" && p.GetCredentialsBehavior() == OAuthSSOProviderCredentialsBehaviorUseProjectCredentials {
					ctx.Child("secrets", fmt.Sprintf("%d", secretIndex), "data", "items", fmt.Sprintf("%d", matchedItemIndex)).EmitError(
						"required",
						map[string]interface{}{
							"expected": []string{"alias", "client_secret"},
							"actual":   []string{"alias"},
							"missing":  []string{"client_secret"},
						},
					)
				}
			}
		}
	}
}

func (c *SecretConfig) validateConfidentialClients(ctx *validation.Context, confidentialClients []OAuthClientConfig) {
	c.validateRequire(ctx, OAuthClientCredentialsKey, "OAuth client credentials")
	_, data, _ := c.LookupDataWithIndex(OAuthClientCredentialsKey)
	oauth, ok := data.(*OAuthClientCredentials)
	if ok {
		for _, c := range confidentialClients {
			matched := false
			for index := range oauth.Items {
				item := oauth.Items[index]
				if c.ClientID == item.ClientID {
					matched = true
					break
				}
			}
			if !matched {
				ctx.EmitErrorMessage(fmt.Sprintf("OAuth client credentials for '%s' is required", c.ClientID))
			} else {
				// keys are validated by the jsonschema
			}
		}
	}
}

func (c *SecretConfig) validateBotProtectionSecrets(ctx *validation.Context, botProtectionProvider *BotProtectionProvider) {
	c.validateRequire(ctx, BotProtectionProviderCredentialsKey, "bot protection provider credentials")
	_, data, _ := c.LookupDataWithIndex(BotProtectionProviderCredentialsKey)
	botProtectionSecret, ok := data.(*BotProtectionProviderCredentials)
	if ok {
		matched := botProtectionSecret.Type == botProtectionProvider.Type
		if !matched {
			ctx.EmitErrorMessage(fmt.Sprintf("bot protection provider credentials for '%s' is required", botProtectionProvider.Type))
		} else {
			// keys are validated by the jsonschema
		}
	}
}

func (c *SecretConfig) validateLDAPServerUserSecrets(ctx *validation.Context, ldapServerConfig []*LDAPServerConfig) {
	c.validateRequire(ctx, LDAPServerUserCredentialsKey, "LDAP server user credentials")
	_, data, _ := c.LookupDataWithIndex(LDAPServerUserCredentialsKey)
	ldapServerUserCredentials, ok := data.(*LDAPServerUserCredentials)
	if ok {
		for _, server := range ldapServerConfig {
			matched := false
			for _, item := range ldapServerUserCredentials.Items {
				if server.Name == item.Name {
					matched = true
					break
				}
			}
			if !matched {
				ctx.EmitErrorMessage(fmt.Sprintf("LDAP server user credentials for '%s' is required", server.Name))
			} else {
				// keys are validated by the jsonschema
			}
		}
	}
}

func (c *SecretConfig) validateSAMLSigningKey(ctx *validation.Context, keyID string) {
	c.validateRequire(ctx, SAMLIdpSigningMaterialsKey, "saml idp signing key materials")
	_, data, ok := c.LookupDataWithIndex(SAMLIdpSigningMaterialsKey)
	if ok {
		signingMaterials, _ := data.(*SAMLIdpSigningMaterials)

		for _, m := range signingMaterials.Certificates {
			if m.Key.Key.KeyID() == keyID {
				return
			}
		}
	}
	ctx.EmitErrorMessage(fmt.Sprintf("saml idp signing key '%s' does not exist", keyID))
}

func (c *SecretConfig) validateSAMLServiceProviderCerts(ctx *validation.Context, sp *SAMLServiceProviderConfig) {
	_, data, _ := c.LookupDataWithIndex(SAMLSpSigningMaterialsKey)
	signingMaterials, _ := data.(*SAMLSpSigningMaterials)
	certs, _, ok := signingMaterials.Resolve(sp.GetID())
	if !ok || len(certs.Certificates) < 1 {
		ctx.EmitErrorMessage(fmt.Sprintf("certificates of saml sp '%s' is not configured", sp.GetID()))
	}
}

func (c *SecretConfig) validateSSOOAuthDemoCredentials(ctx context.Context, vctx *validation.Context, demoCredentials *SSOOAuthDemoCredentials) {
	for i, item := range demoCredentials.Items {
		providerConfig := item.ProviderConfig
		provider := providerConfig.MustGetProvider()
		schema := validation.SchemaBuilder(provider.GetJSONSchema()).ToSimpleSchema()
		itemCtx := vctx.Child("items", strconv.Itoa(i), "provider_config")
		itemCtx.AddError(schema.Validator().ValidateValue(ctx, providerConfig))
	}
}

func (c *SecretConfig) Validate(ctx context.Context, appConfig *AppConfig) error {
	vctx := &validation.Context{}

	c.validateRequire(vctx, DatabaseCredentialsKey, "database credentials")
	// AuditDatabaseCredentialsKey is not required
	// ElasticsearchCredentialsKey is not required
	c.validateRequire(vctx, RedisCredentialsKey, "redis credentials")
	c.validateRequire(vctx, AdminAPIAuthKeyKey, "admin API auth key materials")

	if len(appConfig.Identity.OAuth.Providers) > 0 {
		c.validateOAuthProviders(vctx, appConfig)
	}

	confidentialClients := []OAuthClientConfig{}
	for _, c := range appConfig.OAuth.Clients {
		if c.IsConfidential() {
			confidentialClients = append(confidentialClients, c)
		}
	}

	if len(confidentialClients) > 0 {
		c.validateConfidentialClients(vctx, confidentialClients)
	}

	c.validateRequire(vctx, OAuthKeyMaterialsKey, "OAuth key materials")
	c.validateRequire(vctx, CSRFKeyMaterialsKey, "CSRF key materials")
	if len(appConfig.Hook.BlockingHandlers) > 0 || len(appConfig.Hook.NonBlockingHandlers) > 0 {
		c.validateRequire(vctx, WebhookKeyMaterialsKey, "web-hook signing key materials")
	}
	if appConfig.BotProtection != nil && appConfig.BotProtection.Enabled && appConfig.BotProtection.Provider != nil {
		c.validateRequire(vctx, BotProtectionProviderCredentialsKey, "bot protection key materials")
		c.validateBotProtectionSecrets(vctx, appConfig.BotProtection.Provider)
	}
	if appConfig.Identity.LDAP != nil && len(appConfig.Identity.LDAP.Servers) > 0 {
		c.validateRequire(vctx, LDAPServerUserCredentialsKey, "LDAP server user credentials")
		c.validateLDAPServerUserSecrets(vctx, appConfig.Identity.LDAP.Servers)
	}

	if len(appConfig.SAML.ServiceProviders) > 0 {
		c.validateRequire(vctx, SAMLIdpSigningMaterialsKey, "saml idp signing key materials")
		for _, sp := range appConfig.SAML.ServiceProviders {
			if sp.SignatureVerificationEnabled {
				c.validateSAMLServiceProviderCerts(vctx, sp)
			}
		}
	}
	if appConfig.SAML.Signing.KeyID != "" {
		c.validateSAMLSigningKey(vctx, appConfig.SAML.Signing.KeyID)
	}

	idx, demoCredentials, ok := c.LookupDataWithIndex(SSOOAuthDemoCredentialsKey)
	if ok {
		childCtx := vctx.Child("secrets", strconv.Itoa(idx), "data")
		c.validateSSOOAuthDemoCredentials(ctx, childCtx, demoCredentials.(*SSOOAuthDemoCredentials))
	}

	return vctx.Error("invalid secrets")
}

func (c *SecretConfig) GetCustomSMSProviderConfig() *CustomSMSProviderConfig {
	s, _ := c.LookupData(CustomSMSProviderConfigKey).(*CustomSMSProviderConfig)
	return s
}

type SecretKey string

const (
	DatabaseCredentialsKey       SecretKey = "db"
	AuditDatabaseCredentialsKey  SecretKey = "audit.db"
	ElasticsearchCredentialsKey  SecretKey = "elasticsearch"
	SearchDatabaseCredentialsKey SecretKey = "search.db"
	RedisCredentialsKey          SecretKey = "redis"
	// nolint: gosec
	AnalyticRedisCredentialsKey SecretKey = "analytic.redis"
	AdminAPIAuthKeyKey          SecretKey = "admin-api.auth"
	// nolint: gosec
	OAuthSSOProviderCredentialsKey SecretKey = "sso.oauth.client"
	// nolint: gosec
	SSOOAuthDemoCredentialsKey SecretKey = "sso.oauth.demo_credentials"
	SMTPServerCredentialsKey   SecretKey = "mail.smtp"
	// nolint: gosec
	TwilioCredentialsKey SecretKey = "sms.twilio"
	// nolint: gosec
	NexmoCredentialsKey        SecretKey = "sms.nexmo"
	CustomSMSProviderConfigKey SecretKey = "sms.custom"
	OAuthKeyMaterialsKey       SecretKey = "oauth"
	CSRFKeyMaterialsKey        SecretKey = "csrf"
	WebhookKeyMaterialsKey     SecretKey = "webhook"
	ImagesKeyMaterialsKey      SecretKey = "images"
	WATICredentialsKey         SecretKey = "whatsapp.wati"
	// nolint: gosec
	OAuthClientCredentialsKey SecretKey = "oauth.client_secrets"
	// nolint: gosec
	Deprecated_CaptchaCloudflareCredentialsKey SecretKey = "captcha.cloudflare"
	BotProtectionProviderCredentialsKey        SecretKey = "bot_protection.provider"
	WhatsappOnPremisesCredentialsKey           SecretKey = "whatsapp.on-premises"
	WhatsappCloudAPICredentialsKey             SecretKey = "whatsapp.cloud-api"
	LDAPServerUserCredentialsKey               SecretKey = "ldap"

	SAMLIdpSigningMaterialsKey SecretKey = "saml.idp.signing"
	SAMLSpSigningMaterialsKey  SecretKey = "saml.service_providers.signing"
)

func (key SecretKey) IsUpdatable() bool {
	switch key {
	case OAuthSSOProviderCredentialsKey,
		SMTPServerCredentialsKey:
		return true
	default:
		return false
	}
}

type SecretItemData interface {
	SensitiveStrings() []string
}

type secretKeyDef struct {
	schemaID    string
	dataFactory func() SecretItemData
}

var secretItemKeys = map[SecretKey]secretKeyDef{
	DatabaseCredentialsKey:                     {"DatabaseCredentials", func() SecretItemData { return &DatabaseCredentials{} }},
	AuditDatabaseCredentialsKey:                {"AuditDatabaseCredentials", func() SecretItemData { return &AuditDatabaseCredentials{} }},
	SearchDatabaseCredentialsKey:               {"SearchDatabaseCredentials", func() SecretItemData { return &SearchDatabaseCredentials{} }},
	ElasticsearchCredentialsKey:                {"ElasticsearchCredentials", func() SecretItemData { return &ElasticsearchCredentials{} }},
	RedisCredentialsKey:                        {"RedisCredentials", func() SecretItemData { return &RedisCredentials{} }},
	AnalyticRedisCredentialsKey:                {"AnalyticRedisCredentials", func() SecretItemData { return &AnalyticRedisCredentials{} }},
	AdminAPIAuthKeyKey:                         {"AdminAPIAuthKey", func() SecretItemData { return &AdminAPIAuthKey{} }},
	OAuthSSOProviderCredentialsKey:             {"OAuthSSOProviderCredentials", func() SecretItemData { return &OAuthSSOProviderCredentials{} }},
	SSOOAuthDemoCredentialsKey:                 {"SSOOAuthDemoCredentials", func() SecretItemData { return &SSOOAuthDemoCredentials{} }},
	SMTPServerCredentialsKey:                   {"SMTPServerCredentials", func() SecretItemData { return &SMTPServerCredentials{} }},
	TwilioCredentialsKey:                       {"TwilioCredentials", func() SecretItemData { return &TwilioCredentials{} }},
	NexmoCredentialsKey:                        {"NexmoCredentials", func() SecretItemData { return &NexmoCredentials{} }},
	OAuthKeyMaterialsKey:                       {"OAuthKeyMaterials", func() SecretItemData { return &OAuthKeyMaterials{} }},
	CSRFKeyMaterialsKey:                        {"CSRFKeyMaterials", func() SecretItemData { return &CSRFKeyMaterials{} }},
	WebhookKeyMaterialsKey:                     {"WebhookKeyMaterials", func() SecretItemData { return &WebhookKeyMaterials{} }},
	ImagesKeyMaterialsKey:                      {"ImagesKeyMaterials", func() SecretItemData { return &ImagesKeyMaterials{} }},
	WATICredentialsKey:                         {"WATICredentials", func() SecretItemData { return &WATICredentials{} }},
	OAuthClientCredentialsKey:                  {"OAuthClientCredentials", func() SecretItemData { return &OAuthClientCredentials{} }},
	CustomSMSProviderConfigKey:                 {"CustomSMSProviderConfig", func() SecretItemData { return &CustomSMSProviderConfig{} }},
	Deprecated_CaptchaCloudflareCredentialsKey: {"Deprecated_CaptchaCloudflareCredentials", func() SecretItemData { return &Deprecated_CaptchaCloudflareCredentials{} }},
	BotProtectionProviderCredentialsKey:        {"BotProtectionProviderCredentials", func() SecretItemData { return &BotProtectionProviderCredentials{} }},
	WhatsappOnPremisesCredentialsKey:           {"WhatsappOnPremisesCredentials", func() SecretItemData { return &WhatsappOnPremisesCredentials{} }},
	WhatsappCloudAPICredentialsKey:             {"WhatsappCloudAPICredentials", func() SecretItemData { return &WhatsappCloudAPICredentials{} }},
	LDAPServerUserCredentialsKey:               {"LDAPServerUserCredentials", func() SecretItemData { return &LDAPServerUserCredentials{} }},
	SAMLIdpSigningMaterialsKey:                 {"SAMLIdpSigningMaterials", func() SecretItemData { return &SAMLIdpSigningMaterials{} }},
	SAMLSpSigningMaterialsKey:                  {"SAMLSpSigningMaterials", func() SecretItemData { return &SAMLSpSigningMaterials{} }},
}

var _ = SecretConfigSchema.AddJSON("SecretKey", map[string]interface{}{
	"type": "string",
	"enum": func() []string {
		var keys []string
		for key := range secretItemKeys {
			keys = append(keys, string(key))
		}
		sort.Strings(keys)
		return keys
	}(),
})

var _ = SecretConfigSchema.AddJSON("SecretItem", map[string]interface{}{
	"type":                 "object",
	"additionalProperties": false,
	"properties": map[string]interface{}{
		"key":  map[string]interface{}{"$ref": "#/$defs/SecretKey"},
		"data": map[string]interface{}{},
	},
	"allOf": func() []interface{} {
		var keys []string
		for key := range secretItemKeys {
			keys = append(keys, string(key))
		}
		sort.Strings(keys)

		var schemas []interface{}
		for _, key := range keys {
			schemas = append(schemas, map[string]interface{}{
				"if": map[string]interface{}{
					"properties": map[string]interface{}{
						"key": map[string]interface{}{"const": string(key)},
					},
				},
				"then": map[string]interface{}{
					"properties": map[string]interface{}{
						"data": map[string]interface{}{"$ref": "#/$defs/" + secretItemKeys[SecretKey(key)].schemaID},
					},
				},
			})
		}
		return schemas
	}(),
	"required": []string{"key", "data"},
})

type SecretItem struct {
	Key     SecretKey        `json:"key,omitempty"`
	RawData json.RawMessage  `json:"data,omitempty"`
	Data    SecretItemData   `json:"-"`
	FsLevel resource.FsLevel `json:"-"`
}

func (i *SecretItem) parse(ctx context.Context, vctx *validation.Context) {
	def, ok := secretItemKeys[i.Key]
	if !ok {
		vctx.Child("key").EmitErrorMessage("unknown secret key")
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(i.RawData))
	data := def.dataFactory()
	err := decoder.Decode(data)
	if err != nil {
		vctx.Child("data").AddError(err)
		return
	}

	SetFieldDefaults(data)

	err = validation.ValidateValue(ctx, data)
	if err != nil {
		vctx.Child("data").AddError(err)
		return
	}

	i.Data = data
}
