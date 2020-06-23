package hook

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/db"
)

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor db.SQLExecutor

	events []*event.Event `wire:"-"`
}

func (store *Store) NextSequenceNumber() (seq int64, err error) {
	builder := store.SQLBuilder.Global().
		Select(fmt.Sprintf("nextval('%s')", store.SQLBuilder.FullTableName("event_sequence")))
	row, err := store.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return
	}
	err = row.Scan(&seq)
	return
}

func (store *Store) AddEvents(events []*event.Event) error {
	// TODO(webhook): persist events
	store.events = append(store.events, events...)
	return nil
}

func (store *Store) GetEventsForDelivery() ([]*event.Event, error) {
	// TODO(webhook): get events
	return store.events, nil
}
