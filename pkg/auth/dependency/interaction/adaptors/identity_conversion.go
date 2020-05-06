package adaptors

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func loginIDToIdentityInfo(l *loginid.Identity) *interaction.IdentityInfo {
	claims := map[string]interface{}{
		interaction.IdentityClaimLoginIDKey:           l.LoginIDKey,
		interaction.IdentityClaimLoginIDValue:         l.LoginID,
		interaction.IdentityClaimLoginIDOriginalValue: l.OriginalLoginID,
		interaction.IdentityClaimLoginIDUniqueKey:     l.UniqueKey,
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

func loginIDFromIdentityInfo(userID string, i *interaction.IdentityInfo) *loginid.Identity {
	l := &loginid.Identity{
		ID:     i.ID,
		UserID: userID,
		Claims: map[string]string{},
	}
	for k, v := range i.Claims {
		switch k {
		case interaction.IdentityClaimLoginIDKey:
			l.LoginIDKey = v.(string)
		case interaction.IdentityClaimLoginIDValue:
			l.LoginID = v.(string)
		case interaction.IdentityClaimLoginIDOriginalValue:
			l.OriginalLoginID = v.(string)
		case interaction.IdentityClaimLoginIDUniqueKey:
			l.UniqueKey = v.(string)
		default:
			l.Claims[k] = v.(string)
		}
	}
	return l
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

func oauthFromIdentityInfo(userID string, i *interaction.IdentityInfo) *oauth.Identity {
	o := &oauth.Identity{
		ID:     i.ID,
		UserID: userID,
		Claims: map[string]interface{}{},
	}
	for k, v := range i.Claims {
		switch k {
		case interaction.IdentityClaimOAuthProvider:
			o.ProviderID.Keys = map[string]interface{}{}
			for k, v := range v.(map[string]interface{}) {
				if k == "type" {
					o.ProviderID.Type = v.(string)
				} else {
					o.ProviderID.Keys[k] = v
				}
			}
		case interaction.IdentityClaimOAuthSubjectID:
			o.ProviderSubjectID = v.(string)
		case interaction.IdentityClaimOAuthProfile:
			o.UserProfile = v.(map[string]interface{})
		default:
			o.Claims[k] = v
		}
	}
	return o
}

func anonymousToIdentityInfo(a *anonymous.Identity) *interaction.IdentityInfo {
	claims := map[string]interface{}{
		interaction.IdentityClaimAnonymousKeyID: a.KeyID,
		interaction.IdentityClaimAnonymousKey:   string(a.Key),
	}

	return &interaction.IdentityInfo{
		Type:     authn.IdentityTypeAnonymous,
		ID:       a.ID,
		Claims:   claims,
		Identity: a,
	}
}

func anonymousFromIdentityInfo(userID string, i *interaction.IdentityInfo) *anonymous.Identity {
	a := &anonymous.Identity{
		ID:     i.ID,
		UserID: userID,
	}
	for k, v := range i.Claims {
		switch k {
		case interaction.IdentityClaimAnonymousKeyID:
			a.KeyID = v.(string)
		case interaction.IdentityClaimAnonymousKey:
			a.Key = []byte(v.(string))
		}
	}
	return a
}
