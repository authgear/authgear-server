package pq

import (
	"database/sql"

	"github.com/oursky/ourd/oddb"
)

type conn struct {
	sql.DB
}

func (c *conn) CreateUser(userinfo *oddb.UserInfo) error         { return nil }
func (c *conn) GetUser(id string, userinfo *oddb.UserInfo) error { return nil }
func (c *conn) UpdateUser(userinfo *oddb.UserInfo) error         { return nil }
func (c *conn) DeleteUser(id string) error                       { return nil }
func (c *conn) GetDevice(id string, device *oddb.Device) error   { return nil }
func (c *conn) SaveDevice(device *oddb.Device) error             { return nil }
func (c *conn) DeleteDevice(id string) error                     { return nil }
func (c *conn) PublicDB() oddb.Database                          { return &database{} }
func (c *conn) PrivateDB(userKey string) oddb.Database           { return &database{} }
func (c *conn) AddDBRecordHook(hook oddb.DBHookFunc)             {}
func (c *conn) Close() error                                     { return nil }

type database struct {
	sql.DB
	c *conn
}

func (db *database) Conn() oddb.Conn                             { return db.c }
func (db *database) ID() string                                  { return "" }
func (db *database) Get(key string, record *oddb.Record) error   { return nil }
func (db *database) Save(record *oddb.Record) error              { return nil }
func (db *database) Delete(key string) error                     { return nil }
func (db *database) Query(query *oddb.Query) (*oddb.Rows, error) { return &oddb.Rows{}, nil }
func (db *database) GetMatchingSubscription(record *oddb.Record) []oddb.Subscription {
	return []oddb.Subscription{}
}
func (db *database) GetSubscription(key string, subscription *oddb.Subscription) error { return nil }
func (db *database) SaveSubscription(subscription *oddb.Subscription) error            { return nil }
func (db *database) DeleteSubscription(key string) error                               { return nil }

// Open returns a new connection to postgresql implementation
func Open(appName, dir string) (oddb.Conn, error) {
	return &conn{}, nil
}

func init() {
	oddb.Register("pq", oddb.DriverFunc(Open))
}

var (
	_ oddb.Conn     = &conn{}
	_ oddb.Database = &database{}
)
