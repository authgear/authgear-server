// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pq

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

// NOTE(limouren): postgresql uses this error to signify a non-exist
// schema
func isInvalidSchemaName(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "3F000" {
		return true
	}

	return false
}

func testAppName() string {
	return "io.skygear.test"
}

func getTestConn(t *testing.T) *conn {
	if runtime.GOMAXPROCS(0) > 1 {
		t.Skip("skipping zmq test in GOMAXPROCS>1")
	}
	defaultTo := func(envvar string, value string) {
		if os.Getenv(envvar) == "" {
			os.Setenv(envvar, value)
		}
	}
	defaultTo("PGDATABASE", "skygear_test")
	defaultTo("PGSSLMODE", "disable")
	appName := testAppName()
	c, err := Open(context.Background(), appName, skydb.RoleBasedAccess, "", true)
	if err != nil {
		t.Fatal(err)
	}

	// create schema
	err = mustInitDB(c.(*conn).Db().(*sqlx.DB), appName, true)
	if err != nil {
		t.Fatal(err)
	}
	return c.(*conn)
}

func dropAllRecordTables(t *testing.T, c *conn) {
	tx, err := c.db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	db := c.PublicDB().(*database)
	for recordType := range c.RecordSchema {
		if err := dropTable(tx, db.TableName(recordType)); err != nil {
			t.Fatal(err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func cleanupConn(t *testing.T, c *conn) {
	if len(c.RecordSchema) > 0 {
		dropAllRecordTables(t, c)
	}

	// Do not use the context.Context from conn struct because
	// we don't want a cancelled context preventing clean up.
	ctx := context.Background()

	schemaName := fmt.Sprintf("app_%s", toLowerAndUnderscore(c.appName))
	_, err := c.db.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA if exists %s CASCADE", schemaName))
	if err != nil && !isInvalidSchemaName(err) {
		t.Fatal(err)
	}
}

func addAuth(t *testing.T, c *conn, userid string) {
	if _, err := c.Exec("INSERT INTO _auth (id, password) VALUES ($1, 'somepassword')", userid); err != nil {
		t.Fatal(err)
	}
}

func addUser(t *testing.T, c *conn, userid string) {
	addAuth(t, c, userid)

	_, err := c.Exec(`INSERT INTO "user" (_id, _owner_id, _database_id, _created_at, _updated_at) VALUES ($1, $1, '', now(), now())`, userid)
	if err != nil {
		t.Fatal(err)
	}
}

func addUserWithInfo(t *testing.T, c *conn, userid string, email string) {
	addAuth(t, c, userid)

	_, err := c.Exec(`INSERT INTO "user" (_id, _owner_id, _database_id, _created_at, _updated_at, email) VALUES ($1, $1, '', now(), now(), $2)`, userid, email)
	if err != nil {
		t.Fatal(err)
	}
}

func addUserWithUsername(t *testing.T, c *conn, userid string, username string) {
	addAuth(t, c, userid)

	_, err := c.Exec(`INSERT INTO "user" (_id, _owner_id, _database_id, _created_at, _updated_at, username) VALUES ($1, $1, '', now(), now(), $2)`, userid, username)
	if err != nil {
		t.Fatal(err)
	}
}

type execor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func insertRow(t *testing.T, db execor, query string, args ...interface{}) {
	result, err := db.ExecContext(context.Background(), query, args...)
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
