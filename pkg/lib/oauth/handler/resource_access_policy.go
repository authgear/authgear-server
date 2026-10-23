package handler

//go:generate go tool mockgen -source=resource_access_policy.go -destination=resource_access_policy_mock_test.go -package handler_test

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

// ResourceAccessPolicyService is the access-policy read path, shared by
// AuthorizationHandler (/oauth2/authorize) and TokenHandler (/oauth2/token)
// -- unlike TokenHandlerClientResourceScopeService, which checks an
// explicit M2M client-resource association. Both methods already filter
// by client.
type ResourceAccessPolicyService interface {
	GetResourceByURI(ctx context.Context, uri string, client model.ClientCategoryClassifier) (*resourcescope.Resource, error)
	ListScopesByResourceID(ctx context.Context, resourceID string, client model.ClientCategoryClassifier) ([]*resourcescope.Scope, error)
	ListAllAndAllowedScopesByResourceID(ctx context.Context, resourceID string, client model.ClientCategoryClassifier) (all []*resourcescope.Scope, allowed []*resourcescope.Scope, err error)
}

// errResourceNotAvailable covers the Resource not existing, being deleted,
// or its access policy not admitting this client -- distinguishing them
// would enumerate a project's Resources to the client.
func errResourceNotAvailable() error {
	return protocol.NewError("invalid_target", "resource not found or not accessible to this client")
}

// resourceScopes returns every scope of resourceURI and the subset its
// access policy currently opens to client, or errResourceNotAvailable if
// the Resource itself does not. This is the spec's two-level check read as
// one: the Resource level is an error, the Scope level is a filter. See
// docs/specs/api-resource.md § Two-level check.
func resourceScopes(
	ctx context.Context,
	svc ResourceAccessPolicyService,
	client *config.OAuthClientConfig,
	resourceURI string,
) (resource *resourcescope.Resource, known []string, allowed []string, err error) {
	resource, err = svc.GetResourceByURI(ctx, resourceURI, client)
	if err != nil {
		if errors.Is(err, resourcescope.ErrResourceNotFound) {
			return nil, nil, nil, errResourceNotAvailable()
		}
		return nil, nil, nil, err
	}

	all, allowedScopes, err := svc.ListAllAndAllowedScopesByResourceID(ctx, resource.ID, client)
	if err != nil {
		return nil, nil, nil, err
	}
	toNames := func(s *resourcescope.Scope) string { return s.Scope }
	return resource, slice.Map(all, toNames), slice.Map(allowedScopes, toNames), nil
}

// resourceScopesForIssuance re-reads the access policy and returns granted
// minus any resource-specific scope the Resource no longer opens to the
// client's category. Only the returned value is narrowed -- the caller's
// own grant is left alone, so a re-enabled key is restored on the next
// issuance without the user re-authorizing. A token not bound to a Resource
// carries no resource-specific scope at all. See
// docs/specs/api-resource.md § Revocation.
func resourceScopesForIssuance(
	ctx context.Context,
	svc ResourceAccessPolicyService,
	client *config.OAuthClientConfig,
	resourceURI string,
	granted []string,
) ([]string, error) {
	if resourceURI == "" {
		return slice.Filter(granted, func(s string) bool { return !oauth.IsResourceScope(s) }), nil
	}
	_, _, stillAllowed, err := resourceScopes(ctx, svc, client, resourceURI)
	if err != nil {
		return nil, err
	}
	return slice.Filter(granted, func(s string) bool {
		return !oauth.IsResourceScope(s) || slices.Contains(stillAllowed, s)
	}), nil
}

// newGrantScopes checks resource and scope on a token request that creates
// a grant, as /oauth2/authorize would. defaultScopes is used when scope is
// omitted. See docs/specs/access-token-audience-binding.md § New grant.
func newGrantScopes(
	ctx context.Context,
	svc ResourceAccessPolicyService,
	issuer string,
	client *config.OAuthClientConfig,
	resourceURI string,
	requestedScopes []string,
	defaultScopes []string,
) (resourceID string, scopes []string, err error) {
	var known, allowed []string
	if resourceURI != "" {
		if strings.HasPrefix(resourceURI, issuer) {
			return "", nil, protocol.NewError("invalid_target", "resource URI must not be a prefixed by authgear endpoint")
		}
		var resource *resourcescope.Resource
		resource, known, allowed, err = resourceScopes(ctx, svc, client, resourceURI)
		if err != nil {
			return "", nil, err
		}
		resourceID = resource.ID
	}

	if len(requestedScopes) == 0 {
		return resourceID, defaultScopes, nil
	}
	if err := oauth.ValidateScopesByClientConfig(requestedScopes, known); err != nil {
		return "", nil, err
	}
	return resourceID, oauth.GrantedScopes(client, requestedScopes, allowed), nil
}

// checkOfflineFullAccessScopes rejects a biometric or anonymous grant that
// would issue no refresh token or no full-access.
func checkOfflineFullAccessScopes(scopes []string) error {
	if !oauth.ContainsAllScopes(scopes, []string{oauth.OfflineAccess, oauth.FullAccessScope}) {
		return protocol.NewError("invalid_scope", "offline_access and full-access must be requested")
	}
	return nil
}

// derivedGrantResourceScopes returns the Resource Scopes a derived grant
// carries for resourceURI. originals maps each Resource URI an original
// refresh token is bound to ("" for unbound) to its scopes. See
// docs/specs/access-token-audience-binding.md § Derived grant.
func derivedGrantResourceScopes(
	ctx context.Context,
	svc ResourceAccessPolicyService,
	originalClient *config.OAuthClientConfig,
	newClient *config.OAuthClientConfig,
	originals map[string][]string,
	resourceURI string,
) (resourceID string, scopes []string, err error) {
	if resourceURI == "" {
		return "", nil, nil
	}

	held, ok := originals[resourceURI]
	if !ok {
		return "", nil, protocol.NewError("invalid_target", "resource is not bound to the original session")
	}

	_, _, allowedOriginal, err := resourceScopes(ctx, svc, originalClient, resourceURI)
	if err != nil {
		return "", nil, err
	}
	resource, _, allowedNew, err := resourceScopes(ctx, svc, newClient, resourceURI)
	if err != nil {
		return "", nil, err
	}

	scopes = slice.Filter(held, func(s string) bool {
		return oauth.IsResourceScope(s) && slices.Contains(allowedOriginal, s) && slices.Contains(allowedNew, s)
	})
	return resource.ID, scopes, nil
}

// defaultOfflineFullAccessScopes is the scope of a biometric or anonymous
// grant that does not request one.
func defaultOfflineFullAccessScopes() []string {
	return []string{oauth.ScopeOpenID, oauth.OfflineAccess, oauth.FullAccessScope}
}
