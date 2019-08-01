package hook

import "github.com/skygeario/skygear-server/pkg/auth/event"

type mockStore struct {
	nextSequenceNumber int64
	persistedEvents    []*event.Event
}

func newMockStore() *mockStore {
	return &mockStore{
		nextSequenceNumber: 1,
	}
}

func (store *mockStore) NextSequenceNumber() (seq int64, err error) {
	seq = store.nextSequenceNumber
	store.nextSequenceNumber++
	err = nil
	return
}

func (store *mockStore) AddEvents(events []*event.Event) error {
	store.persistedEvents = append(store.persistedEvents, events...)
	return nil
}

func (store *mockStore) GetEventsForDelivery() ([]*event.Event, error) {
	return store.persistedEvents, nil
}

var _ Store = &mockStore{}
