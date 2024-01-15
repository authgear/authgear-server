package loader

import (
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AuditLogQuery interface {
	GetByIDs(ids []string) ([]*audit.Log, error)
}

type AuditLogLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	AuditDatabase *auditdb.ReadHandle
	Query         AuditLogQuery
}

func NewAuditLogLoader(query AuditLogQuery, handle *auditdb.ReadHandle) *AuditLogLoader {
	l := &AuditLogLoader{
		Query:         query,
		AuditDatabase: handle,
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

	var entities []*audit.Log
	// Get entities.
	err := l.AuditDatabase.ReadOnly(func() (err error) {
		entities, err = l.Query.GetByIDs(ids)
		return
	})
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
		out[i] = entity
	}

	return out, nil
}
