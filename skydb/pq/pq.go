package pq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
	"github.com/oursky/skygear/skydb"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var underscoreRe = regexp.MustCompile(`[.:]`)

var dbs = map[string]*sqlx.DB{}
var dbsMutex sync.RWMutex

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

type userInfo struct {
	ID             string        `db:"id"`
	Email          string        `db:"email"`
	HashedPassword []byte        `db:"password"`
	Auth           authInfoValue `db:"auth"`
}

// authInfoValue implements sql.Valuer and sql.Scanner s.t.
// skydb.AuthInfo can be saved into and recovered from postgresql
type authInfoValue skydb.AuthInfo

func (auth authInfoValue) Value() (driver.Value, error) {
	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(auth); err != nil {
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
		fmt.Errorf("skydb: unsupported Scan pair: %T -> %T", value, auth)
	}

	return json.Unmarshal(b, auth)
}

type conn struct {
	Db      *sqlx.DB
	appName string
	option  string
}

func (c *conn) CreateUser(userinfo *skydb.UserInfo) error {
	sql, args, err := psql.Insert(c.tableName("_user")).
		Columns("id", "email", "password", "auth").
		Values(userinfo.ID, userinfo.Email, userinfo.HashedPassword, authInfoValue(userinfo.Auth)).
		ToSql()
	if err != nil {
		panic(err)
	}

	_, err = c.Db.Exec(sql, args...)
	if isUniqueViolated(err) {
		return skydb.ErrUserDuplicated
	}

	return err
}

func (c *conn) GetUser(id string, userinfo *skydb.UserInfo) error {
	selectSql, args, err := psql.Select("email", "password", "auth").
		From(c.tableName("_user")).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		panic(err)
	}

	email, password, auth := "", []byte{}, authInfoValue{}
	err = c.Db.QueryRow(selectSql, args...).Scan(
		&email,
		&password,
		&auth,
	)
	if err == sql.ErrNoRows {
		return skydb.ErrUserNotFound
	}

	userinfo.ID = id
	userinfo.Email = email
	userinfo.HashedPassword = password
	userinfo.Auth = skydb.AuthInfo(auth)

	return err
}

func (c *conn) GetUserByPrincipalID(principalID string, userinfo *skydb.UserInfo) error {
	selectSql, args, err := psql.Select("id", "email", "password", "auth").
		From(c.tableName("_user")).
		Where("jsonb_exists(auth, ?)", principalID).
		ToSql()
	if err != nil {
		panic(err)
	}

	id, email, password, auth := "", "", []byte{}, authInfoValue{}
	err = c.Db.QueryRow(selectSql, args...).Scan(
		&id,
		&email,
		&password,
		&auth,
	)
	if err == sql.ErrNoRows {
		return skydb.ErrUserNotFound
	}

	userinfo.ID = id
	userinfo.Email = email
	userinfo.HashedPassword = password
	userinfo.Auth = skydb.AuthInfo(auth)

	return err
}

func (c *conn) QueryUser(emails []string) ([]skydb.UserInfo, error) {

	emailargs := make([]interface{}, len(emails))
	for i, v := range emails {
		emailargs[i] = interface{}(v)
	}

	selectSql, args, err := psql.Select("id", "email", "password", "auth").
		From(c.tableName("_user")).
		Where("email IN ("+sq.Placeholders(len(emailargs))+") AND email IS NOT NULL AND email != ''", emailargs...).
		ToSql()
	if err != nil {
		panic(err)
	}

	rows, err := c.Db.Query(selectSql, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"sql":  selectSql,
			"args": args,
			"err":  err,
		}).Debugln("Failed to query user table")
		panic(err)
	}
	defer rows.Close()
	results := []skydb.UserInfo{}
	for rows.Next() {
		id, email, password, auth := "", "", []byte{}, authInfoValue{}
		if err := rows.Scan(&id, &email, &password, &auth); err != nil {
			panic(err)
		}

		userinfo := skydb.UserInfo{}
		userinfo.ID = id
		userinfo.Email = email
		userinfo.HashedPassword = password
		userinfo.Auth = skydb.AuthInfo(auth)
		results = append(results, userinfo)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return results, nil
}

