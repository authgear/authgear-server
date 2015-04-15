package oddbtest

import (
	"github.com/oursky/ourd/oddb"
)

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

// NewMapDB returns new new MapDB ready for use.
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
func (db *MapDB) Get(key string, record *oddb.Record) error {
	r, ok := db.RecordMap[key]
	if !ok {
		return oddb.ErrRecordNotFound
	}
	*record = r
	return nil

}

// Save assigns Record to RecordMap.
func (db *MapDB) Save(record *oddb.Record) error {
	db.RecordMap[record.Key] = *record
	return nil
}

// Delete remove the specified key from RecordMap.
func (db *MapDB) Delete(key string) error {
	_, ok := db.RecordMap[key]
	if !ok {
		return oddb.ErrRecordNotFound
	}
	delete(db.RecordMap, key)
	return nil
}

// Query is not implemented.
func (db *MapDB) Query(query *oddb.Query) (*oddb.Rows, error) {
	panic("oddbtest: MapDB.Query not supported")
}

// GetSubscription return a Subscription from SubscriptionMap.
func (db *MapDB) GetSubscription(key string, subscription *oddb.Subscription) error {
	s, ok := db.SubscriptionMap[key]
	if !ok {
		return oddb.ErrRecordNotFound
	}
	*subscription = s
	return nil
}

// SaveSubscription assigns to SubscriptionMap.
func (db *MapDB) SaveSubscription(subscription *oddb.Subscription) error {
	db.SubscriptionMap[subscription.ID] = *subscription
	return nil
}

// DeleteSubscription deletes the specified key from SubscriptionMap.
func (db *MapDB) DeleteSubscription(key string) error {
	_, ok := db.SubscriptionMap[key]
	if !ok {
		return oddb.ErrRecordNotFound
	}
	delete(db.SubscriptionMap, key)
	return nil
}
