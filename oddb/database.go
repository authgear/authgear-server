package oddb

import (
	"fmt"
)

// A RecordNotFoundError is an implementation of error which represents
// a record is not found in a Database
type RecordNotFoundError struct {
	Key  string
	Conn Conn
}

func (e *RecordNotFoundError) Error() string {
	return fmt.Sprintf("Record of %v not found in Database", e.Key)
}

// A Database represents a collection of record (either public or private)
// in a container
type Database interface {
	ID() string
	Get(key string, record *Record) error
	Save(record *Record) error
	Delete(key string) error
}
