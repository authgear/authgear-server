package audit

import (
	"encoding/json"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
)

type Store struct {
	SQLBuilder  *auditdb.SQLBuilder
	SQLExecutor *auditdb.SQLExecutor
}

func (s *Store) PersistEvent(e *event.Event) (err error) {
	data, err := json.Marshal(e)
	if err != nil {
		return
	}

	builder := s.SQLBuilder.Tenant().
		Insert(s.SQLBuilder.TableName("_audit_log")).
		Columns(
			"id",
			"created_at",
			"user_id",
			"activity_type",
			"ip_address",
			"user_agent",
			"client_id",
			"data",
		).
		Values(
			e.ID,
			time.Unix(e.Context.Timestamp, 0).UTC(),
			e.Context.UserID,
			string(e.Type),
			e.Context.IPAddress,
			e.Context.UserAgent,
			e.Context.ClientID,
			data,
		)

	_, err = s.SQLExecutor.ExecWith(builder)
	if err != nil {
		return
	}

	return nil
}
