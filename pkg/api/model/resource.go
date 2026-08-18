package model

// AccessPolicy controls whether a Resource/Scope is reachable by clients
// outside the static per-client association mechanism. Missing keys default
// to false, so the zero value is "no access" for every grant this policy
// governs.
type AccessPolicy struct {
	AllowDynamicThirdPartyClientAccess bool `json:"allow_dynamic_third_party_client_access,omitempty"`
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
