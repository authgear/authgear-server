package config

import (
	"bytes"
	"encoding/json"
	"reflect"

	"sigs.k8s.io/yaml"
)

var _ = FeatureConfigSchema.Add("FeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"identity": { "$ref": "#/$defs/IdentityFeatureConfig" },
		"authentication": { "$ref": "#/$defs/AuthenticationFeatureConfig" },
		"authenticator": { "$ref": "#/$defs/AuthenticatorFeatureConfig" },
		"custom_domain": { "$ref": "#/$defs/CustomDomainFeatureConfig" },
		"ui": { "$ref": "#/$defs/UIFeatureConfig" },
		"oauth": { "$ref": "#/$defs/OAuthFeatureConfig" },
		"hook": { "$ref": "#/$defs/HookFeatureConfig" },
		"audit_log": { "$ref": "#/$defs/AuditLogFeatureConfig" },
		"google_tag_manager": { "$ref": "#/$defs/GoogleTagManagerFeatureConfig" },
		"rate_limits": { "$ref": "#/$defs/RateLimitsFeatureConfig" },
		"messaging": { "$ref": "#/$defs/MessagingFeatureConfig" },
		"collaborator": { "$ref": "#/$defs/CollaboratorFeatureConfig" },
		"web3": { "$ref": "#/$defs/Web3FeatureConfig" },
		"admin_api": { "$ref": "#/$defs/AdminAPIFeatureConfig" },
		"test_mode": { "$ref": "#/$defs/TestModeFeatureConfig" }
	}
}
`)

type FeatureConfig struct {
	Identity         *IdentityFeatureConfig         `json:"identity,omitempty"`
	Authentication   *AuthenticationFeatureConfig   `json:"authentication,omitempty"`
	Authenticator    *AuthenticatorFeatureConfig    `json:"authenticator,omitempty"`
	CustomDomain     *CustomDomainFeatureConfig     `json:"custom_domain,omitempty"`
	UI               *UIFeatureConfig               `json:"ui,omitempty"`
	OAuth            *OAuthFeatureConfig            `json:"oauth,omitempty"`
	Hook             *HookFeatureConfig             `json:"hook,omitempty"`
	AuditLog         *AuditLogFeatureConfig         `json:"audit_log,omitempty"`
	GoogleTagManager *GoogleTagManagerFeatureConfig `json:"google_tag_manager,omitempty"`
	RateLimits       *RateLimitsFeatureConfig       `json:"rate_limits,omitempty"`
	Messaging        *MessagingFeatureConfig        `json:"messaging,omitempty"`
	Collaborator     *CollaboratorFeatureConfig     `json:"collaborator,omitempty"`
	Deprecated_Web3  *Deprecated_Web3FeatureConfig  `json:"web3,omitempty"`
	AdminAPI         *AdminAPIFeatureConfig         `json:"admin_api,omitempty"`
	TestMode         *TestModeFeatureConfig         `json:"test_mode,omitempty"`
}

func (c *FeatureConfig) Merge(layer *FeatureConfig) *FeatureConfig {
	t := reflect.TypeOf(*c)
	v := reflect.ValueOf(c).Elem()
	numField := t.NumField()
	for j := 0; j < numField; j++ {
		field := v.Field(j)
		if mergeable, ok := field.Interface().(MergeableFeatureConfig); ok {
			newValue := mergeable.Merge(layer)
			newV := reflect.ValueOf(newValue).Elem()
			if newV.CanAddr() {
				field.Set(newV.Addr())
			}
		}
	}

	newFeatureConfig := v.Interface().(FeatureConfig)
	return &newFeatureConfig
}

func ParseFeatureConfigWithoutDefaults(inputYAML []byte) (*FeatureConfig, error) {
	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}

	err = FeatureConfigSchema.Validator().ValidateWithMessage(
		bytes.NewReader(jsonData),
		"invalid feature config",
	)
	if err != nil {
		return nil, err
	}

	var config FeatureConfig
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func ParseFeatureConfig(inputYAML []byte) (*FeatureConfig, error) {
	config, err := ParseFeatureConfigWithoutDefaults(inputYAML)
	if err != nil {
		return nil, err
	}

	SetFieldDefaults(config)

	return config, nil
}

func NewEffectiveDefaultFeatureConfig() *FeatureConfig {
	config := FeatureConfig{}
	SetFieldDefaults(&config)
	return &config
}

func PopulateFeatureConfigDefaultValues(config *FeatureConfig) {
	SetFieldDefaults(config)
}

type MergeableFeatureConfig interface {
	Merge(layer *FeatureConfig) MergeableFeatureConfig
}
