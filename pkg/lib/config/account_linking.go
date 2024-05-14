package config

import "github.com/iawaknahc/jsonschema/pkg/jsonpointer"

var _ = Schema.Add("AccountLinkingConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"oauth": {
			"type": "array",
			"items": { "$ref": "#/$defs/AccountLinkingOAuthItem" }
		}
	}
}
`)

var _ = Schema.Add("AccountLinkingOAuthItem", `
{
	"type": "object",
	"required": ["alias", "oauth_claim", "user_profile", "action"],
	"properties": {
		"name": { "type": "string" },
		"alias": { "type": "string" },
		"oauth_claim": { "$ref": "#/$defs/AccountLinkingJSONPointer" },
		"user_profile": { "$ref": "#/$defs/AccountLinkingJSONPointer" },
		"action": { "$ref": "#/$defs/AccountLinkingAction" }
	}
}
`)

var _ = Schema.Add("AccountLinkingAction", `
{
	"type": "string",
	"enum": [
		"error",
		"login_and_link"
	]
}
`)

var _ = Schema.Add("AccountLinkingJSONPointer", `
{
	"type": "object",
	"required": ["pointer"],
	"additionalProperties": false,
	"properties": {
		"pointer": {
			"type": "string",
			"format": "json-pointer",
			"enum": [
				"/email",
				"/phone_number",
				"/preferred_username"
			]
		}
	}
}
`)

type AccountLinkingConfig struct {
	OAuth []*AccountLinkingOAuthItem `json:"oauth,omitempty"`
}

type AccountLinkingOAuthItem struct {
	Name        string                     `json:"name,omitempty"`
	Alias       string                     `json:"alias,omitempty"`
	OAuthClaim  *AccountLinkingJSONPointer `json:"oauth_claim,omitempty"`
	UserProfile *AccountLinkingJSONPointer `json:"user_profile,omitempty"`
	Action      AccountLinkingAction       `json:"action,omitempty"`
}

type AccountLinkingAction string

const (
	AccountLinkingActionError        AccountLinkingAction = "error"
	AccountLinkingActionLoginAndLink AccountLinkingAction = "login_and_link"
)

type AccountLinkingJSONPointer struct {
	Pointer string `json:"pointer,omitempty"`
}

func (p *AccountLinkingJSONPointer) GetJSONPointer() jsonpointer.T {
	return jsonpointer.MustParse(p.Pointer)
}

var DefaultAccountLinkingOAuthItem = &AccountLinkingOAuthItem{
	OAuthClaim:  &AccountLinkingJSONPointer{Pointer: "/email"},
	UserProfile: &AccountLinkingJSONPointer{Pointer: "/email"},
	Action:      AccountLinkingActionError,
}
