package analytic

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
)

type AuditDBStore struct {
	SQLBuilder  *auditdb.SQLBuilder
	SQLExecutor *auditdb.ReadSQLExecutor
}
