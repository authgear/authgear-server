package hook

import (
	"fmt"

	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
)

type Store struct {
	SQLBuilder  *tenantdb.SQLBuilder
	SQLExecutor *tenantdb.SQLExecutor
}

func (store *Store) NextSequenceNumber() (seq int64, err error) {
	builder := store.SQLBuilder.Global().
		Select(fmt.Sprintf("nextval('%s')", store.SQLBuilder.TableName("_auth_event_sequence")))
	row, err := store.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return
	}
	err = row.Scan(&seq)
	return
}
