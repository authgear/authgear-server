package resourcescope

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

// AccessPolicyService is the access-policy read path, shared by
// AuthorizationHandler (/oauth2/authorize) and TokenHandler (/oauth2/token)
// -- unlike ClientResourceScopeService, which checks an explicit m2m
// client-resource association. Both methods here filter by client.
type AccessPolicyService struct {
	Store *Store
}

// GetResourceByURI returns the Resource at uri if its access policy admits
// client, or ErrResourceNotFound otherwise -- folding "not accessible" into
// "not found" so callers don't leak which case a client hit.
func (s *AccessPolicyService) GetResourceByURI(ctx context.Context, uri string, client model.ClientCategoryClassifier) (*Resource, error) {
	resource, err := s.Store.GetResourceByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	if !resource.AccessPolicy.AllowsClient(client) {
		return nil, ErrResourceNotFound
	}
	return resource, nil
}

// ListScopesByResourceID returns resourceID's scopes that its access policy
// currently opens to client.
func (s *AccessPolicyService) ListScopesByResourceID(ctx context.Context, resourceID string, client model.ClientCategoryClassifier) ([]*Scope, error) {
	scopes, err := s.Store.ListScopesByResourceID(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	var allowed []*Scope
	for _, sc := range scopes {
		if sc.AccessPolicy.AllowsClient(client) {
			allowed = append(allowed, sc)
		}
	}
	return allowed, nil
}
