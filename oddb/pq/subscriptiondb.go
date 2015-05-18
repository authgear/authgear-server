package pq

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/lib/pq"

	"github.com/oursky/ourd/oddb"
)

func isDeviceNotFound(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23503" && pqErr.Constraint == "_subscription_device_id_fkey"
	}

	return false
}

type nullNotificationInfo struct {
	NotificationInfo oddb.NotificationInfo
	Valid            bool
}

func (ni nullNotificationInfo) Value() (driver.Value, error) {
	if !ni.Valid {
		return nil, nil
	}
	return json.Marshal(ni.NotificationInfo)
}

func (ni *nullNotificationInfo) Scan(value interface{}) error {
	if value == nil {
		ni.NotificationInfo, ni.Valid = oddb.NotificationInfo{}, false
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		fmt.Errorf("oddb: unsupported Scan pair: %T -> %T", value, ni.NotificationInfo)
	}

	if err := json.Unmarshal(b, &ni.NotificationInfo); err != nil {
		return err
	}

	ni.Valid = true
	return nil
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

func (db *database) GetSubscription(key string, deviceID string, subscription *oddb.Subscription) error {
	nullinfo := nullNotificationInfo{}
	err := psql.Select("type", "notification_info", "query").
		From(db.tableName("_subscription")).
		Where("user_id = ? AND device_id = ? AND id = ?", db.userID, deviceID, key).
		RunWith(db.Db.DB).
		QueryRow().
		Scan(
		&subscription.Type,
		&nullinfo,
		(*queryValue)(&subscription.Query))

	if err == sql.ErrNoRows {
		return oddb.ErrSubscriptionNotFound
	} else if err != nil {
		return err
	}

	if nullinfo.Valid {
		subscription.NotificationInfo = &nullinfo.NotificationInfo
	} else {
		subscription.NotificationInfo = nil
	}
	subscription.DeviceID = deviceID
	subscription.ID = key

	return nil
}

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

	nullinfo := nullNotificationInfo{}
	if subscription.NotificationInfo != nil {
		nullinfo.NotificationInfo, nullinfo.Valid = *subscription.NotificationInfo, true
	}

	pkData := map[string]interface{}{
		"id":        subscription.ID,
		"user_id":   db.userID,
		"device_id": subscription.DeviceID,
	}

	data := map[string]interface{}{
		"type":              subscription.Type,
		"notification_info": nullinfo,
		"query":             queryValue(subscription.Query),
	}

	sql, args := upsertQuery(db.tableName("_subscription"), pkData, data, []string{})
	_, err := db.Db.Exec(sql, args...)

	if isDeviceNotFound(err) {
		return oddb.ErrDeviceNotFound
	}

	return err
}

func (db *database) DeleteSubscription(key string, deviceID string) error {
	result, err := psql.Delete(db.tableName("_subscription")).
		Where("user_id = ? AND device_id = ? AND id = ?", db.userID, deviceID, key).
		RunWith(db.Db.DB).
		Exec()

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return oddb.ErrSubscriptionNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (db *database) GetSubscriptionsByDeviceID(deviceID string) (subscriptions []oddb.Subscription) {
	rows, err := psql.Select("id", "type", "notification_info", "query").
		From(db.tableName("_subscription")).
		Where(`user_id = ? AND device_id = ?`, db.userID, deviceID).
		RunWith(db.Db.DB).
		Query()

	if err != nil {
		log.WithFields(log.Fields{
			"user_id":  db.userID,
			"deviceID": deviceID,
			"err":      err,
		}).Errorln("failed to query subscriptions by device id")

		return nil
	}

	subscriptions = []oddb.Subscription{}
	var s oddb.Subscription
	for rows.Next() {
		var nullinfo nullNotificationInfo
		err := rows.Scan(&s.ID, &s.Type, &nullinfo, (*queryValue)(&s.Query))
		if err != nil {
			log.WithFields(log.Fields{
				"userID":   db.userID,
				"deviceID": deviceID,
				"err":      err,
			}).Errorln("failed to scan a subscription row by device id, skipping...")

			continue
		}

		if nullinfo.Valid {
			s.NotificationInfo = &nullinfo.NotificationInfo
		} else {
			s.NotificationInfo = nil
		}
		s.DeviceID = deviceID

		subscriptions = append(subscriptions, s)
	}

	if rows.Err() != nil {
		log.WithFields(log.Fields{
			"userID":   db.userID,
			"deviceID": deviceID,
			"err":      rows.Err(),
		}).Errorln("failed to scan subscriptions by device id")

		return nil
	}

	log.Debug(subscriptions)
	return subscriptions
}

func (db *database) GetMatchingSubscriptions(record *oddb.Record) (subscriptions []oddb.Subscription) {
	sql, args, err := psql.Select("id", "device_id", "type", "notification_info", "query").
		From(db.tableName("_subscription")).
		Where(`user_id = ? AND query @> ?::jsonb`, db.userID, fmt.Sprintf(`{"record_type":"%s"}`, record.ID.Type)).
		ToSql()

	if err != nil {
		panic(err)
	}

	rows, err := db.Db.Query(sql, args...)
	if err != nil {
		log.WithFields(log.Fields{
			"sql":    sql,
			"args":   args,
			"record": record,
			"userID": db.userID,
			"err":    err,
		}).Errorln("failed to select subscriptions")

		return nil
	}

	var s oddb.Subscription
	for rows.Next() {
		var nullinfo nullNotificationInfo
		err := rows.Scan(&s.ID, &s.DeviceID, &s.Type, &nullinfo, (*queryValue)(&s.Query))
		if err != nil {
			log.WithField("err", err).Errorln("failed to scan a subscription row, skipping...")
			continue
		}

		if nullinfo.Valid {
			s.NotificationInfo = &nullinfo.NotificationInfo
		} else {
			s.NotificationInfo = nil
		}

		subscriptions = append(subscriptions, s)
	}

	if rows.Err() != nil {
		log.WithFields(log.Fields{
			"record": record,
			"userID": db.userID,
			"err":    rows.Err(),
		}).Errorln("failed to scan matching subscriptions")

		return nil
	}

	return subscriptions
}
