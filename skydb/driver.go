package skydb

// Driver opens an connection to the underlying database.
type Driver interface {
	Open(appName string, accessModel AccessModel, optionString string, migrate bool) (Conn, error)
}

// The DriverFunc type is an adapter such that an ordinary function
// can be used as a Driver.
type DriverFunc func(appName string, accessModel AccessModel, optionString string, migrate bool) (Conn, error)

// Open returns a Conn by calling the DriverFunc itself.
func (f DriverFunc) Open(appName string, accessModel AccessModel, name string, migrate bool) (Conn, error) {
	return f(appName, accessModel, name, migrate)
}
