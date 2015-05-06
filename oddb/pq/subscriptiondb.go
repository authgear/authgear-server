package pq

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/oursky/ourd/oddb"
)

func isDeviceNotFound(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23503" && pqErr.Constraint == "_subscription_device_id_fkey"
	}

	return false
}

type notificationInfoValue oddb.NotificationInfo

func (info notificationInfoValue) Value() (driver.Value, error) {
	return json.Marshal(info)
}

func (info *notificationInfoValue) Scan(value interface{}) error {
	if value == nil {
		*info = notificationInfoValue{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		fmt.Errorf("oddb: unsupported Scan pair: %T -> %T", value, info)
	}

	return json.Unmarshal(b, info)
}

type queryValue oddb.Query

func (query queryValue) Value() (driver.Value, error) {
	return json.Marshal(query)
}

func (query *queryValue) Scan(value interface{}) error {
	if value == nil {
		*query = queryValue{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		fmt.Errorf("oddb: unsupported Scan pair: %T -> %T", value, query)
	}

	return json.Unmarshal(b, query)
}

func (db *database) GetSubscription(key string, subscription *oddb.Subscription) error { return nil }

func (db *database) SaveSubscription(subscription *oddb.Subscription) error {
	if subscription.ID == "" {
		return errors.New("empty id")
	}
	if subscription.Type == "" {
		return errors.New("empty type")
	}
	if subscription.Query.Type == "" {
		return errors.New("empty query type")
	}
	if subscription.DeviceID == "" {
		return errors.New("empty device id")
	}

	pkData := map[string]interface{}{
		"id":      subscription.ID,
		"user_id": db.userID,
	}

	data := map[string]interface{}{
		"device_id":         subscription.DeviceID,
		"type":              subscription.Type,
		"notification_info": notificationInfoValue(subscription.NotificationInfo),
		"query":             queryValue(subscription.Query),
	}

	sql, args := upsertQuery(db.tableName("_subscription"), pkData, data)
	_, err := db.Db.Exec(sql, args...)

	if isDeviceNotFound(err) {
		return oddb.ErrDeviceNotFound
	}

	return err
}

func (db *database) DeleteSubscription(key string) error { return nil }

func (db *database) GetMatchingSubscription(record *oddb.Record) []oddb.Subscription { return nil }
