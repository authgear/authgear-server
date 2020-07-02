package provider

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/oauth"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func loginIDToIdentityInfo(l *loginid.Identity) *identity.Info {
	claims := map[string]interface{}{
		identity.IdentityClaimLoginIDKey:           l.LoginIDKey,
		identity.IdentityClaimLoginIDValue:         l.LoginID,
		identity.IdentityClaimLoginIDOriginalValue: l.OriginalLoginID,
		identity.IdentityClaimLoginIDUniqueKey:     l.UniqueKey,
	}
	for k, v := range l.Claims {
		claims[k] = v
	}

	return &identity.Info{
		Type:     authn.IdentityTypeLoginID,
		ID:       l.ID,
		Claims:   claims,
		Identity: l,
	}
}

func loginIDFromIdentityInfo(userID string, i *identity.Info) *loginid.Identity {
	l := &loginid.Identity{
		ID:     i.ID,
		UserID: userID,
		Claims: map[string]string{},
	}
	for k, v := range i.Claims {
		switch k {
		case identity.IdentityClaimLoginIDKey:
			l.LoginIDKey = v.(string)
		case identity.IdentityClaimLoginIDValue:
			l.LoginID = v.(string)
		case identity.IdentityClaimLoginIDOriginalValue:
			l.OriginalLoginID = v.(string)
		case identity.IdentityClaimLoginIDUniqueKey:
			l.UniqueKey = v.(string)
		default:
			l.Claims[k] = v.(string)
		}
	}
	return l
}

func oauthFromIdentityInfo(userID string, i *identity.Info) *oauth.Identity {
	o := &oauth.Identity{
		ID:     i.ID,
		UserID: userID,
		Claims: map[string]interface{}{},
	}
	for k, v := range i.Claims {
		switch k {
		case identity.IdentityClaimOAuthProviderKeys:
			o.ProviderID.Keys = map[string]interface{}{}
			for k, v := range v.(map[string]interface{}) {
				if k == "type" {
					o.ProviderID.Type = v.(string)
				} else {
					o.ProviderID.Keys[k] = v.(string)
				}
			}
		case identity.IdentityClaimOAuthSubjectID:
			o.ProviderSubjectID = v.(string)
		case identity.IdentityClaimOAuthProfile:
			o.UserProfile = v.(map[string]interface{})
		default:
			o.Claims[k] = v
		}
	}
	return o
}

func anonymousToIdentityInfo(a *anonymous.Identity) *identity.Info {
	claims := map[string]interface{}{
		identity.IdentityClaimAnonymousKeyID: a.KeyID,
		identity.IdentityClaimAnonymousKey:   string(a.Key),
	}

	return &identity.Info{
		Type:     authn.IdentityTypeAnonymous,
		ID:       a.ID,
		Claims:   claims,
		Identity: a,
	}
}

func anonymousFromIdentityInfo(userID string, i *identity.Info) *anonymous.Identity {
	a := &anonymous.Identity{
		ID:     i.ID,
		UserID: userID,
	}
	for k, v := range i.Claims {
		switch k {
		case identity.IdentityClaimAnonymousKeyID:
			a.KeyID = v.(string)
		case identity.IdentityClaimAnonymousKey:
			a.Key = []byte(v.(string))
		}
	}
	return a
}
