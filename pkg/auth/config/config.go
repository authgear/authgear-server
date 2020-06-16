package config

import (
	"bytes"
	"encoding/json"
	"sigs.k8s.io/yaml"
)

var _ = Schema.Add("AppConfig", `
{
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"metadata": { "$ref": "#/$defs/AppMetadata" },
		"http": { "$ref": "#/$defs/HTTPConfig" },
		"hook": { "$ref": "#/$defs/HookConfig" },
		"template": { "$ref": "#/$defs/TemplateConfig" },
		"ui": { "$ref": "#/$defs/UIConfig" },
		"authentication": { "$ref": "#/$defs/AuthenticationConfig" },
		"session": { "$ref": "#/$defs/SessionConfig" },
		"oauth": { "$ref": "#/$defs/OAuthConfig" },
		"identity": { "$ref": "#/$defs/IdentityConfig" },
		"authenticator": { "$ref": "#/$defs/AuthenticatorConfig" },
		"forgot_password": { "$ref": "#/$defs/ForgotPasswordConfig" },
		"welcome_message": { "$ref": "#/$defs/WelcomeMessageConfig" }
	},
	"required": ["id"]
}
`)

type AppConfig struct {
	ID       string      `json:"id"`
	Metadata AppMetadata `json:"metadata,omitempty"`

	HTTP *HTTPConfig `json:"http,omitempty"`
	Hook *HookConfig `json:"hook,omitempty"`

	Template *TemplateConfig `json:"template,omitempty"`
	UI       *UIConfig       `json:"ui,omitempty"`

	Authentication *AuthenticationConfig `json:"authentication,omitempty"`
	Session        *SessionConfig        `json:"session,omitempty"`
	OAuth          *OAuthConfig          `json:"oauth,omitempty"`
	Identity       *IdentityConfig       `json:"identity,omitempty"`
	Authenticator  *AuthenticatorConfig  `json:"authenticator,omitempty"`

	ForgotPassword *ForgotPasswordConfig `json:"forgot_password,omitempty"`
	WelcomeMessage *WelcomeMessageConfig `json:"welcome_message,omitempty"`
}

func Parse(inputYAML []byte) (*AppConfig, error) {
	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}

	err = Schema.ValidateReader(bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	var config AppConfig
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
