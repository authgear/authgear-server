package oddb

// Conn encapsulates the interface of an Ourd database connection.
type Conn interface {
	Container(containerKey string) (Container, error)
	Close() error
}

// Driver opens an connection to the underlying database.
type Driver interface {
	Open(name string) (Conn, error)
}

// The DriverFunc type is an adapter such that an ordinary function
// can be used as a Driver.
type DriverFunc func(name string) (Conn, error)

// Open returns a Conn by calling the DriverFunc itself.
func (f DriverFunc) Open(name string) (Conn, error) {
	return f(name)
}
