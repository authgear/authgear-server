package model

import "time"

type OAuthInitialAccessTokenType string

const (
	OAuthInitialAccessTokenTypeThirdParty OAuthInitialAccessTokenType = "THIRD_PARTY"
	OAuthInitialAccessTokenTypeFirstParty OAuthInitialAccessTokenType = "FIRST_PARTY"
)

// OAuthInitialAccessToken is the Admin API-facing representation.
//
// model.Meta MUST be embedded even though the GraphQL InitialAccessToken type
// (docs/specs/dcr.md) exposes no updatedAt field. pkg/admin/graphql's
// entityIDField and entityCreatedAtField both do an unchecked
// obj.(EntityRef) assertion, where EntityRef is `GetMeta() model.Meta`
// (pkg/admin/graphql/entity.go) — a model without Meta panics at resolve
// time, not compile time. UpdatedAt is set equal to CreatedAt (an IAT is
// immutable) and never surfaced.
type OAuthInitialAccessToken struct {
	Meta
	ExpiresAt time.Time
	Type      OAuthInitialAccessTokenType
}
