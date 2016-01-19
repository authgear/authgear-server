package pq

import (
	"database/sql"
	"os"
	"testing"

	"github.com/lib/pq"
	"github.com/oursky/skygear/skydb"
)

// NOTE(limouren): postgresql uses this error to signify a non-exist
// schema
func isInvalidSchemaName(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "3F000" {
		return true
	}

	return false
}

func getTestConn(t *testing.T) *conn {
	defaultTo := func(envvar string, value string) {
		if os.Getenv(envvar) == "" {
			os.Setenv(envvar, value)
		}
	}
	defaultTo("PGDATABASE", "skygear_test")
	defaultTo("PGSSLMODE", "disable")
	c, err := Open("com.oursky.skygear", skydb.RoleBasedAccess, "")
	if err != nil {
		t.Fatal(err)
	}
	return c.(*conn)
}

func cleanupConn(t *testing.T, c *conn) {
	_, err := c.db.Exec("DROP SCHEMA app_com_oursky_skygear CASCADE")
	if err != nil && !isInvalidSchemaName(err) {
		t.Fatal(err)
	}
}

func addUser(t *testing.T, c *conn, userid string) {
	_, err := c.Exec("INSERT INTO app_com_oursky_skygear._user (id, password) VALUES ($1, 'somepassword')", userid)
	if err != nil {
		t.Fatal(err)
	}
}

type execor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func insertRow(t *testing.T, db execor, query string, args ...interface{}) {
	result, err := db.Exec(query, args...)
	if err != nil {
		t.Fatal(err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}

	if n != 1 {
		t.Fatalf("got rows affected = %v, want 1", n)
	}
}

func exhaustRows(rows *skydb.Rows, errin error) (records []skydb.Record, err error) {
	if errin != nil {
		err = errin
		return
	}

	for rows.Scan() {
		records = append(records, rows.Record())
	}

	err = rows.Err()
	return
}
