package event

import (
	"fmt"

	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
)

type StoreImpl struct {
	SQLBuilder  *tenantdb.SQLBuilder
	SQLExecutor *tenantdb.SQLExecutor
}

func (s *StoreImpl) NextSequenceNumber() (seq int64, err error) {
	builder := s.SQLBuilder.Global().
		Select(fmt.Sprintf("nextval('%s')", s.SQLBuilder.TableName("_auth_event_sequence")))
	row, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return
	}
	err = row.Scan(&seq)
	return
}
