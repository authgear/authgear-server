package pq

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/modl"
	sq "github.com/lann/squirrel"
	"regexp"
	"strings"

	"github.com/lib/pq"
	"github.com/oursky/ourd/oddb"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

var underscoreRe = regexp.MustCompile(`[.:]`)

func toLowerAndUnderscore(s string) string {
	return underscoreRe.ReplaceAllLiteralString(strings.ToLower(s), "_")
}

type conn struct {
	DBMap   *modl.DbMap
	appName string
}

func (c *conn) CreateUser(userinfo *oddb.UserInfo) error         { return nil }
func (c *conn) GetUser(id string, userinfo *oddb.UserInfo) error { return nil }
func (c *conn) UpdateUser(userinfo *oddb.UserInfo) error         { return nil }
func (c *conn) DeleteUser(id string) error                       { return nil }
func (c *conn) GetDevice(id string, device *oddb.Device) error   { return nil }
func (c *conn) SaveDevice(device *oddb.Device) error             { return nil }
func (c *conn) DeleteDevice(id string) error                     { return nil }

func (c *conn) PublicDB() oddb.Database {
	return &database{
		DBMap: c.DBMap,
		c:     c,
	}
}

func (c *conn) PrivateDB(userKey string) oddb.Database {
	return &database{
		DBMap: c.DBMap,
		c:     c,
	}
}

func (c *conn) AddDBRecordHook(hook oddb.DBHookFunc) {}
func (c *conn) Close() error                         { return nil }

func (c *conn) schemaName() string {
	return "app_" + toLowerAndUnderscore(c.appName)
}

type database struct {
	DBMap *modl.DbMap
	c     *conn
}

func (db *database) Conn() oddb.Conn { return db.c }
func (db *database) ID() string      { return "" }

func (db *database) Get(key string, record *oddb.Record) error {
	const SelectFmt = `SELECT * FROM %v.note WHERE _id = $1`

	m := map[string]interface{}{}
	err := db.DBMap.Get(&m, fmt.Sprintf(SelectFmt, db.schemaName()), key)
	if err != nil {
		return fmt.Errorf("get %v: %v", key, err)
	}

	delete(m, "_id")

	record.Key = key
	record.Type = "note"
	record.Data = m

	return nil
}

// Save attempts to do a upsert
func (db *database) Save(record *oddb.Record) error {
	const CreateTableFmt = `
CREATE TABLE IF NOT EXISTS %v.note (
	_id varchar(255) PRIMARY KEY,
	content text,
	createdDateTime timestamp,
	lastModified timestamp,
	noteOrder integer
)`

	if record.Type != "note" {
		return errors.New("only record type 'note' is supported for now")
	}

	createTableStmt := fmt.Sprintf(CreateTableFmt, db.schemaName())
	if _, err := db.DBMap.Exec(createTableStmt); err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	tablename := db.schemaName() + ".note"

	data := map[string]interface{}{}
	data["_id"] = record.Key
	for key, value := range record.Data {
		data[key] = value
	}
	insert := psql.Insert(tablename).SetMap(sq.Eq(data))

	sql, args, err := insert.ToSql()
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sql":  sql,
		"args": args,
	}).Debug("Inserting record")

	_, err = db.DBMap.Exec(sql, args...)

	pqErr, ok := err.(*pq.Error)
	// if duplicated insert
	if ok && pqErr.Code == "23505" {
		update := psql.Update(tablename).Where("_id = ?", record.Key).SetMap(sq.Eq(record.Data))

		sql, args, err = update.ToSql()
		if err != nil {
			panic(err)
		}

		log.WithFields(log.Fields{
			"sql":  sql,
			"args": args,
		}).Debug("Updating record")

		_, err = db.DBMap.Exec(sql, args...)
	}

	return err
}

func (db *database) Delete(key string) error { return nil }

func (db *database) Query(query *oddb.Query) (*oddb.Rows, error) { return &oddb.Rows{}, nil }
func (db *database) GetMatchingSubscription(record *oddb.Record) []oddb.Subscription {
	return []oddb.Subscription{}
}
func (db *database) GetSubscription(key string, subscription *oddb.Subscription) error { return nil }
func (db *database) SaveSubscription(subscription *oddb.Subscription) error            { return nil }
func (db *database) DeleteSubscription(key string) error                               { return nil }

// schemaName is a convenient method to access parent conn's schemaName
func (db *database) schemaName() string {
	return db.c.schemaName()
}

// Open returns a new connection to postgresql implementation
func Open(appName, connString string) (oddb.Conn, error) {
	const CreateSchemaFmt = `CREATE SCHEMA IF NOT EXISTS %v`

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	dbmap := modl.NewDbMap(db, modl.PostgresDialect{})

	stmt := fmt.Sprintf(CreateSchemaFmt, toLowerAndUnderscore("app_"+appName))
	if _, err := db.Exec(stmt); err != nil {
		return nil, err
	}

	return &conn{
		DBMap:   dbmap,
		appName: appName,
	}, nil
}

func init() {
	oddb.Register("pq", oddb.DriverFunc(Open))
}

var (
	_ oddb.Conn     = &conn{}
	_ oddb.Database = &database{}
)
