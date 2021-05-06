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

func (c *SecretConfig) Validate(appConfig *AppConfig) error {
	ctx := &validation.Context{}
	require := func(key SecretKey, item string) {
		if _, _, ok := c.Lookup(key); !ok {
			ctx.EmitErrorMessage(fmt.Sprintf("%s (secret '%s') is required", item, key))
		}
	}

	require(DatabaseCredentialsKey, "database credentials")
	require(ElasticsearchCredentialsKey, "elasticsearch credentials")
	require(RedisCredentialsKey, "redis credentials")
	require(AdminAPIAuthKeyKey, "admin API auth key materials")

	if len(appConfig.Identity.OAuth.Providers) > 0 {
		require(OAuthClientCredentialsKey, "OAuth client credentials")
		secretIndex, data, _ := c.LookupDataWithIndex(OAuthClientCredentialsKey)
		oauth, ok := data.(*OAuthClientCredentials)
		if ok {
			for _, p := range appConfig.Identity.OAuth.Providers {
				var matchedItem *OAuthClientCredentialsItem = nil
				var matchedItemIndex int = -1
				for index := range oauth.Items {
					item := oauth.Items[index]
					if p.Alias == item.Alias {
						matchedItem = &item
						matchedItemIndex = index
						break
					}
				}
				if matchedItem == nil {
					ctx.EmitErrorMessage(fmt.Sprintf("OAuth client credentials for '%s' is required", p.Alias))
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

	require(OAuthKeyMaterialsKey, "OAuth key materials")
	require(CSRFKeyMaterialsKey, "CSRF key materials")
	if len(appConfig.Hook.BlockingHandlers) > 0 || len(appConfig.Hook.NonBlockingHandlers) > 0 {
		require(WebhookKeyMaterialsKey, "web-hook signing key materials")
	}

	return ctx.Error("invalid secrets")
}

type SecretKey string

const (
	DatabaseCredentialsKey      SecretKey = "db"
	ElasticsearchCredentialsKey SecretKey = "elasticsearch"
	RedisCredentialsKey         SecretKey = "redis"
	AdminAPIAuthKeyKey          SecretKey = "admin-api.auth"
	OAuthClientCredentialsKey   SecretKey = "sso.oauth.client"
	SMTPServerCredentialsKey    SecretKey = "mail.smtp"
	TwilioCredentialsKey        SecretKey = "sms.twilio"
	NexmoCredentialsKey         SecretKey = "sms.nexmo"
	OAuthKeyMaterialsKey        SecretKey = "oauth"
	CSRFKeyMaterialsKey         SecretKey = "csrf"
	WebhookKeyMaterialsKey      SecretKey = "webhook"
)

type SecretItemData interface {
	SensitiveStrings() []string
}

type secretKeyDef struct {
	schemaID    string
	dataFactory func() SecretItemData
}

var secretItemKeys = map[SecretKey]secretKeyDef{
	DatabaseCredentialsKey:      {"DatabaseCredentials", func() SecretItemData { return &DatabaseCredentials{} }},
	ElasticsearchCredentialsKey: {"ElasticsearchCredentials", func() SecretItemData { return &ElasticsearchCredentials{} }},
	RedisCredentialsKey:         {"RedisCredentials", func() SecretItemData { return &RedisCredentials{} }},
	AdminAPIAuthKeyKey:          {"AdminAPIAuthKey", func() SecretItemData { return &AdminAPIAuthKey{} }},
	OAuthClientCredentialsKey:   {"OAuthClientCredentials", func() SecretItemData { return &OAuthClientCredentials{} }},
	SMTPServerCredentialsKey:    {"SMTPServerCredentials", func() SecretItemData { return &SMTPServerCredentials{} }},
	TwilioCredentialsKey:        {"TwilioCredentials", func() SecretItemData { return &TwilioCredentials{} }},
	NexmoCredentialsKey:         {"NexmoCredentials", func() SecretItemData { return &NexmoCredentials{} }},
	OAuthKeyMaterialsKey:        {"OAuthKeyMaterials", func() SecretItemData { return &OAuthKeyMaterials{} }},
	CSRFKeyMaterialsKey:         {"CSRFKeyMaterials", func() SecretItemData { return &CSRFKeyMaterials{} }},
	WebhookKeyMaterialsKey:      {"WebhookKeyMaterials", func() SecretItemData { return &WebhookKeyMaterials{} }},
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

	setFieldDefaults(data)

	err = validation.ValidateValue(data)
	if err != nil {
		ctx.Child("data").AddError(err)
		return
	}

	i.Data = data
}
