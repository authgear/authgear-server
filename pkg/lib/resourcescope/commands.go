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
	resource := c.Store.NewResource(options)
	err := c.Store.CreateResource(ctx, resource)
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
	resource, err := c.Store.GetResourceByURI(ctx, uri)
	if err != nil {
		return err
	}
	// Delete all client-resource associations
	if err := c.Store.DeleteAllClientResourceAssociations(ctx, resource.ID); err != nil {
		return err
	}
	// Delete all client-scope associations for all scopes of this resource
	if err := c.Store.DeleteAllClientScopeAssociationsByResourceID(ctx, resource.ID); err != nil {
		return err
	}
	// Delete all resource-scopes
	if err := c.Store.DeleteAllResourceScopes(ctx, resource.ID); err != nil {
		return err
	}
	// Delete the resource itself
	return c.Store.DeleteResource(ctx, resource.ID)
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
	// Remove all client-scope associations for all scopes of this resource for this client
	if err := c.Store.DeleteClientScopeAssociationsByResourceID(ctx, clientID, resource.ID); err != nil {
		return err
	}
	// Remove the client-resource association
	return c.Store.RemoveResourceFromClientID(ctx, resource.ID, clientID)
}

func (c *Commands) CreateScope(ctx context.Context, resourceURI string, options *NewScopeOptions) (*model.Scope, error) {
	resource, err := c.Store.GetResourceByURI(ctx, resourceURI)
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
	resource, err := c.Store.GetResourceByURI(ctx, options.ResourceURI)
	if err != nil {
		return nil, err
	}
	err = c.Store.UpdateScope(ctx, options)
	if err != nil {
		return nil, err
	}
	scope, err := c.Store.GetResourceScope(ctx, resource.ID, options.Scope)
	if err != nil {
		return nil, err
	}
	return scope.ToModel(), nil
}

func (c *Commands) DeleteScope(ctx context.Context, resourceURI string, scope string) error {
	resource, err := c.Store.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return err
	}
	s, err := c.Store.GetResourceScope(ctx, resource.ID, scope)
	if err != nil {
		return err
	}
	// Remove all client-scope associations for this scope
	if err := c.Store.DeleteAllClientScopeAssociationsByScopeID(ctx, s.ID); err != nil {
		return err
	}
	return c.Store.DeleteScope(ctx, resourceURI, scope)
}

func (c *Commands) checkResourceAssociatedToClient(ctx context.Context, resourceID, clientID string) error {
	_, err := c.Store.GetClientResource(ctx, clientID, resourceID)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commands) AddScopesToClientID(ctx context.Context, resourceURI, clientID string, scopes []string) ([]*model.Scope, error) {
	if _, found := c.OAuthConfig.GetClient(clientID); !found {
		return nil, ErrClientNotFound
	}
	resource, err := c.Store.GetResourceByURI(ctx, resourceURI)
	if err != nil {
		return nil, err
	}
	if err := c.checkResourceAssociatedToClient(ctx, resource.ID, clientID); err != nil {
		return nil, err
	}
	scopeMap, err := c.Store.GetScopesByResourceIDAndScopes(ctx, resource.ID, scopes)
	if err != nil {
		return nil, err
	}
	scopeIDs := []string{}
	for _, scopeStr := range scopes {
		sc, ok := scopeMap[scopeStr]
		if !ok {
			return nil, ErrScopeNotFound
		}
		scopeIDs = append(scopeIDs, sc.ID)
	}
	if err := c.Store.AddScopesToClientID(ctx, resource.ID, scopeIDs, clientID); err != nil {
		return nil, err
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
	if err := c.checkResourceAssociatedToClient(ctx, resource.ID, clientID); err != nil {
		return nil, err
	}
	scopeMap, err := c.Store.GetScopesByResourceIDAndScopes(ctx, resource.ID, scopes)
	if err != nil {
		return nil, err
	}
	scopeIDs := []string{}
	for _, scopeStr := range scopes {
		sc, ok := scopeMap[scopeStr]
		if !ok {
			return nil, ErrScopeNotFound
		}
		scopeIDs = append(scopeIDs, sc.ID)
	}
	if err := c.Store.RemoveScopesFromClientID(ctx, scopeIDs, clientID); err != nil {
		return nil, err
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
	if err := c.checkResourceAssociatedToClient(ctx, resource.ID, clientID); err != nil {
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
	desiredScopeMap, err := c.Store.GetScopesByResourceIDAndScopes(ctx, resource.ID, scopes)
	if err != nil {
		return nil, err
	}
	// Add missing scopes
	toAddIDs := []string{}
	for _, scopeStr := range scopes {
		if _, ok := currentSet[scopeStr]; !ok {
			sc, ok := desiredScopeMap[scopeStr]
			if !ok {
				return nil, ErrScopeNotFound
			}
			toAddIDs = append(toAddIDs, sc.ID)
		}
	}
	if len(toAddIDs) > 0 {
		if err := c.Store.AddScopesToClientID(ctx, resource.ID, toAddIDs, clientID); err != nil {
			return nil, err
		}
	}
	// Remove extra scopes
	toRemoveIDs := []string{}
	for scopeStr, s := range currentSet {
		if _, ok := desiredSet[scopeStr]; !ok {
			toRemoveIDs = append(toRemoveIDs, s.ID)
		}
	}
	if len(toRemoveIDs) > 0 {
		if err := c.Store.RemoveScopesFromClientID(ctx, toRemoveIDs, clientID); err != nil {
			return nil, err
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
