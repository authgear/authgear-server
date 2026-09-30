package model

type Authorization struct {
	Meta

	ClientID string `json:"clientID"`
	// Deprecated: Scopes holds the project-level scopes only. Use AuthorizedScopes.
	Scopes           []string          `json:"scopes"`
	AuthorizedScopes []AuthorizedScope `json:"authorizedScopes"`
}

type AuthorizedScope struct {
	// ResourceID is empty for a project-level scope.
	ResourceID string `json:"resourceID,omitempty"`
	Scope      string `json:"scope"`
}
