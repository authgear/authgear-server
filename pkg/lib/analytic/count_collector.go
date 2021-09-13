package analytic

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type CountCollector struct {
	GlobalHandle  *globaldb.Handle
	GlobalDBStore *GlobalDBStore
	AppDBHandle   *appdb.Handle
	AppDBStore    *AppDBStore
	AuditDBHandle *auditdb.WriteHandle
	AuditDBStore  *AuditDBStore
}

func (c *CountCollector) CollectDaily() error {
	return nil
}
