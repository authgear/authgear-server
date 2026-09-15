package model

// AccessPolicy controls whether a Resource/Scope is reachable by clients
// outside the static per-client association mechanism. Missing keys default
// to false, so the zero value is "no access" for every grant this policy
// governs.
type AccessPolicy struct {
	AllowStaticFirstPartyClientAccess  bool `json:"allow_static_first_party_client_access,omitempty"`
	AllowStaticThirdPartyClientAccess  bool `json:"allow_static_third_party_client_access,omitempty"`
	AllowDynamicFirstPartyClientAccess bool `json:"allow_dynamic_first_party_client_access,omitempty"`
	AllowDynamicThirdPartyClientAccess bool `json:"allow_dynamic_third_party_client_access,omitempty"`
}

// ClientCategoryClassifier is the client shape AllowsClient needs to place
// a client into one of the spec's four categories
// (docs/specs/api-resource.md § Client categories). Declared here instead
// of accepting *config.OAuthClientConfig directly, since pkg/lib/config
// imports this package; config.OAuthClientConfig satisfies it structurally.
type ClientCategoryClassifier interface {
	IsDynamicClient() bool
	IsThirdParty() bool
}

// AllowsClient reports whether p opens this Resource or Scope to client.
//
// An m2m client is neither dynamic nor third-party and so would read
// AllowStaticFirstPartyClientAccess here, which would be wrong -- m2m is in
// no category at all. Nothing calls this for one: client_credentials is its
// only grant, and access_policy does not govern that grant.
func (p AccessPolicy) AllowsClient(client ClientCategoryClassifier) bool {
	switch {
	case client.IsDynamicClient() && client.IsThirdParty():
		return p.AllowDynamicThirdPartyClientAccess
	case client.IsDynamicClient():
		return p.AllowDynamicFirstPartyClientAccess
	case client.IsThirdParty():
		return p.AllowStaticThirdPartyClientAccess
	default:
		return p.AllowStaticFirstPartyClientAccess
	}
}

// AccessPolicyPatch is a partial AccessPolicy. A nil field is left unchanged
// by an update, and defaults to false on create. Its json tags match
// AccessPolicy's exactly -- the store merges a marshalled patch straight
// into the stored JSONB object (see resourcescope.Store.UpdateResource).
type AccessPolicyPatch struct {
	AllowStaticFirstPartyClientAccess  *bool `json:"allow_static_first_party_client_access,omitempty"`
	AllowStaticThirdPartyClientAccess  *bool `json:"allow_static_third_party_client_access,omitempty"`
	AllowDynamicFirstPartyClientAccess *bool `json:"allow_dynamic_first_party_client_access,omitempty"`
	AllowDynamicThirdPartyClientAccess *bool `json:"allow_dynamic_third_party_client_access,omitempty"`
}

// Apply returns base with every non-nil field of p overlaid. A nil receiver
// returns base unchanged, so a create path can call
// options.AccessPolicy.Apply(AccessPolicy{}) whether or not a patch was
// given.
func (p *AccessPolicyPatch) Apply(base AccessPolicy) AccessPolicy {
	if p == nil {
		return base
	}
	if p.AllowStaticFirstPartyClientAccess != nil {
		base.AllowStaticFirstPartyClientAccess = *p.AllowStaticFirstPartyClientAccess
	}
	if p.AllowStaticThirdPartyClientAccess != nil {
		base.AllowStaticThirdPartyClientAccess = *p.AllowStaticThirdPartyClientAccess
	}
	if p.AllowDynamicFirstPartyClientAccess != nil {
		base.AllowDynamicFirstPartyClientAccess = *p.AllowDynamicFirstPartyClientAccess
	}
	if p.AllowDynamicThirdPartyClientAccess != nil {
		base.AllowDynamicThirdPartyClientAccess = *p.AllowDynamicThirdPartyClientAccess
	}
	return base
}

type Resource struct {
	Meta
	ResourceURI  string       `json:"resourceURI"`
	Name         *string      `json:"name,omitzero"`
	AccessPolicy AccessPolicy `json:"accessPolicy"`
}

type Scope struct {
	Meta
	ResourceID   string       `json:"resource_id"`
	Scope        string       `json:"scope"`
	Description  *string      `json:"description,omitzero"`
	AccessPolicy AccessPolicy `json:"accessPolicy"`
}
