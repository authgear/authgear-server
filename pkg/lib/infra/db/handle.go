package db

// Handle allows a function to be run within a transaction.
type Handle interface {
	// WithTx runs do within a transaction.
	// If there is no error, the transaction is committed.
	WithTx(do func() error) (err error)

	// ReadOnly runs do within a transaction.
	// The transaction is always rolled back.
	ReadOnly(do func() error) (err error)

	// conn allows internal access to the ongoing transaction.
	conn() (*txConn, error)
}
