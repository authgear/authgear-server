package db

// Scanner is sqlx.Row or sqlx.Rows.
type Scanner interface {
	Scan(dest ...interface{}) error
}
