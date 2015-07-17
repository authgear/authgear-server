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

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	sq "github.com/lann/squirrel"
	"github.com/lib/pq"

	"github.com/oursky/ourd/oddb"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var underscoreRe = regexp.MustCompile(`[.:]`)

var initDBOnce sync.Once
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
// oddb.AuthInfo can be saved into and recovered from postgresql
type authInfoValue oddb.AuthInfo

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
		fmt.Errorf("oddb: unsupported Scan pair: %T -> %T", value, auth)
	}

	return json.Unmarshal(b, auth)
}

type conn struct {
	Db      *sqlx.DB
	appName string
	option  string
}

func (c *conn) CreateUser(userinfo *oddb.UserInfo) error {
	sql, args, err := psql.Insert(c.tableName("_user")).
		Columns("id", "email", "password", "auth").
		Values(userinfo.ID, userinfo.Email, userinfo.HashedPassword, authInfoValue(userinfo.Auth)).
		ToSql()
	if err != nil {
		panic(err)
	}

	_, err = c.Db.Exec(sql, args...)
	if isUniqueViolated(err) {
		return oddb.ErrUserDuplicated
	}

	return err
}

func (c *conn) GetUser(id string, userinfo *oddb.UserInfo) error {
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
		return oddb.ErrUserNotFound
	}

	userinfo.ID = id
	userinfo.Email = email
	userinfo.HashedPassword = password
	userinfo.Auth = oddb.AuthInfo(auth)

	return err
}

