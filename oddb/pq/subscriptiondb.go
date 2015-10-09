package pq

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

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

	v := struct {
		Type         string
		Predicate    *jsonPredicate
		Sorts        []oddb.Sort
		ReadableBy   string
		ComputedKeys map[string]oddb.Expression
		DesiredKeys  []string
		Limit        uint64
		Offset       uint64
	}{}

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	query.Type = v.Type
	query.Predicate = (*oddb.Predicate)(v.Predicate)
	query.Sorts = v.Sorts
	query.ReadableBy = v.ReadableBy
	query.ComputedKeys = v.ComputedKeys
	query.DesiredKeys = v.DesiredKeys
	query.Limit = v.Limit
	query.Offset = v.Offset

	return nil
}

type jsonPredicate oddb.Predicate

func (p *jsonPredicate) UnmarshalJSON(data []byte) error {
	v := struct {
		Operator oddb.Operator
		Children json.RawMessage
	}{}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	p.Operator = v.Operator

	if v.Operator.IsCompound() {
		predicates := []jsonPredicate{}
		if err := json.Unmarshal(v.Children, &predicates); err != nil {
			return err
		}
		for _, pred := range predicates {
			p.Children = append(p.Children, oddb.Predicate(pred))
		}
	} else {
		expressions := []oddb.Expression{}
		if err := json.Unmarshal(v.Children, &expressions); err != nil {
			return err
		}
		for _, expr := range expressions {
			p.Children = append(p.Children, expr)
		}
	}

	return nil
}

func (db *database) GetSubscription(key string, deviceID string, subscription *oddb.Subscription) error {
	nullinfo := nullNotificationInfo{}

	builder := psql.Select("type", "notification_info", "query").
		From(db.tableName("_subscription")).
		Where("user_id = ? AND device_id = ? AND id = ?", db.userID, deviceID, key)
	err := queryRowWith(db.Db, builder).
		Scan(&subscription.Type, &nullinfo, (*queryValue)(&subscription.Query))

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

	builder := upsertQuery(db.tableName("_subscription"), pkData, data)

	_, err := execWith(db.Db, builder)
	if isDeviceNotFound(err) {
		return oddb.ErrDeviceNotFound
	} else if err != nil {
		sql, args, _ := builder.ToSql()
		log.WithFields(log.Fields{
			"sql":          sql,
			"args":         args,
			"err":          err,
			"subscription": subscription,
		}).Errorln("Failed to save subscription")
	}

	return err
}

func (db *database) DeleteSubscription(key string, deviceID string) error {
	result, err := execWith(
		db.Db,
		psql.Delete(db.tableName("_subscription")).
			Where("user_id = ? AND device_id = ? AND id = ?", db.userID, deviceID, key),
	)

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
	rows, err := queryWith(
		db.Db,
		psql.Select("id", "type", "notification_info", "query").
			From(db.tableName("_subscription")).
			Where(`user_id = ? AND device_id = ?`, db.userID, deviceID),
	)

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
	builder := psql.Select("id", "device_id", "type", "notification_info", "query").
		From(db.tableName("_subscription")).
		Where(`user_id = ? AND query @> ?::jsonb`, db.userID, fmt.Sprintf(`{"Type":"%s"}`, record.ID.Type))

	rows, err := queryWith(db.Db, builder)
	if err != nil {
		sql, args, _ := builder.ToSql()
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

	// filter without allocation
	matchingSubs := subscriptions[:0]
	for _, subscription := range subscriptions {
		if predMatchRecord(subscription.Query.Predicate, record) {
			matchingSubs = append(matchingSubs, subscription)
		}
	}

	if rows.Err() != nil {
		log.WithFields(log.Fields{
			"record": record,
			"userID": db.userID,
			"err":    rows.Err(),
		}).Errorln("failed to scan matching subscriptions")

		return nil
	}

	return matchingSubs
}

func predMatchRecord(p *oddb.Predicate, record *oddb.Record) (b bool) {
	if p == nil {
		return true
	}

	switch p.Operator {
	case oddb.And:
		b = true
		for _, childPred := range p.GetSubPredicates() {
			if !predMatchRecord(&childPred, record) {
				b = false
				break
			}
		}
	case oddb.Or:
		for _, childPred := range p.GetSubPredicates() {
			if predMatchRecord(&childPred, record) {
				b = true
				break
			}
		}
	case oddb.Not:
		b = !predMatchRecord(&p.GetSubPredicates()[0], record)
	case oddb.Equal:
		lv, rv := extractBinaryOperands(p.GetExpressions(), record)
		return reflect.DeepEqual(lv, rv)
	// case oddb.GreaterThan:
	// case oddb.LessThan:
	// case oddb.GreaterThanOrEqual:
	// case oddb.LessThanOrEqual:
	case oddb.NotEqual:
		lv, rv := extractBinaryOperands(p.GetExpressions(), record)
		return !reflect.DeepEqual(lv, rv)
	// case oddb.Like:
	// case oddb.ILike:
	default:
		log.Panicf("unknown Predicate.Operator = %v", p.Operator)
	}

	return
}

func extractBinaryOperands(exprs []oddb.Expression, record *oddb.Record) (lv interface{}, rv interface{}) {
	lv = extractValue(exprs[0], record)
	rv = extractValue(exprs[1], record)
	return
}

func extractValue(expr oddb.Expression, record *oddb.Record) interface{} {
	switch expr.Type {
	case oddb.Literal:
		switch expr.Value.(type) {
		case bool, float64, string, time.Time, *oddb.Location, oddb.Reference:
			return expr.Value
		default:
			panic(fmt.Sprintf("unknown type %[1]T of Expression.Value = %[1]v", expr.Value))
		}
	case oddb.KeyPath:
		return record.Get(expr.Value.(string))
	case oddb.Function:
		panic("unsupported type of predicate expression = Function")
	}

	panic("unreachable code")
}
