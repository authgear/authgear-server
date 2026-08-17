package dcr

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type InitialAccessTokenType string

const (
	InitialAccessTokenTypeThirdParty InitialAccessTokenType = "THIRD_PARTY"
	InitialAccessTokenTypeFirstParty InitialAccessTokenType = "FIRST_PARTY"
)

const DefaultInitialAccessTokenExpiresIn = 3600 // seconds, per spec "e.g. 3600"

type InitialAccessToken struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
	Type      InitialAccessTokenType
	TokenHash string
}

func (t *InitialAccessToken) ToModel() *model.OAuthInitialAccessToken {
	return &model.OAuthInitialAccessToken{
		Meta: model.Meta{
			ID:        t.ID,
			CreatedAt: t.CreatedAt,
			// An IAT is immutable; UpdatedAt exists only to satisfy
			// model.Meta / EntityRef and is never exposed in GraphQL.
			UpdatedAt: t.CreatedAt,
		},
		ExpiresAt: t.ExpiresAt,
		Type:      model.OAuthInitialAccessTokenType(t.Type),
	}
}

type NewInitialAccessTokenOptions struct {
	ExpiresIn *int // seconds; nil means DefaultInitialAccessTokenExpiresIn
	Type      InitialAccessTokenType
}
