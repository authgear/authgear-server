package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"sigs.k8s.io/yaml"

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

// ParsePartialSecret unmarshals inputYAML into a full SecretConfig,
// without performing validation.
func ParsePartialSecret(inputYAML []byte) (*SecretConfig, error) {
	const validationErrorMessage = "invalid secrets"

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

	ctx := &validation.Context{}
	for i := range config.Secrets {
		config.Secrets[i].parse(ctx.Child("secrets", strconv.Itoa(i)))
	}
	if err := ctx.Error(validationErrorMessage); err != nil {
		return nil, err
	}

	return &config, nil
}

func ParseSecret(inputYAML []byte) (*SecretConfig, error) {
	const validationErrorMessage = "invalid secrets"

	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}

	err = SecretConfigSchema.Validator().ValidateWithMessage(
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

	ctx := &validation.Context{}
	for i := range config.Secrets {
		config.Secrets[i].parse(ctx.Child("secrets", strconv.Itoa(i)))
	}
	if err := ctx.Error(validationErrorMessage); err != nil {
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
				if matchedItem.ClientSecret == "" {
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

func (c *SecretConfig) Validate(appConfig *AppConfig) error {
	ctx := &validation.Context{}

	c.validateRequire(ctx, DatabaseCredentialsKey, "database credentials")
	// AuditDatabaseCredentialsKey is not required
	// ElasticsearchCredentialsKey is not required
	c.validateRequire(ctx, RedisCredentialsKey, "redis credentials")
	c.validateRequire(ctx, AdminAPIAuthKeyKey, "admin API auth key materials")

	if len(appConfig.Identity.OAuth.Providers) > 0 {
		c.validateOAuthProviders(ctx, appConfig)
	}

	confidentialClients := []OAuthClientConfig{}
	for _, c := range appConfig.OAuth.Clients {
		if c.IsConfidential() {
			confidentialClients = append(confidentialClients, c)
		}
	}

	if len(confidentialClients) > 0 {
		c.validateConfidentialClients(ctx, confidentialClients)
	}

	c.validateRequire(ctx, OAuthKeyMaterialsKey, "OAuth key materials")
	c.validateRequire(ctx, CSRFKeyMaterialsKey, "CSRF key materials")
	if len(appConfig.Hook.BlockingHandlers) > 0 || len(appConfig.Hook.NonBlockingHandlers) > 0 {
		c.validateRequire(ctx, WebhookKeyMaterialsKey, "web-hook signing key materials")
	}
	if appConfig.BotProtection != nil && appConfig.BotProtection.Enabled != nil && *appConfig.BotProtection.Enabled && appConfig.BotProtection.Provider != nil {
		c.validateRequire(ctx, BotProtectionProviderCredentialsKey, "bot protection key materials")
		c.validateBotProtectionSecrets(ctx, appConfig.BotProtection.Provider)
	}

	return ctx.Error("invalid secrets")
}

func (c *SecretConfig) GetCustomSMSProviderConfig() *CustomSMSProviderConfig {
	s, _ := c.LookupData(CustomSMSProviderConfigKey).(*CustomSMSProviderConfig)
	return s
}

type SecretKey string

const (
	DatabaseCredentialsKey      SecretKey = "db"
	AuditDatabaseCredentialsKey SecretKey = "audit.db"
	ElasticsearchCredentialsKey SecretKey = "elasticsearch"
	RedisCredentialsKey         SecretKey = "redis"
	// nolint: gosec
	AnalyticRedisCredentialsKey SecretKey = "analytic.redis"
	AdminAPIAuthKeyKey          SecretKey = "admin-api.auth"
	// nolint: gosec
	OAuthSSOProviderCredentialsKey SecretKey = "sso.oauth.client"
	SMTPServerCredentialsKey       SecretKey = "mail.smtp"
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
	ElasticsearchCredentialsKey:                {"ElasticsearchCredentials", func() SecretItemData { return &ElasticsearchCredentials{} }},
	RedisCredentialsKey:                        {"RedisCredentials", func() SecretItemData { return &RedisCredentials{} }},
	AnalyticRedisCredentialsKey:                {"AnalyticRedisCredentials", func() SecretItemData { return &AnalyticRedisCredentials{} }},
	AdminAPIAuthKeyKey:                         {"AdminAPIAuthKey", func() SecretItemData { return &AdminAPIAuthKey{} }},
	OAuthSSOProviderCredentialsKey:             {"OAuthSSOProviderCredentials", func() SecretItemData { return &OAuthSSOProviderCredentials{} }},
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
	Key     SecretKey       `json:"key,omitempty"`
	RawData json.RawMessage `json:"data,omitempty"`
	Data    SecretItemData  `json:"-"`
}

func (i *SecretItem) parse(ctx *validation.Context) {
	def, ok := secretItemKeys[i.Key]
	if !ok {
		ctx.Child("key").EmitErrorMessage("unknown secret key")
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(i.RawData))
	data := def.dataFactory()
	err := decoder.Decode(data)
	if err != nil {
		ctx.Child("data").AddError(err)
		return
	}

	SetFieldDefaults(data)

	err = validation.ValidateValue(data)
	if err != nil {
		ctx.Child("data").AddError(err)
		return
	}

	i.Data = data
}
