package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AuditLogQuery interface {
	GetByIDs(ctx context.Context, ids []string) ([]*audit.Log, error)
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

func (l *AuditLogLoader) LoadFunc(ctx context.Context, keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	var entities []*audit.Log
	// Get entities.
	err := l.AuditDatabase.ReadOnly(ctx, func(ctx context.Context) (err error) {
		entities, err = l.Query.GetByIDs(ctx, ids)
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
