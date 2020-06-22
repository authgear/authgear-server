package oauth

import (
	"reflect"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	oldconfig "github.com/skygeario/skygear-server/pkg/core/config"
)

type Identity struct {
	ID                string
	UserID            string
	ProviderID        config.ProviderID
	ProviderSubjectID string
	UserProfile       map[string]interface{}
	Claims            map[string]interface{}
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// FIXME: remove this
type ProviderID struct {
	Type string
	Keys map[string]interface{}
}

func NewProviderID(c oldconfig.OAuthProviderConfiguration) ProviderID {
	keys := map[string]interface{}{}
	switch c.Type {
	case oldconfig.OAuthProviderTypeGoogle:
		// Google supports OIDC.
		// sub is public, not scoped to anything so changing client_id does not affect sub.
		// Therefore, ProviderID is simply Type.
		//
		// Rotating the OAuth application is OK.
		break
	case oldconfig.OAuthProviderTypeFacebook:
		// Facebook does NOT support OIDC.
		// Facebook user ID is scoped to client_id.
		// Therefore, ProviderID is Type + client_id.
		//
		// Rotating the OAuth application is problematic.
		// But if email remains unchanged, the user can associate their account.
		keys["client_id"] = c.ClientID
	case oldconfig.OAuthProviderTypeLinkedIn:
		// LinkedIn is the same as Facebook.
		keys["client_id"] = c.ClientID
	case oldconfig.OAuthProviderTypeAzureADv2:
		// Azure AD v2 supports OIDC.
		// sub is pairwise and is scoped to client_id.
		// However, oid is powerful alternative to sub.
		// oid is also pairwise and is scoped to tenant.
		// We use oid as ProviderSubjectID so ProviderID is Type + tenant.
		//
		// Rotating the OAuth application is OK.
		// But rotating the tenant is problematic.
		// But if email remains unchanged, the user can associate their account.
		keys["tenant"] = c.Tenant
	case oldconfig.OAuthProviderTypeApple:
		// Apple supports OIDC.
		// sub is pairwise and is scoped to team_id.
		// Therefore, ProviderID is Type + team_id.
		//
		// Rotating the OAuth application is OK.
		// But rotating the Apple Developer account is problematic.
		// Since Apple has private relay to hide the real email,
		// the user may not be associate their account.
		keys["team_id"] = c.TeamID
	}

	return ProviderID{
		Type: string(c.Type),
		Keys: keys,
	}
}

func (p *ProviderID) ClaimsValue() map[string]interface{} {
	claim := map[string]interface{}{}
	claim["type"] = p.Type
	for k, v := range p.Keys {
		claim[k] = v
	}
	return claim
}

func (p *ProviderID) Equal(that *ProviderID) bool {
	return reflect.DeepEqual(p, that)
}
