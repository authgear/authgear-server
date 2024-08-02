package config

import (
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

var _ = Schema.Add("AccountLinkingConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"oauth": {
			"type": "array",
			"items": { "$ref": "#/$defs/AccountLinkingOAuthItem" }
		},
		"login_id": {
			"type": "array",
			"items": { "$ref": "#/$defs/AccountLinkingLoginIDItem" }
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

var _ = Schema.Add("AccountLinkingLoginIDItem", `
{
	"type": "object",
	"required": ["key", "user_profile", "action"],
	"properties": {
		"name": { "type": "string" },
		"key": { "type": "string" },
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

// If you ever need to change this schema, please read https://linear.app/authgear/issue/DEV-1701/high-sql-cpu-utilization-when-running-list-by-claim-query
// Our database index only supports email, phone_number, and preferred_username.
// Other claims are NOT indexed, and they will result in very high CPU usage.
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
	OAuth   []*AccountLinkingOAuthItem   `json:"oauth,omitempty"`
	LoginID []*AccountLinkingLoginIDItem `json:"login_id,omitempty"`
}

type AccountLinkingLoginIDItem struct {
	Name        string                     `json:"name,omitempty"`
	Key         string                     `json:"key,omitempty"`
	UserProfile *AccountLinkingJSONPointer `json:"user_profile,omitempty"`
	Action      AccountLinkingAction       `json:"action,omitempty"`
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

func (p *AccountLinkingJSONPointer) MustGetOneLevelJSONPointerOrPanic() jsonpointer.T {
	pointer := jsonpointer.MustParse(p.Pointer)
	if len(pointer) != 1 {
		// See https://linear.app/authgear/issue/DEV-1701/high-sql-cpu-utilization-when-running-list-by-claim-query
		panic(fmt.Errorf("account linking json pointer must be /email, /phone_number or /preferred_username"))
	}
	return pointer
}

func (p *AccountLinkingJSONPointer) MustGetFirstLevelReferenceTokenOrPanic() string {
	pointer := p.MustGetOneLevelJSONPointerOrPanic()
	return pointer[0]
}

var DefaultAccountLinkingOAuthItem = &AccountLinkingOAuthItem{
	OAuthClaim:  &AccountLinkingJSONPointer{Pointer: "/email"},
	UserProfile: &AccountLinkingJSONPointer{Pointer: "/email"},
	Action:      AccountLinkingActionError,
}

var DefaultAccountLinkingLoginIDEmailItem = &AccountLinkingLoginIDItem{
	UserProfile: &AccountLinkingJSONPointer{Pointer: "/email"},
	Action:      AccountLinkingActionError,
}

var DefaultAccountLinkingLoginIDPhoneItem = &AccountLinkingLoginIDItem{
	UserProfile: &AccountLinkingJSONPointer{Pointer: "/phone_number"},
	Action:      AccountLinkingActionError,
}

var DefaultAccountLinkingLoginIDUsernameItem = &AccountLinkingLoginIDItem{
	UserProfile: &AccountLinkingJSONPointer{Pointer: "/preferred_username"},
	Action:      AccountLinkingActionError,
}
