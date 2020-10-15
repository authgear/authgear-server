package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type DomainService interface {
	ListDomains(appID string) ([]*model.Domain, error)
	CreateDomain(appID string, domain string, isVerified bool, isCustom bool) (*model.Domain, error)
	DeleteDomain(appID string, id string) error
	VerifyDomain(appID string, id string) (*model.Domain, error)
}

type DomainLoader struct {
	Domains DomainService
	Authz   AuthzService
}

func (l *DomainLoader) ListDomains(appID string) *graphqlutil.Lazy {
	err := l.Authz.CheckAccessOfViewer(appID)
	if err != nil {
		return graphqlutil.NewLazyError(err)
	}

	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Domains.ListDomains(appID)
	})
}

func (l *DomainLoader) CreateDomain(appID string, domain string) *graphqlutil.Lazy {
	err := l.Authz.CheckAccessOfViewer(appID)
	if err != nil {
		return graphqlutil.NewLazyError(err)
	}

	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Domains.CreateDomain(appID, domain, false, true)
	})
}

func (l *DomainLoader) DeleteDomain(appID string, id string) *graphqlutil.Lazy {
	err := l.Authz.CheckAccessOfViewer(appID)
	if err != nil {
		return graphqlutil.NewLazyError(err)
	}

	return graphqlutil.NewLazy(func() (interface{}, error) {
		return nil, l.Domains.DeleteDomain(appID, id)
	})
}

func (l *DomainLoader) VerifyDomain(appID string, id string) *graphqlutil.Lazy {
	err := l.Authz.CheckAccessOfViewer(appID)
	if err != nil {
		return graphqlutil.NewLazyError(err)
	}

	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Domains.VerifyDomain(appID, id)
	})
}
