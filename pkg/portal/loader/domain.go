package loader

import (
	"context"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type DomainLoaderDomainService interface {
	GetMany(ctx context.Context, ids []string) ([]*apimodel.Domain, error)
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

func (l *DomainLoader) LoadFunc(ctx context.Context, keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	domains, err := l.DomainService.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*apimodel.Domain)
	for _, domain := range domains {
		entityMap[domain.ID] = domain
	}

	// Ensure output is in correct order.
	out := make([]interface{}, len(keys))
	for i, id := range ids {
		entity := entityMap[id]
		_, err := l.Authz.CheckAccessOfViewer(ctx, entity.AppID)
		if err != nil {
			out[i] = nil
		} else {
			out[i] = entity
		}
	}
	return out, nil
}
