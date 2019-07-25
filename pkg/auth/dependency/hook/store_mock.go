package hook

type mockStore struct {
	nextSequenceNumber int64
}

func newMockStore() *mockStore {
	return &mockStore{
		nextSequenceNumber: 1,
	}
}

func (store *mockStore) Reset() {
	*store = mockStore{
		nextSequenceNumber: 1,
	}
}

func (store *mockStore) NextSequenceNumber() (seq int64, err error) {
	seq = store.nextSequenceNumber
	store.nextSequenceNumber++
	err = nil
	return
}

var _ Store = &mockStore{}
