package oauth

import (
	"slices"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// AuthorizationProjectScopesKey is the ScopesByResources key of the
// project-level scopes. Resource IDs are UUIDs, so it cannot collide.
const AuthorizationProjectScopesKey = "authgear"

type Authorization struct {
	ID        string
	AppID     string
	ClientID  string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	// ScopesByResources maps a resource ID, or AuthorizationProjectScopesKey,
	// to the scopes granted on it.
	ScopesByResources map[string][]string
}

// groupScopes keys each scope by where it is granted: project-level scopes
// under AuthorizationProjectScopesKey, the rest under resourceID. With no
// resourceID, resource scopes cannot be attributed, so they are left out.
func groupScopes(resourceID string, scopes []string) map[string][]string {
	grouped := map[string][]string{}
	for _, s := range scopes {
		key := AuthorizationProjectScopesKey
		if IsResourceScope(s) {
			if resourceID == "" {
				continue
			}
			key = resourceID
		}
		grouped[key] = append(grouped[key], s)
	}
	return grouped
}

// IsAuthorized reports whether scopes are all granted, with resource scopes
// granted on resourceID. resourceID is empty when no resource is requested,
// and then resource scopes are ignored.
func (z Authorization) IsAuthorized(resourceID string, scopes []string) bool {
	grouped := groupScopes(resourceID, scopes)
	for key, ss := range grouped {
		granted := z.ScopesByResources[key]
		for _, s := range ss {
			if !slices.Contains(granted, s) {
				return false
			}
		}
	}
	return true
}

// WithScopesAdded grants scopes, with resource scopes granted on resourceID.
// resourceID is empty when no resource is requested, and then resource scopes
// are ignored.
func (z Authorization) WithScopesAdded(resourceID string, scopes []string) *Authorization {
	grouped := groupScopes(resourceID, scopes)
	newScopesByResources := make(map[string][]string, len(z.ScopesByResources)+len(grouped))
	for key, ss := range z.ScopesByResources {
		newScopesByResources[key] = slices.Clone(ss)
	}
	for key, ss := range grouped {
		for _, s := range ss {
			if !slices.Contains(newScopesByResources[key], s) {
				newScopesByResources[key] = append(newScopesByResources[key], s)
			}
		}
	}
	z.ScopesByResources = newScopesByResources
	return &z
}

// ResourceIDs returns the IDs of the resources that have granted scopes, sorted.
func (z Authorization) ResourceIDs() []string {
	var ids []string
	for key := range z.ScopesByResources {
		if key != AuthorizationProjectScopesKey {
			ids = append(ids, key)
		}
	}
	slices.Sort(ids)
	return ids
}

func (z Authorization) WithResourcesRemoved(resourceIDs []string) *Authorization {
	newScopesByResources := make(map[string][]string, len(z.ScopesByResources))
	for key, ss := range z.ScopesByResources {
		if !slices.Contains(resourceIDs, key) {
			newScopesByResources[key] = ss
		}
	}
	z.ScopesByResources = newScopesByResources
	return &z
}

func (z Authorization) ProjectScopes() []string {
	return z.ScopesByResources[AuthorizationProjectScopesKey]
}

// AllScopes returns every granted scope name, project-level scopes first.
// A name granted on more than one resource appears once.
func (z Authorization) AllScopes() []string {
	var all []string
	add := func(ss []string) {
		for _, s := range ss {
			if !slices.Contains(all, s) {
				all = append(all, s)
			}
		}
	}
	add(z.ProjectScopes())
	for _, id := range z.ResourceIDs() {
		add(z.ScopesByResources[id])
	}
	return all
}

func (z Authorization) ToAPIModel() *model.Authorization {
	authorizedScopes := []model.AuthorizedScope{}
	for _, s := range z.ProjectScopes() {
		authorizedScopes = append(authorizedScopes, model.AuthorizedScope{Scope: s})
	}
	for _, id := range z.ResourceIDs() {
		for _, s := range z.ScopesByResources[id] {
			authorizedScopes = append(authorizedScopes, model.AuthorizedScope{ResourceID: id, Scope: s})
		}
	}

	scopes := z.ProjectScopes()
	if scopes == nil {
		scopes = []string{}
	}

	return &model.Authorization{
		Meta: model.Meta{
			ID:        z.ID,
			CreatedAt: z.CreatedAt,
			UpdatedAt: z.UpdatedAt,
		},
		ClientID:         z.ClientID,
		Scopes:           scopes,
		AuthorizedScopes: authorizedScopes,
	}
}
