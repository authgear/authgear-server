package handler

//go:generate go tool mockgen -source=resource_access_policy.go -destination=resource_access_policy_mock_test.go -package handler_test

import (
	"context"
	"errors"
	"slices"

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
}

// errResourceNotAvailable covers the Resource not existing, being deleted,
// or its access policy not admitting this client -- distinguishing them
// would enumerate a project's Resources to the client.
func errResourceNotAvailable() error {
	return protocol.NewError("invalid_target", "resource not found or not accessible to this client")
}

// allowedResourceScopes returns the scopes of resourceURI that its access
// policy currently opens to client, or errResourceNotAvailable if the
// Resource itself does not. This is the spec's two-level check read as one:
// the Resource level is an error, the Scope level is a filter. See
// docs/specs/api-resource.md § Two-level check.
func allowedResourceScopes(
	ctx context.Context,
	svc ResourceAccessPolicyService,
	client *config.OAuthClientConfig,
	resourceURI string,
) ([]string, error) {
	resource, err := svc.GetResourceByURI(ctx, resourceURI, client)
	if err != nil {
		if errors.Is(err, resourcescope.ErrResourceNotFound) {
			return nil, errResourceNotAvailable()
		}
		return nil, err
	}

	scopes, err := svc.ListScopesByResourceID(ctx, resource.ID, client)
	if err != nil {
		return nil, err
	}
	allowed := make([]string, len(scopes))
	for i, s := range scopes {
		allowed[i] = s.Scope
	}
	return allowed, nil
}

// resourceScopesForIssuance re-reads the access policy and returns granted
// minus any resource-specific scope the Resource no longer opens to the
// client's category. Only the returned value is narrowed -- the caller's
// own grant is left alone, so a re-enabled key is restored on the next
// issuance without the user re-authorizing. See
// docs/specs/api-resource.md § Revocation.
func (h *TokenHandler) resourceScopesForIssuance(
	ctx context.Context,
	client *config.OAuthClientConfig,
	resourceURI string,
	granted []string,
) ([]string, error) {
	if resourceURI == "" {
		return granted, nil
	}
	stillAllowed, err := allowedResourceScopes(ctx, h.ResourceAccessPolicyService, client, resourceURI)
	if err != nil {
		return nil, err
	}
	return slice.Filter(granted, func(s string) bool {
		return !oauth.IsResourceScope(s) || slices.Contains(stillAllowed, s)
	}), nil
}
