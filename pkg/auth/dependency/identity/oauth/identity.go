package oauth

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Identity struct {
	ID     string
	UserID string
	// (ProviderID.Type, ProviderID.Keys, ProviderSubjectID) together form a unique index.
	ProviderID        ProviderID
	ProviderSubjectID string
	UserProfile       map[string]interface{}
	Claims            map[string]interface{}
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ProviderID struct {
	Type string
	Keys map[string]interface{}
}

func NewProviderID(config config.OAuthProviderConfiguration) ProviderID {
	keys := map[string]interface{}{}
	if config.Tenant != "" {
		keys["tenant"] = config.Tenant
	}
	if config.TeamID != "" {
		keys["team_id"] = config.TeamID
	}

	return ProviderID{
		Type: string(config.Type),
		Keys: keys,
	}
}
