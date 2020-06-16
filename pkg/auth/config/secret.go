package config

import (
	"bytes"
	"encoding/json"
	"strconv"

	"sigs.k8s.io/yaml"

	"github.com/skygeario/skygear-server/pkg/validation"
)

var _ = SecretConfigSchema.Add("SecretConfig", `
{
	"type": "object",
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
	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}

	err = SecretConfigSchema.ValidateReader(bytes.NewReader(jsonData))
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
	if err := ctx.Error(); err != nil {
		return nil, err
	}

	return &config, nil
}

var _ = SecretConfigSchema.Add("SecretKey", `{ "type": "string" }`)

type SecretKey string

const (
	DatabaseCredentialsKey   SecretKey = "db"
	RedisCredentialsKey      SecretKey = "redis"
	SMTPServerCredentialsKey SecretKey = "mail.smtp"
	TwilioCredentialsKey     SecretKey = "sms.twilio"
	NexmoCredentialsKey      SecretKey = "sms.nexmo"
	JWTKeyMaterialsKey       SecretKey = "jwt.keys"
	OIDCKeyMaterialsKey      SecretKey = "oidc.keys"
)

var _ = SecretConfigSchema.Add("SecretItem", `
{
	"type": "object",
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
	Data    interface{}     `json:"-"`
}

func (i *SecretItem) parse(ctx *validation.Context) {
	var err error
	r := bytes.NewReader(i.RawData)
	var data interface{}

	switch i.Key {
	case DatabaseCredentialsKey:
		err = SecretConfigSchema.ValidateReaderByPart(r, "DatabaseCredentials")
		data = &DatabaseCredentials{}
	case RedisCredentialsKey:
		err = SecretConfigSchema.ValidateReaderByPart(r, "RedisCredentials")
		data = &RedisCredentials{}
	case SMTPServerCredentialsKey:
		err = SecretConfigSchema.ValidateReaderByPart(r, "SMTPServerCredentials")
		data = &SMTPServerCredentials{}
	case TwilioCredentialsKey:
		err = SecretConfigSchema.ValidateReaderByPart(r, "TwilioCredentials")
		data = &TwilioCredentials{}
	case NexmoCredentialsKey:
		err = SecretConfigSchema.ValidateReaderByPart(r, "NexmoCredentials")
		data = &NexmoCredentials{}
	case JWTKeyMaterialsKey:
		err = SecretConfigSchema.ValidateReaderByPart(r, "JWTKeyMaterials")
		data = &JWTKeyMaterials{}
	case OIDCKeyMaterialsKey:
		err = SecretConfigSchema.ValidateReaderByPart(r, "OIDCKeyMaterials")
		data = &OIDCKeyMaterials{}
	default:
		ctx.Child("key").EmitErrorMessage("unknown secret key")
		return
	}
	if err != nil {
		ctx.Child("data").AddError(err)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(i.RawData))
	err = decoder.Decode(data)
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
