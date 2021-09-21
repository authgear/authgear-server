package oauth

import "time"

type Authorization struct {
	ID        string
	AppID     string
	ClientID  string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Scopes    []string
}

func (z Authorization) IsAuthorized(scopes []string) bool {
	scopeMap := map[string]struct{}{}
	for _, s := range z.Scopes {
		scopeMap[s] = struct{}{}
	}
	for _, s := range scopes {
		if _, ok := scopeMap[s]; !ok {
			return false
		}
	}
	return true
}

func (z Authorization) WithScopesAdded(scopes []string) *Authorization {
	seen := map[string]struct{}{}
	var newScopes []string
	for _, s := range append(z.Scopes, scopes...) {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			newScopes = append(newScopes, s)
		}
	}
	z.Scopes = newScopes
	return &z
}
