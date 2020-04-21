package adaptors

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func loginIDToIdentityInfo(l *loginid.Identity) *interaction.IdentityInfo {
	claims := map[string]interface{}{
		interaction.IdentityClaimLoginIDKey:       l.LoginIDKey,
		interaction.IdentityClaimLoginIDValue:     l.LoginID,
		interaction.IdentityClaimLoginIDUniqueKey: l.UniqueKey,
	}
	for k, v := range l.Claims {
		claims[k] = v
	}

	return &interaction.IdentityInfo{
		Type:     authn.IdentityTypeLoginID,
		ID:       l.ID,
		Claims:   claims,
		Identity: l,
	}
}

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
		Type:     authn.IdentityTypeOAuth,
		ID:       o.ID,
		Claims:   claims,
		Identity: o,
	}
}
