package loader

import (
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AuditLogQuery interface {
	GetByIDs(ids []string) ([]*audit.Log, error)
}

type AuditLogLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Query AuditLogQuery
}

func NewAuditLogLoader(query AuditLogQuery) *AuditLogLoader {
	l := &AuditLogLoader{
		Query: query,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *AuditLogLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	entities, err := l.Query.GetByIDs(ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*audit.Log)
	for _, entity := range entities {
		entityMap[entity.ID] = entity
	}

	// Ensure output is in correct order.
	out := make([]interface{}, len(keys))
	for i, id := range ids {
		entity := entityMap[id]
		if err != nil {
			out[i] = nil
		} else {
			out[i] = entity
		}
	}

	return out, nil
}
