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
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/skydb"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// authInfoValue implements sql.Valuer and sql.Scanner s.t.
// skydb.AuthInfo can be saved into and recovered from postgresql
type authInfoValue struct {
	AuthInfo skydb.AuthInfo
	Valid    bool
}

func (auth authInfoValue) Value() (driver.Value, error) {
	if !auth.Valid {
		return nil, nil
	}

	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(auth.AuthInfo); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (auth *authInfoValue) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		fmt.Errorf("skydb: unsupported Scan pair: %T -> %T", value, auth.AuthInfo)
	}

	err := json.Unmarshal(b, &auth.AuthInfo)
	if err == nil {
		auth.Valid = true
	}
	return err
}

// Ext is an interface for both sqlx.DB and sqlx.Tx
type Ext interface {
	sqlx.Ext
	Get(dest interface{}, query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) *sql.Row
}

type conn struct {
	db             *sqlx.DB // database wrapper
	tx             *sqlx.Tx // transaction wrapper, nil when no transaction
	txDone         bool     // transaction is done, can only have one tx currently
	RecordSchema   map[string]skydb.RecordSchema
	appName        string
	option         string
	statementCount uint64
	accessModel    skydb.AccessModel
}

// Db returns the current database wrapper, or a transaction wrapper when
// a transaction is in effect.
func (c *conn) Db() Ext {
	if c.tx != nil {
		return c.tx
	}
	return c.db
}

// Begin begins a transaction.
func (c *conn) Begin() (err error) {
	log.Debugf("%p: Beginning transaction", c)
	if c.txDone {
		return skydb.ErrDatabaseTxDone
	}

	if c.tx != nil {
		return skydb.ErrDatabaseTxDidBegin
	}

	c.tx, err = c.db.Beginx()
	log.Debugf("%p: Done beginning transaction %p", c, c.tx)
	return
}

// Commit commits a transaction.
func (c *conn) Commit() (err error) {
	if c.txDone {
		return skydb.ErrDatabaseTxDone
	}

	if c.tx == nil {
		return skydb.ErrDatabaseTxDidNotBegin
	}

	err = c.tx.Commit()
	if err == nil {
		c.txDone = true
	}
	log.Debugf("%p: Committed transaction %p", c, c.tx)
	return
}

// Rollback rollbacks a transaction.
func (c *conn) Rollback() (err error) {
	if c.txDone {
		return skydb.ErrDatabaseTxDone
	}

	if c.tx == nil {
		return skydb.ErrDatabaseTxDidNotBegin
	}

	err = c.tx.Rollback()
	if err == nil {
		c.txDone = true
	}
	log.Debugf("%p: Rolled back transaction %p", c, c.tx)
	return
}

func (c *conn) PublicDB() skydb.Database {
	return &database{
		c:            c,
		databaseType: skydb.PublicDatabase,
	}
}

func (c *conn) PrivateDB(userKey string) skydb.Database {
	return &database{
		c:            c,
		databaseType: skydb.PrivateDatabase,
		userID:       userKey,
	}
}

func (c *conn) UnionDB() skydb.Database {
	return &database{
		c:            c,
		databaseType: skydb.UnionDatabase,
	}
}

func (c *conn) Close() error { return nil }

// return the raw unquoted schema name of this app
func (c *conn) schemaName() string {
	return "app_" + toLowerAndUnderscore(c.appName)
}

// return the quoted table name ready to be used as identifier (in the form
// "schema"."table")
func (c *conn) tableName(table string) string {
	return pq.QuoteIdentifier(c.schemaName()) + "." + pq.QuoteIdentifier(table)
}

type database struct {
	c            *conn
	userID       string
	txDone       bool
	databaseType skydb.DatabaseType
}

func (db *database) Conn() skydb.Conn       { return db.c }
func (db *database) UserRecordType() string { return "user" }

func (db *database) ID() string {
	if db.DatabaseType() == skydb.PublicDatabase {
		return skydb.PublicDatabaseIdentifier
	} else if db.DatabaseType() == skydb.UnionDatabase {
		return skydb.UnionDatabaseIdentifier
	}

	if db.userID == "" {
		panic("Private database but userID is empty")
	}
	return db.userID
}

func (db *database) DatabaseType() skydb.DatabaseType { return db.databaseType }
func (db *database) IsReadOnly() bool                 { return db.DatabaseType() == skydb.UnionDatabase }

// schemaName is a convenient method to access parent conn's schemaName
func (db *database) schemaName() string {
	return db.c.schemaName()
}

// tableName is a convenient method to access parent conn's tableName
func (db *database) tableName(table string) string {
	return db.c.tableName(table)
}

// this ensures that our structure conform to certain interfaces.
var (
	_ skydb.Conn     = &conn{}
	_ skydb.Database = &database{}

	_ driver.Valuer = authInfoValue{}
)
