package pq

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"regexp"
	"strings"
	"text/template"

	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/oursky/skygear/skydb"
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
		Db:           db,
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

	if versionNum == dbVersionNum {
		return nil
	} else if versionNum == "" {
		if err := initSchema(tx, schema); err != nil {
			return fmt.Errorf("skydb/pq: failed to init database: %v", err)
		}
	} else {
		return fmt.Errorf("skydb/pq: got version_num = %s, want %s", versionNum, dbVersionNum)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("skydb/pq: failed to commit DDL: %v", err)
	}

	return nil
}

func initSchema(tx *sqlx.Tx, schema string) error {
	stmt, err := templateExecString(createAppSchemaStmtTmpl, struct {
		Schema     string
		VersionNum string
	}{schema, dbVersionNum})

	if err != nil {
		return err
	}

	_, err = tx.Exec(stmt)
	return err
}

func templateExecString(t *template.Template, i interface{}) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, i); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type sqlizer sq.Sqlizer

func execWith(db queryxRunner, sqlizeri sqlizer) (sql.Result, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return db.Exec(sql, args...)
}

func queryWith(db queryxRunner, sqlizeri sqlizer) (*sqlx.Rows, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return db.Queryx(sql, args...)
}

func queryRowWith(db queryxRunner, sqlizeri sqlizer) *sqlx.Row {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return db.QueryRowx(sql, args...)
}

func init() {
	createAppSchemaStmtTmpl = template.Must(template.New("createAppSchemaStmtTmpl").Parse(createAppSchemaStmtTmplText))
	skydb.Register("pq", skydb.DriverFunc(Open))
	go dbInitializer()
}

var createAppSchemaStmtTmpl *template.Template

const dbVersionNum = "30d0a626888"
const createAppSchemaStmtTmplText = `
CREATE SCHEMA IF NOT EXISTS {{.Schema}};
CREATE TABLE IF NOT EXISTS public.pending_notification (
	id SERIAL NOT NULL PRIMARY KEY,
	op text NOT NULL,
	appname text NOT NULL,
	recordtype text NOT NULL,
	record jsonb NOT NULL
);
CREATE OR REPLACE FUNCTION public.notify_record_change() RETURNS TRIGGER AS $$
	DECLARE
		affected_record RECORD;
		inserted_id integer;
	BEGIN
		IF (TG_OP = 'DELETE') THEN
			affected_record := OLD;
		ELSE
			affected_record := NEW;
		END IF;
		INSERT INTO pending_notification (op, appname, recordtype, record)
			VALUES (TG_OP, TG_TABLE_SCHEMA, TG_TABLE_NAME, row_to_json(affected_record)::jsonb)
			RETURNING id INTO inserted_id;
		PERFORM pg_notify('record_change', inserted_id::TEXT);
		RETURN affected_record;
	END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS {{.Schema}}._version (
	version_num character varying(32) NOT NULL
);
INSERT INTO {{.Schema}}._version (version_num) VALUES('{{.VersionNum}}');

CREATE TABLE {{.Schema}}._user (
	id text PRIMARY KEY,
	username text,
	email text,
	password text,
	auth jsonb,
	UNIQUE (username),
	UNIQUE (email)
);
CREATE TABLE {{.Schema}}._asset (
	id text PRIMARY KEY,
	content_type text NOT NULL,
	size bigint NOT NULL
);
CREATE TABLE {{.Schema}}._device (
	id text PRIMARY KEY,
	user_id text REFERENCES {{.Schema}}._user (id),
	type text NOT NULL,
	token text,
	last_registered_at timestamp without time zone NOT NULL,
	UNIQUE (user_id, type, token)
);
CREATE INDEX ON {{.Schema}}._device (token, last_registered_at);
CREATE TABLE {{.Schema}}._subscription (
	id text NOT NULL,
	user_id text NOT NULL,
	device_id text REFERENCES {{.Schema}}._device (id) ON DELETE CASCADE NOT NULL,
	type text NOT NULL,
	notification_info jsonb,
	query jsonb,
	PRIMARY KEY(user_id, device_id, id)
);
CREATE TABLE {{.Schema}}._friend (
	left_id text NOT NULL,
	right_id text REFERENCES {{.Schema}}._user (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
CREATE TABLE {{.Schema}}._follow (
	left_id text NOT NULL,
	right_id text REFERENCES {{.Schema}}._user (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
`
