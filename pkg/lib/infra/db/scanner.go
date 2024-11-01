package db

// Scanner is *sql.Row or *sql.Rows.
type Scanner interface {
	Scan(dest ...interface{}) error
}
