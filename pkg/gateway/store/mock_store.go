package store

import (
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

type MockStore struct {
	Domains map[string]model.Domain
}

func NewMockStore() *MockStore {
	return &MockStore{
		Domains: map[string]model.Domain{},
	}
}

func (s *MockStore) GetDomain(domain string) (*model.Domain, error) {
	for _, d := range s.Domains {
		if d.Domain == domain {
			return &d, nil
		}
	}
	return nil, NewNotFoundError("domain")
}

func (s *MockStore) GetDefaultDomain(domain string) (*model.Domain, error) {
	for _, d := range s.Domains {
		if d.Domain == domain && d.Assignment == model.AssignmentTypeDefault {
			return &d, nil
		}
	}
	return nil, NewNotFoundError("domain")
}

func (s *MockStore) GetApp(id string) (*model.App, error) {
	return nil, nil
}

func (s *MockStore) GetLastDeploymentRoutes(app model.App) ([]*model.DeploymentRoute, error) {
	return nil, nil
}

func (s *MockStore) GetLastDeploymentHooks(app model.App) (*model.DeploymentHooks, error) {
	return nil, nil
}

func (s *MockStore) Close() error {
	return nil
}

var _ GatewayStore = &MockStore{}
