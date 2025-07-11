package resourcescope

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Commands struct {
	Store *Store

	OAuthConfig *config.OAuthConfig
}

func (c *Commands) CreateResource(ctx context.Context, options *NewResourceOptions) (*model.Resource, error) {
	err := ValidateResourceURI(ctx, options.URI)
	if err != nil {
		return nil, err
	}

	resource := c.Store.NewResource(options)
	err = c.Store.CreateResource(ctx, resource)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (c *Commands) UpdateResource(ctx context.Context, options *UpdateResourceOptions) (*model.Resource, error) {
	err := c.Store.UpdateResource(ctx, options)
	if err != nil {
		return nil, err
	}
	resource, err := c.Store.GetResourceByURI(ctx, options.ResourceURI)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (c *Commands) DeleteResourceByURI(ctx context.Context, uri string) error {
	return c.Store.DeleteResourceByURI(ctx, uri)
}

func (c *Commands) GetResourceByURI(ctx context.Context, uri string) (*model.Resource, error) {
	resource, err := c.Store.GetResourceByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	return resource.ToModel(), nil
}

func (c *Commands) AddResourceToClientID(ctx context.Context, resourceURI, clientID string) error {
	if _, found := c.OAuthConfig.GetClient(clientID); !found {
		return ErrClientNotFound
	}
	resource, err := c.Store.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return err
	}
	return c.Store.AddResourceToClientID(ctx, resource.ID, clientID)
}

func (c *Commands) RemoveResourceFromClientID(ctx context.Context, resourceURI, clientID string) error {
	if _, found := c.OAuthConfig.GetClient(clientID); !found {
		return ErrClientNotFound
	}
	resource, err := c.Store.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return err
	}
	return c.Store.RemoveResourceFromClientID(ctx, resource.ID, clientID)
}

func (c *Commands) CreateScope(ctx context.Context, options *NewScopeOptions) (*model.Scope, error) {
	err := ValidateScope(ctx, options.Scope)
	if err != nil {
		return nil, err
	}
	resource, err := c.Store.GetResourceByURI(ctx, options.ResourceURI)
	if err != nil {
		return nil, err
	}
	scope := c.Store.NewScope(resource, options)
	err = c.Store.CreateScope(ctx, scope)
	if err != nil {
		return nil, err
	}
	return scope.ToModel(), nil
}

func (c *Commands) UpdateScope(ctx context.Context, options *UpdateScopeOptions) (*model.Scope, error) {
	err := c.Store.UpdateScope(ctx, options)
	if err != nil {
		return nil, err
	}
	scope, err := c.Store.GetScope(ctx, options.ResourceURI, options.Scope)
	if err != nil {
		return nil, err
	}
	return scope.ToModel(), nil
}

func (c *Commands) DeleteScope(ctx context.Context, resourceURI string, scope string) error {
	return c.Store.DeleteScope(ctx, resourceURI, scope)
}

func (c *Commands) AddScopesToClientID(ctx context.Context, resourceURI, clientID string, scopes []string) ([]*model.Scope, error) {
	if _, found := c.OAuthConfig.GetClient(clientID); !found {
		return nil, ErrClientNotFound
	}
	resource, err := c.Store.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return nil, err
	}
	for _, scopeStr := range scopes {
		scope, err := c.Store.GetScopeByResourceIDAndScope(ctx, resource.ID, scopeStr)
		if err != nil {
			return nil, err
		}
		err = c.Store.AddScopeToClientID(ctx, resource.ID, scope.ID, clientID)
		if err != nil {
			return nil, err
		}
	}
	finalScopes, err := c.Store.ListClientScopesByResourceID(ctx, resource.ID, clientID)
	if err != nil {
		return nil, err
	}
	var result []*model.Scope
	for _, s := range finalScopes {
		result = append(result, s.ToModel())
	}
	return result, nil
}

func (c *Commands) RemoveScopesFromClientID(ctx context.Context, resourceURI, clientID string, scopes []string) ([]*model.Scope, error) {
	if _, found := c.OAuthConfig.GetClient(clientID); !found {
		return nil, ErrClientNotFound
	}
	resource, err := c.Store.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return nil, err
	}
	for _, scopeStr := range scopes {
		scope, err := c.Store.GetScopeByResourceIDAndScope(ctx, resource.ID, scopeStr)
		if err != nil {
			return nil, err
		}
		err = c.Store.RemoveScopeFromClientID(ctx, scope.ID, clientID)
		if err != nil {
			return nil, err
		}
	}
	finalScopes, err := c.Store.ListClientScopesByResourceID(ctx, resource.ID, clientID)
	if err != nil {
		return nil, err
	}
	var result []*model.Scope
	for _, s := range finalScopes {
		result = append(result, s.ToModel())
	}
	return result, nil
}

func (c *Commands) ReplaceScopesOfClientID(ctx context.Context, resourceURI, clientID string, scopes []string) ([]*model.Scope, error) {
	if _, found := c.OAuthConfig.GetClient(clientID); !found {
		return nil, ErrClientNotFound
	}
	resource, err := c.Store.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return nil, err
	}
	// Get all current scopes for this resource and client
	currentScopes, err := c.Store.ListClientScopesByResourceID(ctx, resource.ID, clientID)
	if err != nil {
		return nil, err
	}
	currentSet := make(map[string]*Scope)
	for _, s := range currentScopes {
		s := s
		currentSet[s.Scope] = s
	}
	// Build desired set
	desiredSet := make(map[string]struct{})
	for _, scopeStr := range scopes {
		desiredSet[scopeStr] = struct{}{}
	}
	// Add missing scopes
	for _, scopeStr := range scopes {
		if _, ok := currentSet[scopeStr]; !ok {
			scope, err := c.Store.GetScopeByResourceIDAndScope(ctx, resource.ID, scopeStr)
			if err != nil {
				return nil, err
			}
			err = c.Store.AddScopeToClientID(ctx, resource.ID, scope.ID, clientID)
			if err != nil {
				return nil, err
			}
		}
	}
	// Remove extra scopes
	for scopeStr, s := range currentSet {
		if _, ok := desiredSet[scopeStr]; !ok {
			err := c.Store.RemoveScopeFromClientID(ctx, s.ID, clientID)
			if err != nil {
				return nil, err
			}
		}
	}
	// Return the final set
	finalScopes, err := c.Store.ListClientScopesByResourceID(ctx, resource.ID, clientID)
	if err != nil {
		return nil, err
	}
	var result []*model.Scope
	for _, s := range finalScopes {
		result = append(result, s.ToModel())
	}
	return result, nil
}
