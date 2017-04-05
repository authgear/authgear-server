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
	"database/sql"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/pq/migration"
)

var log = logging.LoggerEntry("skydb")

var underscoreRe = regexp.MustCompile(`[.:]`)

func toLowerAndUnderscore(s string) string {
	return underscoreRe.ReplaceAllLiteralString(strings.ToLower(s), "_")
}

func isForeignKeyViolated(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
		return true
	}

	return false
}

func isUniqueViolated(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return true
	}

	return false
}

func isInvalidInputSyntax(err error) bool {
	pqErr, ok := err.(*pq.Error)
	return ok && (pqErr.Code == "22P02" || pqErr.Code == "22P03")
}

func isUndefinedTable(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "42P01" {
		return true
	}

	return false
}

func isNetworkError(err error) bool {
	_, ok := err.(*net.OpError)
	return ok
}

type queryxRunner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
}

// Open returns a new connection to postgresql implementation
func Open(appName string, accessModel skydb.AccessModel, connString string, migrate bool) (skydb.Conn, error) {
	db, err := getDB(appName, connString, migrate)
	if err != nil {
		return nil, err
	}
	if accessModel == skydb.RelationBasedAccess {
		return nil, fmt.Errorf("Unsupported AccessModel: RelationBasedAccess")
	}

	return &conn{
		db:           db,
		RecordSchema: map[string]skydb.RecordSchema{},
		appName:      appName,
		option:       connString,
		accessModel:  accessModel,
		canMigrate:   migrate,
	}, nil
}

type getDBReq struct {
	appName    string
	connString string
	migrate    bool
	done       chan getDBResp
}

type getDBResp struct {
	db  *sqlx.DB
	err error
}

var dbs = map[string]*sqlx.DB{}
var getDBChan = make(chan getDBReq)

func getDB(appName, connString string, migrate bool) (*sqlx.DB, error) {
	ch := make(chan getDBResp)
	getDBChan <- getDBReq{appName, connString, migrate, ch}
	resp := <-ch
	return resp.db, resp.err
}

// goroutine that initialize the database for use
func dbInitializer() {
	for {
		req := <-getDBChan
		db, ok := dbs[req.connString]
		if !ok {
			var err error
			db, err = sqlx.Open("postgres", req.connString)
			if err != nil {
				req.done <- getDBResp{nil, fmt.Errorf("failed to open connection: %s", err)}
				continue
			}

			db.SetMaxOpenConns(10)

			if err := mustInitDB(db, req.appName, req.migrate); err != nil {
				db.Close()
				req.done <- getDBResp{nil, fmt.Errorf("failed to open connection: %s", err)}
				continue
			}

			dbs[req.connString] = db
		}

		req.done <- getDBResp{db, nil}
	}
}

// mustInitDB initialize database objects for an application.
func mustInitDB(db *sqlx.DB, appName string, migrate bool) error {
	schema := "app_" + toLowerAndUnderscore(appName)
	err := migration.EnsureLatest(db, schema, migrate)

	if err != nil {
		if isNetworkError(err) {
			return fmt.Errorf("skydb/pq: unable to connect to database because of a network error = %v", err)
		} else if err == migration.ErrMigrationDisabled {
			log.Warnf(`Schema does not match required version and migration ` +
				`is disabled. Database schema can only be modified in dev-mode.`)
			return fmt.Errorf("skydb/pq: unable to open database because schema does not match required version")
		} else {
			return fmt.Errorf("skydb/pq: unable to migrate database because of an error = %v", err)
		}
	}
	return nil
}

func init() {
	skydb.Register("pq", skydb.DriverFunc(Open))
	go dbInitializer()
}
