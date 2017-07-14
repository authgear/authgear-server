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
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// providerInfoValue implements sql.Valuer and sql.Scanner s.t.
// skydb.ProviderInfo can be saved into and recovered from postgresql
type providerInfoValue struct {
	ProviderInfo skydb.ProviderInfo
	Valid        bool
}

func (auth providerInfoValue) Value() (driver.Value, error) {
	if !auth.Valid {
		return nil, nil
	}

	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(auth.ProviderInfo); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (auth *providerInfoValue) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		fmt.Errorf("skydb: unsupported Scan pair: %T -> %T", value, auth.ProviderInfo)
	}

	err := json.Unmarshal(b, &auth.ProviderInfo)
	if err == nil {
		auth.Valid = true
	}
	return err
}

// ExtContext is an interface for both sqlx.DB and sqlx.Tx
type ExtContext interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type conn struct {
	db             *sqlx.DB // database wrapper
	tx             *sqlx.Tx // transaction wrapper, nil when no transaction
	RecordSchema   map[string]skydb.RecordSchema
	FieldACL       *skydb.FieldACL
	appName        string
	authRecordKeys [][]string
	option         string
	statementCount uint64
	accessModel    skydb.AccessModel
	canMigrate     bool
	context        context.Context
}

// Db returns the current database wrapper, or a transaction wrapper when
// a transaction is in effect.
func (c *conn) Db() ExtContext {
	if c.tx != nil {
		return c.tx
	}
	return c.db
}

// Begin begins a transaction.
func (c *conn) Begin() error {
	log.Debugf("%p: Beginning transaction", c)
	if c.tx != nil {
		return skydb.ErrDatabaseTxDidBegin
	}

	tx, err := c.db.Beginx()
	if err != nil {
		log.Debugf("%p: Unable to begin transaction %p: %v", c, err)
		return err
	}
	c.tx = tx
	log.Debugf("%p: Done beginning transaction %p", c, c.tx)
	return nil
}

// Commit commits a transaction.
func (c *conn) Commit() error {
	if c.tx == nil {
		return skydb.ErrDatabaseTxDidNotBegin
	}

	if err := c.tx.Commit(); err != nil {
		log.Errorf("%p: Unable to commit transaction %p: %v", c, c.tx, err)
		return err
	}
	c.tx = nil
	log.Debugf("%p: Committed transaction", c)
	return nil
}

// Rollback rollbacks a transaction.
func (c *conn) Rollback() error {
	if c.tx == nil {
		return skydb.ErrDatabaseTxDidNotBegin
	}

	if err := c.tx.Rollback(); err != nil {
		log.Errorf("%p: Unable to rollback transaction %p: %v", c, c.tx, err)
		return err
	}
	c.tx = nil
	log.Debugf("%p: Rolled back transaction", c)
	return nil
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
func (db *database) TableName(table string) string {
	return db.c.tableName(table)
}

// this ensures that our structure conform to certain interfaces.
var (
	_ skydb.Conn     = &conn{}
	_ skydb.Database = &database{}

	_ driver.Valuer = providerInfoValue{}
)
