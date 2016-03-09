package skydb

import (
	"errors"
	"io"
)

// ErrRecordNotFound is returned from Get and Delete when Database
// cannot find the Record by the specified key
var ErrRecordNotFound = errors.New("skydb: Record not found for the specified key")

// EmptyRows is a convenient variable that acts as an empty Rows.
// Useful for skydb implementators and testing.
var EmptyRows = NewRows(emptyRowsIter(0))

type emptyRowsIter int

func (rs emptyRowsIter) Close() error {
	return nil
}

func (rs emptyRowsIter) Next(record *Record) error {
	return io.EOF
}

func (rs emptyRowsIter) OverallRecordCount() *uint64 {
	return nil
}

var ErrDatabaseTxDidBegin = errors.New("skydb: a transaction has already begun")
var ErrDatabaseTxDidNotBegin = errors.New("skydb: a transaction has not begun")
var ErrDatabaseTxDone = errors.New("skydb: Database's transaction has already commited or rolled back")

// Database represents a collection of record (either public or private)
// in a container.
type Database interface {

	// Conn returns the parent Conn of the Database
	Conn() Conn

	// ID returns the identifier of the Database.
	// We have public and private database. For public DB, the ID is
	// `_public`; for private, the ID is the user identifier
	ID() string

	// UserRecordType returns name of the user record type.
	UserRecordType() string

	// Get fetches the Record identified by the supplied key and
	// writes it onto the supplied Record.
	//
	// Get returns an ErrRecordNotFound if Record identified by
	// the supplied key does not exist in the Database.
	// It also returns error if the underlying implementation
	// failed to read the Record.
	Get(id RecordID, record *Record) error
	GetByIDs(ids []RecordID) (*Rows, error)

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
	Delete(id RecordID) error

	// Query executes the supplied query against the Database and returns
	// an Rows to iterate the results.
	Query(query *Query) (*Rows, error)

	// QueryCount executes the supplied query against the Database and returns
	// the number of records matching the query's predicate.
	QueryCount(query *Query) (uint64, error)

	// Extend extends the Database record schema such that a record
	// arrived subsequently with that schema can be saved
	//
	// Extend returns an error if the specified schema conflicts with
	// existing schem in the Database
	Extend(recordType string, schema RecordSchema) error

	// RenameSchema renames a column of the Database record schema
	RenameSchema(recordType, oldColumnName, newColumnName string) error

	// DeleteSchema removes a column of the Database record schema
	DeleteSchema(recordType, columnName string) error

	// GetSchema returns the record schema of a record type
	GetSchema(recordType string) (RecordSchema, error)

	// FetchRecordTypes returns a list of all existing record type
	GetRecordSchemas() (map[string]RecordSchema, error)

	GetSubscription(key string, deviceID string, subscription *Subscription) error
	SaveSubscription(subscription *Subscription) error
	DeleteSubscription(key string, deviceID string) error
	GetSubscriptionsByDeviceID(deviceID string) []Subscription
	GetMatchingSubscriptions(record *Record) []Subscription
}

// TxDatabase defines the methods for a Database that supports
// transaction.
//
// A Begin'ed transaction must end with a call to Commit or Rollback. After
// that, all opertions on Database will return ErrDatabaseTxDone.
//
// NOTE(limouren): The interface is not Database specific, but currently only
// Database supports it.
type TxDatabase interface {
	// Begin opens a transaction for the current Database.
	//
	// Calling Begin on an already Begin'ed Database returns ErrDatabaseTxDidBegin.
	Begin() error

	// Commit saves all the changes made to Database after Begin atomically.
	Commit() error

	// Rollbacks discards all the changes made to Database after Begin.
	Rollback() error
}

// Rows implements a scanner-like interface for easy iteration on a
// result set returned from a query
type Rows struct {
	iter        RowsIter
	lasterr     error
	closed      bool
	record      Record
	nexted      bool
	recordCount *uint64
}

// NewRows creates a new Rows.
//
// Driver implementators are expected to call this method with
// their implementation of RowsIter to return a Rows from Database.Query.
func NewRows(iter RowsIter) *Rows {
	return &Rows{
		iter: iter,
	}
}

// Close closes the Rows and prevents futher enumerations on the instance.
func (r *Rows) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true
	return r.iter.Close()
}

// Scan tries to prepare the next record and returns whether such record
// is ready to be read.
func (r *Rows) Scan() bool {
	if r.closed {
		return false
	}

	// Make a new record instead of reusing the same record from previous Scan.
	r.record = Record{}
	r.lasterr = r.iter.Next(&r.record)
	if r.lasterr != nil {
		r.Close()
		return false
	}

	return true
}

// Record returns the current record in Rows.
//
// It must be called after calling Scan and Scan returned true.
// If Scan is not called or previous Scan return false, the behaviour
// of Record is unspecified.
func (r *Rows) Record() Record {
	return r.record
}

// OverallRecordCount returns the number of matching records in the database
// if this resultset contains any rows.
func (r *Rows) OverallRecordCount() *uint64 {
	return r.iter.OverallRecordCount()
}

// Err returns the last error encountered during Scan.
//
// NOTE: It is not an error if the underlying result set is exhausted.
func (r *Rows) Err() error {
	if r.lasterr == io.EOF {
		return nil
	}

	return r.lasterr
}

// RowsIter is an iterator on results returned by execution of a query.
type RowsIter interface {
	// Close closes the rows iterator
	Close() error

	// Next populates the next Record in the current rows iterator into
	// the provided record.
	//
	// Next should return io.EOF when there are no more rows
	Next(record *Record) error

	OverallRecordCount() *uint64
}

// MemoryRows is a native implementation of RowIter.
// Can be used in test not support cursor.
type MemoryRows struct {
	CurrentRowIndex int
	Records         []Record
}

func NewMemoryRows(records []Record) *MemoryRows {
	return &MemoryRows{0, records}
}

func (rs *MemoryRows) Close() error {
	return nil
}

func (rs *MemoryRows) Next(record *Record) error {
	if rs.CurrentRowIndex >= len(rs.Records) {
		return io.EOF
	}

	*record = rs.Records[rs.CurrentRowIndex]
	rs.CurrentRowIndex = rs.CurrentRowIndex + 1
	return nil
}

func (rs *MemoryRows) OverallRecordCount() *uint64 {
	result := uint64(len(rs.Records))
	if result == 0 {
		return nil
	}
	return &result
}
