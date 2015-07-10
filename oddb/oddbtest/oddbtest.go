package oddbtest

import (
	"github.com/oursky/ourd/oddb"
)

// MapConn is a naive memory implementation of oddb.Conn
type MapConn struct {
	UserMap map[string]oddb.UserInfo
}

// NewMapConn returns a new MapConn.
func NewMapConn() *MapConn {
	return &MapConn{
		UserMap: map[string]oddb.UserInfo{},
	}
}

// CreateUser creates a UserInfo in UserMap.
func (conn *MapConn) CreateUser(userinfo *oddb.UserInfo) error {
	if _, existed := conn.UserMap[userinfo.ID]; existed {
		return oddb.ErrUserDuplicated
	}

	conn.UserMap[userinfo.ID] = *userinfo
	return nil
}

// GetUser returns a UserInfo in UserMap.
func (conn *MapConn) GetUser(id string, userinfo *oddb.UserInfo) error {
	u, ok := conn.UserMap[id]
	if !ok {
		return oddb.ErrUserNotFound
	}

	*userinfo = u
	return nil
}

func (conn *MapConn) QueryUser(emails []string) ([]oddb.UserInfo, error) {
	panic("not implemented")
}

// UpdateUser updates an existing UserInfo in UserMap.
func (conn *MapConn) UpdateUser(userinfo *oddb.UserInfo) error {
	if _, ok := conn.UserMap[userinfo.ID]; !ok {
		return oddb.ErrUserNotFound
	}

	conn.UserMap[userinfo.ID] = *userinfo
	return nil
}

// DeleteUser remove an existing in UserMap.
func (conn *MapConn) DeleteUser(id string) error {
	if _, ok := conn.UserMap[id]; !ok {
		return oddb.ErrUserNotFound
	}

	delete(conn.UserMap, id)
	return nil
}

func (conn *MapConn) GetAsset(name string, asset *oddb.Asset) error {
	panic("not implemented")
}

func (conn *MapConn) SaveAsset(asset *oddb.Asset) error {
	panic("not implemented")
}

func (conn *MapConn) QueryRelation(user string, name string, direction string) []oddb.UserInfo {
	panic("not implemented")
}

func (conn *MapConn) AddRelation(user string, name string, targetUser string) error {
	panic("not implemented")
}

func (conn *MapConn) RemoveRelation(user string, name string, targetUser string) error {
	panic("not implemented")
}

// GetDevice is not implemented.
func (conn *MapConn) GetDevice(id string, device *oddb.Device) error {
	panic("not implemented")
}

// SaveDevice is not implemented.
func (conn *MapConn) SaveDevice(device *oddb.Device) error {
	panic("not implemented")
}

// DeleteDevice is not implemented.
func (conn *MapConn) DeleteDevice(id string) error {
	panic("not implemented")
}

// PublicDB is not implemented.
func (conn *MapConn) PublicDB() oddb.Database {
	panic("not implemented")
}

// PrivateDB is not implemented.
func (conn *MapConn) PrivateDB(userKey string) oddb.Database {
	panic("not implemented")
}

// Subscribe is not implemented.
func (conn *MapConn) Subscribe(recordEventChan chan oddb.RecordEvent) error {
	panic("not implemented")
}

// Close does nothing.
func (conn *MapConn) Close() error {
	// do nothing
	return nil
}

// RecordMap is a string=>Record map
type RecordMap map[string]oddb.Record

// SubscriptionMap is a string=>Subscription map
type SubscriptionMap map[string]oddb.Subscription

// MapDB is a naive memory implementation of oddb.Database.
type MapDB struct {
	RecordMap       RecordMap
	SubscriptionMap SubscriptionMap
	oddb.Database
}

// NewMapDB returns a new MapDB ready for use.
func NewMapDB() *MapDB {
	return &MapDB{
		RecordMap:       RecordMap{},
		SubscriptionMap: SubscriptionMap{},
	}
}

// ID returns a mock Database ID.
func (db *MapDB) ID() string {
	return "map-db"
}

// Get returns a Record from RecordMap.
func (db *MapDB) Get(id oddb.RecordID, record *oddb.Record) error {
	r, ok := db.RecordMap[id.String()]
	if !ok {
		return oddb.ErrRecordNotFound
	}
	*record = r
	return nil

}

// Save assigns Record to RecordMap.
func (db *MapDB) Save(record *oddb.Record) error {
	db.RecordMap[record.ID.String()] = *record
	return nil
}

// Delete remove the specified key from RecordMap.
func (db *MapDB) Delete(id oddb.RecordID) error {
	_, ok := db.RecordMap[id.String()]
	if !ok {
		return oddb.ErrRecordNotFound
	}
	delete(db.RecordMap, id.String())
	return nil
}

// Query is not implemented.
func (db *MapDB) Query(query *oddb.Query) (*oddb.Rows, error) {
	panic("oddbtest: MapDB.Query not supported")
}

// Extend does nothing.
func (db *MapDB) Extend(recordType string, schema oddb.RecordSchema) error {
	// do nothing
	return nil
}

// GetSubscription return a Subscription from SubscriptionMap.
func (db *MapDB) GetSubscription(name string, deviceID string, subscription *oddb.Subscription) error {
	s, ok := db.SubscriptionMap[deviceID+"/"+name]
	if !ok {
		return oddb.ErrSubscriptionNotFound
	}
	*subscription = s
	return nil
}

// SaveSubscription assigns to SubscriptionMap.
func (db *MapDB) SaveSubscription(subscription *oddb.Subscription) error {
	db.SubscriptionMap[subscription.DeviceID+"/"+subscription.ID] = *subscription
	return nil
}

// DeleteSubscription deletes the specified key from SubscriptionMap.
func (db *MapDB) DeleteSubscription(name string, deviceID string) error {
	key := deviceID + "/" + name
	_, ok := db.SubscriptionMap[key]
	if !ok {
		return oddb.ErrSubscriptionNotFound
	}
	delete(db.SubscriptionMap, key)
	return nil
}

var (
	_ oddb.Conn     = NewMapConn()
	_ oddb.Database = NewMapDB()
)
