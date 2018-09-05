package db

import (
	"database/sql"
)

type DBConn struct {
	*sql.Conn
	ConnectionStr string
}

func (db DBConn) GetRecord(recordID string) string {
	return db.ConnectionStr + ":" + recordID
}
