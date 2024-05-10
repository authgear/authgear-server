package oauthrelyingparty

import (
	"fmt"
)

type ProviderConfig map[string]interface{}

func (c ProviderConfig) MustGetProvider() Provider {
	typ := c.Type()
	p, ok := registry[typ]
	if !ok {
		panic(fmt.Errorf("oauth provider not in registry: %v", typ))
	}
	return p
}

func (c ProviderConfig) Alias() string {
	alias, _ := c["alias"].(string)
	return alias
}

func (c ProviderConfig) Type() string {
	typ, _ := c["type"].(string)
	return typ
}

func (c ProviderConfig) ClientID() string {
	client_id, _ := c["client_id"].(string)
	return client_id
}

func (c ProviderConfig) SetDefaultsModifyDisabledFalse() {
	_, ok := c["modify_disabled"].(bool)
	if !ok {
		c["modify_disabled"] = false
	}
}

func (c ProviderConfig) ModifyDisabled() bool {
	modify_disabled, _ := c["modify_disabled"].(bool)
	return modify_disabled
}

func (c ProviderConfig) SetDefaultsEmailClaimConfig(claim ProviderClaimConfig) {
	claims, ok := c["claims"].(map[string]interface{})
	if !ok {
		claims = map[string]interface{}{}
		c["claims"] = claims
	}

	email, ok := claims["email"].(map[string]interface{})
	if !ok {
		claims["email"] = map[string]interface{}(claim)
	} else {
		if _, ok := email["assume_verified"].(bool); !ok {
			email["assume_verified"] = claim.AssumeVerified()
		}
		if _, ok := email["required"].(bool); !ok {
			email["required"] = claim.Required()
		}
	}
}

func (c ProviderConfig) EmailClaimConfig() ProviderClaimConfig {
	claims, ok := c["claims"].(map[string]interface{})
	if !ok {
		return ProviderClaimConfig{}
	}
	email, ok := claims["email"].(map[string]interface{})
	if !ok {
		return ProviderClaimConfig{}
	}
	return ProviderClaimConfig(email)
}

func (c ProviderConfig) SetDefaults() {
	provider := c.MustGetProvider()
	provider.SetDefaults(c)
}

func (c ProviderConfig) Scope() []string {
	provider := c.MustGetProvider()
	return provider.Scope(c)
}

func (c ProviderConfig) ProviderID() ProviderID {
	provider := c.MustGetProvider()
	return provider.ProviderID(c)
}

type ProviderClaimConfig map[string]interface{}

func (c ProviderClaimConfig) AssumeVerified() bool {
	b, _ := c["assume_verified"].(bool)
	return b
}

func (c ProviderClaimConfig) Required() bool {
	b, _ := c["required"].(bool)
	return b
}

// ProviderID combining with a subject ID identifies an user from an external system.
type ProviderID struct {
	Type string
	Keys map[string]interface{}
}

func NewProviderID(typ string, keys map[string]interface{}) ProviderID {
	id := ProviderID{
		Keys: make(map[string]interface{}),
	}
	id.Type = typ
	for k, v := range keys {
		id.Keys[k] = v
	}
	return id
}

func (i ProviderID) Equal(that ProviderID) bool {
	if i.Type != that.Type {
		return false
	}
	if len(i.Keys) != len(that.Keys) {
		return false
	}
	for key, thisValue := range i.Keys {
		thatValue, ok := that.Keys[key]
		if !ok {
			return false
		}
		if thisValue != thatValue {
			return false
		}
	}
	return true
}

type GetAuthorizationURLOptions struct {
	RedirectURI  string
	ResponseMode string
	Nonce        string
	State        string
	Prompt       []string
}

type GetUserProfileOptions struct {
	Code        string
	RedirectURI string
	Nonce       string
}

type UserProfile struct {
	ProviderRawProfile map[string]interface{}
	// ProviderUserID is not necessarily equal to sub.
	// If there exists a more unique identifier than sub, that identifier is chosen instead.
	ProviderUserID     string
	StandardAttributes map[string]interface{}
}

type Provider interface {
	SetDefaults(cfg ProviderConfig)
	ProviderID(cfg ProviderConfig) ProviderID
	Scope(cfg ProviderConfig) []string
}
