package loader

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type DomainLoaderDomainService interface {
	GetMany(ids []string) ([]*model.Domain, error)
}

type DomainLoader struct {
	*graphqlutil.DataLoader `wire:"-"`
	DomainService           DomainLoaderDomainService
	Authz                   AuthzService
}

func NewDomainLoader(domainService DomainLoaderDomainService, authz AuthzService) *DomainLoader {
	l := &DomainLoader{
		DomainService: domainService,
		Authz:         authz,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *DomainLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	domains, err := l.DomainService.GetMany(ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*model.Domain)
	for _, domain := range domains {
		entityMap[domain.ID] = domain
	}

	// Ensure output is in correct order.
	out := make([]interface{}, len(keys))
	for i, id := range ids {
		entity := entityMap[id]
		_, err := l.Authz.CheckAccessOfViewer(entity.AppID)
		if err != nil {
			out[i] = nil
		} else {
			out[i] = entity
		}
	}
	return out, nil
}
