package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/validation"
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

func (c *SecretConfig) Lookup(key SecretKey) (*SecretItem, bool) {
	for _, item := range c.Secrets {
		if item.Key == key {
			return &item, true
		}
	}
	return nil, false
}

func (c *SecretConfig) LookupData(key SecretKey) SecretItemData {
	if item, ok := c.Lookup(key); ok {
		return item.Data
	}
	return nil
}

func (c *SecretConfig) Validate(appConfig *AppConfig) error {
	ctx := &validation.Context{}
	require := func(key SecretKey, item string) {
		if _, ok := c.Lookup(key); !ok {
			ctx.EmitErrorMessage(fmt.Sprintf("%s (secret '%s') is required", item, key))
		}
	}

	require(DatabaseCredentialsKey, "database credentials")
	require(RedisCredentialsKey, "redis credentials")

	if len(appConfig.Identity.OAuth.Providers) > 0 {
		require(OAuthClientCredentialsKey, "OAuth client credentials")
		oauth, ok := c.LookupData(OAuthClientCredentialsKey).(*OAuthClientCredentials)
		if ok {
			for _, p := range appConfig.Identity.OAuth.Providers {
				found := false
				for _, item := range oauth.Items {
					if p.Alias == item.Alias {
						found = true
					}
				}
				if !found {
					ctx.EmitErrorMessage(fmt.Sprintf("OAuth client credentials for '%s' is required", p.Alias))
				}
			}
		}
	}

	require(JWTKeyMaterialsKey, "JWT key materials")
	require(OIDCKeyMaterialsKey, "OIDC key materials")
	require(CSRFKeyMaterialsKey, "CSRF key materials")
	if len(appConfig.Hook.Handlers) > 0 {
		require(WebhookKeyMaterialsKey, "web-hook signing key materials")
	}

	return ctx.Error("invalid secrets")
}

var _ = SecretConfigSchema.Add("SecretKey", `{ "type": "string" }`)

type SecretKey string

const (
	DatabaseCredentialsKey    SecretKey = "db"
	RedisCredentialsKey       SecretKey = "redis"
	OAuthClientCredentialsKey SecretKey = "sso.oauth.client"
	SMTPServerCredentialsKey  SecretKey = "mail.smtp"
	TwilioCredentialsKey      SecretKey = "sms.twilio"
	NexmoCredentialsKey       SecretKey = "sms.nexmo"
	JWTKeyMaterialsKey        SecretKey = "jwt"
	OIDCKeyMaterialsKey       SecretKey = "oidc"
	CSRFKeyMaterialsKey       SecretKey = "csrf"
	WebhookKeyMaterialsKey    SecretKey = "webhook"
)

type SecretItemData interface {
	SensitiveStrings() []string
}

var _ = SecretConfigSchema.Add("SecretItem", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"key": { "$ref": "#/$defs/SecretKey" },
		"data": { "type": "object" }
	},
	"required": ["key", "data"]
}
`)

type SecretItem struct {
	Key     SecretKey       `json:"key,omitempty"`
	RawData json.RawMessage `json:"data,omitempty"`
	Data    SecretItemData  `json:"-"`
}

func (i *SecretItem) parse(ctx *validation.Context) {
	r := bytes.NewReader(i.RawData)
	var data SecretItemData

	var validator *validation.SchemaValidator
	switch i.Key {
	case DatabaseCredentialsKey:
		validator = SecretConfigSchema.PartValidator("DatabaseCredentials")
		data = &DatabaseCredentials{}
	case RedisCredentialsKey:
		validator = SecretConfigSchema.PartValidator("RedisCredentials")
		data = &RedisCredentials{}
	case OAuthClientCredentialsKey:
		validator = SecretConfigSchema.PartValidator("OAuthClientCredentials")
		data = &OAuthClientCredentials{}
	case SMTPServerCredentialsKey:
		validator = SecretConfigSchema.PartValidator("SMTPServerCredentials")
		data = &SMTPServerCredentials{}
	case TwilioCredentialsKey:
		validator = SecretConfigSchema.PartValidator("TwilioCredentials")
		data = &TwilioCredentials{}
	case NexmoCredentialsKey:
		validator = SecretConfigSchema.PartValidator("NexmoCredentials")
		data = &NexmoCredentials{}
	case JWTKeyMaterialsKey:
		validator = SecretConfigSchema.PartValidator("JWTKeyMaterials")
		data = &JWTKeyMaterials{}
	case OIDCKeyMaterialsKey:
		validator = SecretConfigSchema.PartValidator("OIDCKeyMaterials")
		data = &OIDCKeyMaterials{}
	case CSRFKeyMaterialsKey:
		validator = SecretConfigSchema.PartValidator("CSRFKeyMaterials")
		data = &CSRFKeyMaterials{}
	case WebhookKeyMaterialsKey:
		validator = SecretConfigSchema.PartValidator("WebhookKeyMaterials")
		data = &WebhookKeyMaterials{}
	default:
		ctx.Child("key").EmitErrorMessage("unknown secret key")
		return
	}
	if err := validator.Validate(r); err != nil {
		ctx.Child("data").AddError(err)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(i.RawData))
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
