package handler

//go:generate go tool mockgen -source=resource_access_policy.go -destination=resource_access_policy_mock_test.go -package handler_test

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/resourcescope"
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
