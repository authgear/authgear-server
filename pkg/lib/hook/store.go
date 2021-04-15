package hook

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
)

type Store struct {
	SQLBuilder  db.SQLBuilder
	SQLExecutor *tenantdb.SQLExecutor

	events []*event.Event `wire:"-"`
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

func (store *Store) AddEvents(events []*event.Event) error {
	// TODO(webhook): persist events
	store.events = append(store.events, events...)
	return nil
}

func (store *Store) GetEventsForDelivery() ([]*event.Event, error) {
	// TODO(webhook): get events
	return store.events, nil
}