func (c *conn) UpdateUser(userinfo *skydb.UserInfo) error {
	updateSql, args, err := psql.Update(c.tableName("_user")).
		Set("email", userinfo.Email).
		Set("password", userinfo.HashedPassword).
		Set("auth", authInfoValue(userinfo.Auth)).
		Where("id = ?", userinfo.ID).
		ToSql()
	if err != nil {
		panic(err)
	}

	result, err := c.Db.Exec(updateSql, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) DeleteUser(id string) error {
	query, args, err := psql.Delete(c.tableName("_user")).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		panic(err)
	}

	result, err := c.Db.Exec(query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows deleted, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) QueryRelation(user string, name string, direction string) []skydb.UserInfo {
	log.Debugf("Query Relation: %v, %v", user, name)
	tName := "_" + name
	var (
		selectSql string
		args      []interface{}
		err       error
	)
	if direction == "active" {
		selectSql, args, err = psql.Select("u.id", "u.email").
			From(c.tableName("_user")+" AS u").
			Join(c.tableName(tName)+" AS relation on relation.right_id = u.id").
			Where("relation.left_id = ?", user).
			ToSql()
	} else {
		selectSql, args, err = psql.Select("u.id", "u.email").
			From(c.tableName("_user")+" AS u").
			Join(c.tableName(tName)+" AS relation on relation.left_id = u.id").
			Where("relation.right_id = ?", user).
			ToSql()
	}
	if err != nil {
		panic(err)
	}
	rows, err := c.Db.Query(selectSql, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"sql":  selectSql,
			"args": args,
			"err":  err,
		}).Debugln("Failed to query relation")
		panic(err)
	}
	defer rows.Close()
	results := []skydb.UserInfo{}
	for rows.Next() {
		var id string
		var email string
		if err := rows.Scan(&id, &email); err != nil {
			panic(err)
		}
		userInfo := skydb.UserInfo{
			ID:    id,
			Email: email,
		}
		results = append(results, userInfo)
	}
	return results
}

func (c *conn) GetAsset(name string, asset *skydb.Asset) error {
	selectSql, args, err := psql.Select("content_type", "size").
		From(c.tableName("_asset")).
		Where("id = ?", name).
		ToSql()
	if err != nil {
		panic(err)
	}

	var (
		contentType string
		size        int64
	)
	err = c.Db.QueryRow(selectSql, args...).Scan(
		&contentType,
		&size,
	)
	if err == sql.ErrNoRows {
		return errors.New("asset not found")
	}

	asset.Name = name
	asset.ContentType = contentType
	asset.Size = size

	return err
}

func (c *conn) SaveAsset(asset *skydb.Asset) error {
	pkData := map[string]interface{}{
		"id": asset.Name,
	}
	data := map[string]interface{}{
		"content_type": asset.ContentType,
		"size":         asset.Size,
	}
	upsert := upsertQuery(c.tableName("_asset"), pkData, data)
	_, err := execWith(c.Db, upsert)
	if err != nil {
		sql, args, _ := upsert.ToSql()
		log.WithFields(log.Fields{
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Debugln("Failed to add asset")
	}

	return err
}

func (c *conn) AddRelation(user string, name string, targetUser string) error {
	tName := "_" + name
	ralationPair := map[string]interface{}{
		"left_id":  user,
		"right_id": targetUser,
	}

	upsert := upsertQuery(c.tableName(tName), ralationPair, nil)
	_, err := execWith(c.Db, upsert)
	if err != nil {
		sql, args, _ := upsert.ToSql()
		log.WithFields(log.Fields{
			"sql":  sql,
			"args": args,
			"err":  err,
		}).Debugln("Failed to add relation")
		if isForienKeyViolated(err) {
			return fmt.Errorf("userID not exist")
		}
	}

	return err
}

func (c *conn) RemoveRelation(user string, name string, targetUser string) error {
	tName := "_" + name

	builder := psql.Delete(c.tableName(tName)).
		Where("left_id = ? AND right_id = ?", user, targetUser)
	result, err := execWith(c.Db, builder)

	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%v relation not exist {%v} => {%v}",
			name, user, targetUser)
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}
	return nil
}

func (c *conn) GetDevice(id string, device *skydb.Device) error {
	builder := psql.Select("type", "token", "user_id", "last_registered_at").
		From(c.tableName("_device")).
		Where("id = ?", id)

	var nullToken sql.NullString
	err := queryRowWith(c.Db, builder).Scan(
		&device.Type,
		&nullToken,
		&device.UserInfoID,
		&device.LastRegisteredAt,
	)

	if err == sql.ErrNoRows {
		return skydb.ErrDeviceNotFound
	} else if err != nil {
		return err
	}

	device.Token = nullToken.String

	device.LastRegisteredAt = device.LastRegisteredAt.In(time.UTC)
	device.ID = id

	return nil
}

