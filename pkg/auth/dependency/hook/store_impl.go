package hook

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/db"
)

type storeImpl struct {
	sqlBuilder  db.SQLBuilder
	sqlExecutor db.SQLExecutor
	events      []*event.Event
}

func NewStore(builder db.SQLBuilder, executor db.SQLExecutor) Store {
	return &storeImpl{
		sqlBuilder:  builder,
		sqlExecutor: executor,
	}
}

func (store *storeImpl) NextSequenceNumber() (seq int64, err error) {
	builder := store.sqlBuilder.Global().
		Select(fmt.Sprintf("nextval('%s')", store.sqlBuilder.FullTableName("event_sequence")))
	row, err := store.sqlExecutor.QueryRowWith(builder)
	if err != nil {
		return
	}
	err = row.Scan(&seq)
	return
}

func (store *storeImpl) AddEvents(events []*event.Event) error {
	// TODO(webhook): persist events
	store.events = append(store.events, events...)
	return nil
}

func (store *storeImpl) GetEventsForDelivery() ([]*event.Event, error) {
	// TODO(webhook): get events
	return store.events, nil
}