func (c *conn) QueryUser(emails []string) ([]oddb.UserInfo, error) {

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
	results := []oddb.UserInfo{}
	for rows.Next() {
		id, email, password, auth := "", "", []byte{}, authInfoValue{}
		if err := rows.Scan(&id, &email, &password, &auth); err != nil {
			panic(err)
		}

		userinfo := oddb.UserInfo{}
		userinfo.ID = id
		userinfo.Email = email
		userinfo.HashedPassword = password
		userinfo.Auth = oddb.AuthInfo(auth)
		results = append(results, userinfo)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return results, nil
}

func (c *conn) UpdateUser(userinfo *oddb.UserInfo) error {
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
		return oddb.ErrUserNotFound
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
		return oddb.ErrUserNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows deleted, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) QueryRelation(user string, name string, direction string) []oddb.UserInfo {
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
	results := []oddb.UserInfo{}
	for rows.Next() {
		var id string
		var email string
		if err := rows.Scan(&id, &email); err != nil {
			panic(err)
		}
		userInfo := oddb.UserInfo{
			ID:    id,
			Email: email,
		}
		results = append(results, userInfo)
	}
	return results
}

func (c *conn) GetAsset(name string, asset *oddb.Asset) error {
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

func (c *conn) SaveAsset(asset *oddb.Asset) error {
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

func (c *conn) GetDevice(id string, device *oddb.Device) error {
	builder := psql.Select("type", "token", "user_id").
		From(c.tableName("_device")).
		Where("id = ?", id)
	err := queryRowWith(c.Db, builder).
		Scan(&device.Type, &device.Token, &device.UserInfoID)

	if err == sql.ErrNoRows {
		return oddb.ErrDeviceNotFound
	} else if err != nil {
		return err
	}

	device.ID = id

	return nil
}

func (c *conn) SaveDevice(device *oddb.Device) error {
	if device.ID == "" || device.Token == "" || device.Type == "" || device.UserInfoID == "" {
		return errors.New("invalid device: empty id or token or type or user id")
	}

	pkData := map[string]interface{}{"id": device.ID}
	data := map[string]interface{}{
		"type":    device.Type,
		"token":   device.Token,
		"user_id": device.UserInfoID,
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
		return oddb.ErrDeviceNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (c *conn) PublicDB() oddb.Database {
	return &database{
		Db: c.Db,
		c:  c,
	}
}

func (c *conn) PrivateDB(userKey string) oddb.Database {
	return &database{
		Db:     c.Db,
		c:      c,
		userID: userKey,
	}
}

func (c *conn) Close() error { return nil }

func (c *conn) schemaName() string {
	return "app_" + toLowerAndUnderscore(c.appName)
}

func (c *conn) tableName(table string) string {
	return c.schemaName() + "." + table
}

type database struct {
	Db     *sqlx.DB
	c      *conn
	userID string
}

func (db *database) Conn() oddb.Conn { return db.c }
func (db *database) ID() string      { return "" }

// schemaName is a convenient method to access parent conn's schemaName
func (db *database) schemaName() string {
	return db.c.schemaName()
}

// tableName is a convenient method to access parent conn's tableName
func (db *database) tableName(table string) string {
	return db.c.tableName(table)
}

// Open returns a new connection to postgresql implementation
func Open(appName, connString string) (oddb.Conn, error) {
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

	// TODO: it might be desirable to init DB in start-up time.
	initDBOnce.Do(func() { mustInitDB(db) })

	if err := initAppDB(db, appName); err != nil {
		return nil, fmt.Errorf("failed to init db: %s", err)
	}

	return &conn{
		Db:      db,
		appName: appName,
		option:  connString,
	}, nil
}

// mustInitDB initialize database objects shared across all schemata.
func mustInitDB(db *sqlx.DB) {
	const CreatePendingNotificationTableStmt = `CREATE TABLE IF NOT EXISTS pending_notification (
	id SERIAL NOT NULL PRIMARY KEY,
	op text NOT NULL,
	appname text NOT NULL,
	recordtype text NOT NULL,
	record jsonb NOT NULL
);
`
	const CreateNotificationTriggerFuncStmt = `CREATE OR REPLACE FUNCTION public.notify_record_change() RETURNS TRIGGER AS $$
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
`

	db.MustExec(CreatePendingNotificationTableStmt)
	db.MustExec(CreateNotificationTriggerFuncStmt)
}

func initAppDB(db *sqlx.DB, appName string) error {
	const CreateSchemaFmt = `CREATE SCHEMA IF NOT EXISTS %v;`
	const CreateUserTableFmt = `
CREATE TABLE IF NOT EXISTS %v._user (
	id varchar(255) PRIMARY KEY,
	email varchar(255),
	password varchar(255),
	auth json
);
`
	const CreateAssetTableFmt = `
CREATE TABLE IF NOT EXISTS %v._asset (
	id text PRIMARY KEY,
	content_type text,
	size bigint
);
`
	const CreateDeviceTableFmt = `
CREATE TABLE IF NOT EXISTS %[1]v._device (
	id text PRIMARY KEY,
	user_id text REFERENCES %[1]v._user (id),
	type text NOT NULL,
	token text NOT NULL,
	UNIQUE (user_id, type, token)
);
`
	const CreateSubscriptionTableFmt = `
CREATE TABLE IF NOT EXISTS %[1]v._subscription (
	id text NOT NULL,
	user_id text NOT NULL,
	device_id text REFERENCES %[1]v._device (id) NOT NULL,
	type text NOT NULL,
	notification_info jsonb,
	query jsonb,
	PRIMARY KEY(user_id, device_id, id)
);
`

	const CreateRelationTableFmt = `
CREATE TABLE IF NOT EXISTS %[1]v._friend (
	left_id text NOT NULL,
	right_id text REFERENCES %[1]v._user (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
CREATE TABLE IF NOT EXISTS %[1]v._follow (
	left_id text NOT NULL,
	right_id text REFERENCES %[1]v._user (id) NOT NULL,
	PRIMARY KEY(left_id, right_id)
);
`

	schemaName := "app_" + toLowerAndUnderscore(appName)

	createSchemaStmt := fmt.Sprintf(CreateSchemaFmt, schemaName)
	if _, err := db.Exec(createSchemaStmt); err != nil {
		return fmt.Errorf("failed to create schema: %s", err)
	}

	createUserTableStmt := fmt.Sprintf(CreateUserTableFmt, schemaName)
	if _, err := db.Exec(createUserTableStmt); err != nil {
		return fmt.Errorf("failed to create user table: %s", err)
	}

	createAssetTableStmt := fmt.Sprintf(CreateAssetTableFmt, schemaName)
	if _, err := db.Exec(createAssetTableStmt); err != nil {
		return fmt.Errorf("failed to create asset table: %s", err)
	}

	createDeviceTableStmt := fmt.Sprintf(CreateDeviceTableFmt, schemaName)
	if _, err := db.Exec(createDeviceTableStmt); err != nil {
		return fmt.Errorf("failed to create device table: %s", err)
	}

	createSubscriptionTableSmt := fmt.Sprintf(CreateSubscriptionTableFmt, schemaName)
	if _, err := db.Exec(createSubscriptionTableSmt); err != nil {
		return fmt.Errorf("failed to create subscription table: %s", err)
	}

	createRelationTableSmt := fmt.Sprintf(CreateRelationTableFmt, schemaName)
	if _, err := db.Exec(createRelationTableSmt); err != nil {
		return fmt.Errorf("failed to create relation table: %s", err)
	}

	return nil
}

type sqlizer sq.Sqlizer

func execWith(db *sqlx.DB, sqlizeri sqlizer) (sql.Result, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return db.Exec(sql, args...)
}

func queryWith(db *sqlx.DB, sqlizeri sqlizer) (*sqlx.Rows, error) {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return db.Queryx(sql, args...)
}

func queryRowWith(db *sqlx.DB, sqlizeri sqlizer) *sqlx.Row {
	sql, args, err := sqlizeri.ToSql()
	if err != nil {
		panic(err)
	}
	return db.QueryRowx(sql, args...)
}

func init() {
	oddb.Register("pq", oddb.DriverFunc(Open))
}

var (
	_ oddb.Conn     = &conn{}
	_ oddb.Database = &database{}

	_ driver.Valuer = authInfoValue{}
)
