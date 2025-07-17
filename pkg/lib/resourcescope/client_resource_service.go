package resourcescope

import (
	"context"
)

type ClientResourceScopeService struct {
	Store *Store
}

func (s *ClientResourceScopeService) GetClientResourceByURI(ctx context.Context, clientID string, uri string) (*Resource, error) {
	resource, err := s.Store.GetResourceByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	// Check the association
	_, err = s.Store.GetClientResource(ctx, clientID, resource.ID)
	if err != nil {
		return nil, err
	}
	return resource, nil
}

func (s *ClientResourceScopeService) GetClientResourceScopes(ctx context.Context, clientID string, resourceID string) ([]*Scope, error) {
	scopes, err := s.Store.ListClientScopesByResourceID(ctx, resourceID, clientID)
	if err != nil {
		return nil, err
	}
	return scopes, nil
}
