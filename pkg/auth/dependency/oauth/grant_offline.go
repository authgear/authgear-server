package oauth

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type OfflineGrant struct {
	AppID           string `json:"app_id"`
	ID              string `json:"id"`
	AuthorizationID string `json:"authz_id"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	Scopes    []string  `json:"scopes"`
	TokenHash string    `json:"token_hash"`

	AccessedAt    time.Time        `json:"accessed_at"`
	Attrs         authn.Attrs      `json:"attrs"`
	InitialAccess auth.AccessEvent `json:"initial_access"`
	LastAccess    auth.AccessEvent `json:"last_access"`
}

var _ Grant = OfflineGrant{}

func (g OfflineGrant) Session() (kind GrantSessionKind, id string) {
	return GrantSessionKindOffline, g.ID
}
