package record

import "io"

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

// Close closes the Rows and prevents further enumerations on the instance.
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
