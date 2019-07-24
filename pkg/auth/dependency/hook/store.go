package hook

type Store interface {
	NextSequenceNumber() (int64, error)
}
