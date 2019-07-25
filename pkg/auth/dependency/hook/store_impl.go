package hook

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type storeImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
}

func NewStore(builder db.SQLBuilder, executor db.SQLExecutor) Store {
	return &storeImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
	}
}

func (store *storeImpl) NextSequenceNumber() (seq int64, err error) {
	builder := store.sqlBuilder.Select(
		fmt.Sprintf("nextval('%s')", store.sqlBuilder.FullTableName("event_sequence")),
	)
	row := store.sqlExecutor.QueryRowWith(builder)
	err = row.Scan(&seq)
	return
}

func (store *storeImpl) PersistEvents(events []*event.Event) error {
	// TODO(webhook): persist events
	return nil
}
