package pq

import (
	"database/sql"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/pq/migration"
)

var underscoreRe = regexp.MustCompile(`[.:]`)

func toLowerAndUnderscore(s string) string {
	return underscoreRe.ReplaceAllLiteralString(strings.ToLower(s), "_")
}

func isForienKeyViolated(err error) bool {
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
func Open(appName, connString string) (skydb.Conn, error) {
	db, err := getDB(appName, connString)
	if err != nil {
		return nil, err
	}

	return &conn{
		db:           db,
		RecordSchema: map[string]skydb.RecordSchema{},
		appName:      appName,
		option:       connString,
	}, nil
}

type getDBReq struct {
	appName    string
	connString string
	done       chan getDBResp
}

type getDBResp struct {
	db  *sqlx.DB
	err error
}

var dbs = map[string]*sqlx.DB{}
var getDBChan = make(chan getDBReq)

func getDB(appName, connString string) (*sqlx.DB, error) {
	ch := make(chan getDBResp)
	getDBChan <- getDBReq{appName, connString, ch}
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
			dbs[req.connString] = db
		}

		if err := mustInitDB(db, req.appName); err != nil {
			req.done <- getDBResp{nil, fmt.Errorf("failed to open connection: %s", err)}
			continue
		}

		req.done <- getDBResp{db, nil}
	}
}

// mustInitDB initialize database objects for an application.
func mustInitDB(db *sqlx.DB, appName string) error {
	schema := pq.QuoteIdentifier("app_" + toLowerAndUnderscore(appName))

	var versionNum string
	err := db.QueryRowx(fmt.Sprintf("SELECT version_num FROM %s._version", schema)).
		Scan(&versionNum)

	if err == sql.ErrNoRows || isUndefinedTable(err) {
		// ignore the err here; they are unimportant
		// do nothing
	} else if isNetworkError(err) {
		return fmt.Errorf("skydb/pq: unable to connect to database because of a network error = %v", err)
	} else if err != nil {
		return fmt.Errorf("skydb/pq: unrecgonized error while querying db version_num = %v", err)
	}

	// begin transactional DDL
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("skydb/pq: failed to begin transaction for DDL: %v", err)
	}
	defer tx.Rollback()

	if versionNum == migration.DbVersionNum {
		return nil
	} else if versionNum == "" {
		if err := migration.InitSchema(tx, schema); err != nil {
			return fmt.Errorf("skydb/pq: failed to init database: %v", err)
		}
	} else {
		return fmt.Errorf("skydb/pq: got version_num = %s, want %s", versionNum, migration.DbVersionNum)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("skydb/pq: failed to commit DDL: %v", err)
	}

	return nil
}

func init() {
	skydb.Register("pq", skydb.DriverFunc(Open))
	go dbInitializer()
}
