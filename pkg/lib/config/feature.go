package config

import (
	"bytes"
	"encoding/json"

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
		"rate_limit": { "$ref": "#/$defs/RateLimitFeatureConfig" },
		"collaborator": { "$ref": "#/$defs/CollaboratorFeatureConfig" },
		"web3": { "$ref": "#/$defs/Web3FeatureConfig" }
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
	RateLimit        *RateLimitFeatureConfig        `json:"rate_limit,omitempty"`
	Collaborator     *CollaboratorFeatureConfig     `json:"collaborator,omitempty"`
	Web3             *Web3FeatureConfig             `json:"web3,omitempty"`
}

func ParseFeatureConfig(inputYAML []byte) (*FeatureConfig, error) {
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

	setFieldDefaults(&config)

	return &config, nil
}

func NewEffectiveDefaultFeatureConfig() *FeatureConfig {
	config := FeatureConfig{}
	setFieldDefaults(&config)
	return &config
}

func PopulateFeatureConfigDefaultValues(config *FeatureConfig) {
	setFieldDefaults(config)
}
