package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type DomainService interface {
	ListDomains(appID string) ([]*model.Domain, error)
	CreateDomain(appID string, domain string) (*model.Domain, error)
	DeleteDomain(appID string, id string) error
	VerifyDomain(appID string, id string) (*model.Domain, error)
}

type DomainLoader struct {
	Domains DomainService
}

func (l *DomainLoader) ListDomains(appID string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Domains.ListDomains(appID)
	})
}

func (l *DomainLoader) CreateDomain(appID string, domain string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Domains.CreateDomain(appID, domain)
	})
}

func (l *DomainLoader) DeleteDomain(appID string, id string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		return nil, l.Domains.DeleteDomain(appID, id)
	})
}

func (l *DomainLoader) VerifyDomain(appID string, id string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Domains.VerifyDomain(appID, id)
	})
}