func (c *conn) QueryDevicesByUser(user string) ([]skydb.Device, error) {
	builder := psql.Select("id", "type", "token", "user_id", "last_registered_at").
		From(c.tableName("_device")).
		Where("user_id = ?", user)

	rows, err := queryWith(c.Db, builder)
	if err != nil {
		log.WithFields(log.Fields{
			"sql": builder,
			"err": err,
		}).Debugln("Failed to query device table")
		panic(err)
	}
	defer rows.Close()
	results := []skydb.Device{}
	for rows.Next() {
		d := skydb.Device{}
		if err := rows.Scan(
			&d.ID,
			&d.Type,
			&d.Token,
			&d.UserInfoID,
			&d.LastRegisteredAt); err != nil {

			panic(err)
		}
		d.LastRegisteredAt = d.LastRegisteredAt.UTC()
		results = append(results, d)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return results, nil
}

func (c *conn) SaveDevice(device *skydb.Device) error {
	if device.ID == "" || device.Type == "" || device.LastRegisteredAt.IsZero() {
		return errors.New("invalid device: empty id, type, or last registered at")
	}

	pkData := map[string]interface{}{"id": device.ID}
	data := map[string]interface{}{
		"type":               device.Type,
		"user_id":            device.UserInfoID,
		"last_registered_at": device.LastRegisteredAt.UTC(),
	}

	if device.Token != "" {
		data["token"] = device.Token
	}

	upsert := upsertQuery(c.tableName("_device"), pkData, data)
	_, err := execWith(c.Db, upsert)
	if err != nil {
		sql, args, _ := upsert.ToSql()
		log.WithFields(log.Fields{
			"sql":    sql,
			"args":   args,
			"err":    err,
			"device": device,
		}).Errorln("Failed to save device")
	}

	return err
}

func (c *conn) DeleteDevice(id string) error {
	builder := psql.Delete(c.tableName("_device")).
		Where("id = ?", id)
	result, err := execWith(c.Db, builder)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrDeviceNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) DeleteDeviceByToken(token string, t time.Time) error {
	builder := psql.Delete(c.tableName("_device")).
		Where("token = ?", token)
	if t != skydb.ZeroTime {
		builder = builder.Where("last_registered_at < ?", t)
	}
	result, err := execWith(c.Db, builder)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrDeviceNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) DeleteEmptyDeviceByTime(t time.Time) error {
	builder := psql.Delete(c.tableName("_device")).
		Where("token IS NULL")
	if t != skydb.ZeroTime {
		builder = builder.Where("last_registered_at < ?", t)
	}
	result, err := execWith(c.Db, builder)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return skydb.ErrDeviceNotFound
	}

	return nil
}

func (c *conn) PublicDB() skydb.Database {
	return &database{
		Db: c.Db,
		c:  c,
	}
}

func (c *conn) PrivateDB(userKey string) skydb.Database {
	return &database{
		Db:     c.Db,
		c:      c,
		userID: userKey,
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

type queryxRunner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
}

type database struct {
	Db     queryxRunner
	c      *conn
	userID string
	txDone bool
}

func (db *database) Conn() skydb.Conn { return db.c }
func (db *database) ID() string       { return "" }

// schemaName is a convenient method to access parent conn's schemaName
func (db *database) schemaName() string {
	return db.c.schemaName()
}

// tableName is a convenient method to access parent conn's tableName
func (db *database) tableName(table string) string {
	return db.c.tableName(table)
}

// Open returns a new connection to postgresql implementation
func Open(appName, connString string) (skydb.Conn, error) {
	var (
		db  *sqlx.DB
		err error
	)

	dbsMutex.RLock()
	db, ok := dbs[connString]
	dbsMutex.RUnlock()

	if ok {
		goto DB_OBTAINED
	}

	dbsMutex.Lock()
	db, ok = dbs[connString]
	if !ok {
		db, err = sqlx.Open("postgres", connString)
		if db != nil {
			db.SetMaxOpenConns(10)
			dbs[connString] = db
		}
	}
	dbsMutex.Unlock()

DB_OBTAINED:
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %s", err)
	}

	if err := mustInitDB(db, appName); err != nil {
		return nil, err
	}

	return &conn{
		Db:      db,
		appName: appName,
		option:  connString,
	}, nil
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
}

var (
	_ skydb.Conn     = &conn{}
	_ skydb.Database = &database{}

	_ driver.Valuer = authInfoValue{}
)

var createAppSchemaStmtTmpl *template.Template

const dbVersionNum = "51375067b45"
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
	email text,
	password text,
	auth jsonb
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
