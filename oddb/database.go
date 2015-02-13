package oddb

import (
	"fmt"
)

// A RecordNotFoundError is an implementation of error which represents
// that a record is not found in a Database.
type RecordNotFoundError struct {
	Key  string
	Conn Conn
}

func (e *RecordNotFoundError) Error() string {
	return fmt.Sprintf("Record of %v not found in Database", e.Key)
}

// Database represents a collection of record (either public or private)
// in a container.
type Database interface {
	ID() string
	Get(key string, record *Record) error
	Save(record *Record) error
	Delete(key string) error

	Query(query string, args ...interface{}) (Rows, error)
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
