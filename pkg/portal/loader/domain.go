package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type DomainService interface {
	ListDomains(appID string) ([]*model.Domain, error)
}

type DomainLoader struct {
	Domains DomainService
}

func (l *DomainLoader) ListDomains(appID string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		return l.Domains.ListDomains(appID)
	})
}
