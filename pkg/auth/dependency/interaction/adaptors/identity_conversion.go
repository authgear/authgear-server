package adaptors

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
)

func oauthToIdentityInfo(o *oauth.Identity) *interaction.IdentityInfo {
	provider := map[string]interface{}{
		"type": o.ProviderID.Type,
	}
	for k, v := range o.ProviderID.Keys {
		provider[k] = v
	}

	claims := map[string]interface{}{
		interaction.IdentityClaimOAuthProvider:  provider,
		interaction.IdentityClaimOAuthSubjectID: o.ProviderSubjectID,
		interaction.IdentityClaimOAuthProfile:   o.UserProfile,
	}
	for k, v := range o.Claims {
		claims[k] = v
	}

	return &interaction.IdentityInfo{
		Type:     interaction.IdentityTypeOAuth,
		ID:       o.ID,
		Claims:   claims,
		Identity: o,
	}
}
