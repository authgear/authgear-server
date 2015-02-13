package oddb

// Conn encapsulates the interface of an Ourd connection to a container.
type Conn interface {
	PublicDB() Database
	PrivateDB(userKey string) Database

	Close() error
}
