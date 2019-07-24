package hook

type MockStore struct {
	nextSequenceNumber int64
}

func NewMockStore() *MockStore {
	return &MockStore{
		nextSequenceNumber: 1,
	}
}

func (store *MockStore) NextSequenceNumber() (seq int64, err error) {
	seq = store.nextSequenceNumber
	store.nextSequenceNumber++
	err = nil
	return
}

var _ Store = &MockStore{}
