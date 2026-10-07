package oauth

import "time"

type AccessGrant struct {
	AppID           string `json:"app_id"`
	AuthorizationID string `json:"authz_id"`
	// UserID is the user this access token authenticates as -- the user who
	// authenticated the flow that issued it (the JWT "sub"). It is recorded
	// here so resolution can verify the bound session still belongs to this
	// same user; SessionID alone does not pin the user, which let a token
	// bound to another user's session (via id_token_hint) resolve as that
	// other user. Empty on grants created before this field existed.
	UserID      string           `json:"user_id"`
	SessionID   string           `json:"session_id"`
	SessionKind GrantSessionKind `json:"session_kind"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	Scopes    []string  `json:"scopes"`
	TokenHash string    `json:"token_hash"`
	// Only exist when session_kind is offline_grant
	// It does not change even the refresh token rotated
	InitialRefreshTokenHash string `json:"refresh_token_hash"`
	// ResourceURI is empty when the access token this grant backs was not
	// bound to a resource. Mirrors OfflineGrantRefreshToken.ResourceURI (the
	// same single-resource-per-token constraint applies), persisted here too
	// so the resource binding of an already-issued access token can be
	// recovered from the grant itself -- e.g. for introspection or a
	// revocation cascade -- without re-deriving it from the (possibly since
	// rotated or deleted) refresh token that originally produced it.
	ResourceURI string `json:"resource_uri,omitempty"`
}
