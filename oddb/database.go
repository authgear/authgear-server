package oddb

import "errors"

// ErrRecordNotFound is returned from Get and Delete when Database
// cannot find the Record by the specified key
var ErrRecordNotFound = errors.New("oddb: Record not found for the specified key")

// Database represents a collection of record (either public or private)
// in a container.
//
// TODO: We might need to define standard errors for common failures
// of database operations like ErrRecordNotFound
type Database interface {
	// ID returns the identifier of the Database.
	ID() string

	// Get fetches the Record identified by the supplied key and
	// writes it onto the supplied Record.
	//
	// Get returns an ErrRecordNotFound if Record identified by
	// the supplied key does not exist in the Database.
	// It also returns error if the underlying implementation
	// failed to read the Record.
	Get(key string, record *Record) error

	// Save updates the supplied Record in the Database if Record with
	// the same key exists, else such Record is created.
	//
	// Save returns an error if the underlying implemention failed to
	// create / modify the Record.
	Save(record *Record) error

	// Delete removes the Record identified by the key in the Database.
	//
	// Delete returns an ErrRecordNotFound if the Record identified by
	// the supplied key does not exist in the Database.
	// It also returns an error if the underlying implementation
	// failed to remove the Record.
	Delete(key string) error

	// Query executes the supplied query against the Database and returns
	// an Rows to iterate the results.
	Query(query *Query) (Rows, error)
}

// Rows is a cursor returned by execution of a query.
type Rows interface {
	// Close closes the rows iterator
	Close() error

	// Next populates the next Record in the current rows iterator into
	// the provided record.
	//
	// Next should return io.EOF when there are no more rows
	Next(record *Record) error
}
