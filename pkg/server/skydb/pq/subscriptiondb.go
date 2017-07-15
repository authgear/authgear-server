// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pq

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/pq/builder"
)

func isDeviceNotFound(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23503" && pqErr.Constraint == "_subscription_device_id_fkey"
	}

	return false
}

type nullNotificationInfo struct {
	NotificationInfo skydb.NotificationInfo
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
		ni.NotificationInfo, ni.Valid = skydb.NotificationInfo{}, false
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		fmt.Errorf("skydb: unsupported Scan pair: %T -> %T", value, ni.NotificationInfo)
	}

	if err := json.Unmarshal(b, &ni.NotificationInfo); err != nil {
		return err
	}

	ni.Valid = true
	return nil
}

type queryValue skydb.Query

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
		fmt.Errorf("skydb: unsupported Scan pair: %T -> %T", value, query)
	}

	v := struct {
		Type         string
		Predicate    jsonPredicate
		Sorts        []skydb.Sort
		ComputedKeys map[string]skydb.Expression
		DesiredKeys  []string
		Limit        *uint64
		Offset       uint64
	}{}

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	query.Type = v.Type
	query.Predicate = skydb.Predicate(v.Predicate)
	query.Sorts = v.Sorts
	query.ComputedKeys = v.ComputedKeys
	query.DesiredKeys = v.DesiredKeys
	query.Limit = v.Limit
	query.Offset = v.Offset

	return nil
}

type jsonPredicate skydb.Predicate

func (p *jsonPredicate) UnmarshalJSON(data []byte) error {
	v := struct {
		Operator skydb.Operator
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
			p.Children = append(p.Children, skydb.Predicate(pred))
		}
	} else {
		expressions := []skydb.Expression{}
		if err := json.Unmarshal(v.Children, &expressions); err != nil {
			return err
		}
		for _, expr := range expressions {
			p.Children = append(p.Children, expr)
		}
	}

	return nil
}

func (db *database) GetSubscription(key string, deviceID string, subscription *skydb.Subscription) error {
	if db.DatabaseType() == skydb.UnionDatabase {
		return errors.New("union database does not implement subscription")
	}
	nullinfo := nullNotificationInfo{}

	builder := psql.Select("type", "notification_info", "query").
		From(db.TableName("_subscription")).
		Where("user_id = ? AND device_id = ? AND id = ?", db.userID, deviceID, key)
	err := db.c.QueryRowWith(builder).
		Scan(&subscription.Type, &nullinfo, (*queryValue)(&subscription.Query))

	if err == sql.ErrNoRows {
		return skydb.ErrSubscriptionNotFound
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

func (db *database) SaveSubscription(subscription *skydb.Subscription) error {
	if db.DatabaseType() == skydb.UnionDatabase {
		return errors.New("union database does not implement subscription")
	}
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

	builder := builder.UpsertQuery(db.TableName("_subscription"), pkData, data)

	_, err := db.c.ExecWith(builder)
	if isDeviceNotFound(err) {
		return skydb.ErrDeviceNotFound
	}

	return err
}

func (db *database) DeleteSubscription(key string, deviceID string) error {
	if db.DatabaseType() == skydb.UnionDatabase {
		return errors.New("union database does not implement subscription")
	}
	result, err := db.c.ExecWith(
		psql.Delete(db.TableName("_subscription")).
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
		return skydb.ErrSubscriptionNotFound
	} else if rowsAffected > 1 {
		panic(fmt.Errorf("want 1 rows updated, got %v", rowsAffected))
	}

	return nil
}

func (db *database) GetSubscriptionsByDeviceID(deviceID string) (subscriptions []skydb.Subscription) {
	if db.DatabaseType() == skydb.UnionDatabase {
		log.WithFields(logrus.Fields{
			"user_id":  db.userID,
			"deviceID": deviceID,
		}).Errorln("GetSubscriptionsByDeviceID on union database is not implemented")
		return nil
	}
	rows, err := db.c.QueryWith(
		psql.Select("id", "type", "notification_info", "query").
			From(db.TableName("_subscription")).
			Where(`user_id = ? AND device_id = ?`, db.userID, deviceID),
	)

	if err != nil {
		log.WithFields(logrus.Fields{
			"user_id":  db.userID,
			"deviceID": deviceID,
			"err":      err,
		}).Errorln("failed to query subscriptions by device id")

		return nil
	}

	subscriptions = []skydb.Subscription{}
	var s skydb.Subscription
	for rows.Next() {
		var nullinfo nullNotificationInfo
		err := rows.Scan(&s.ID, &s.Type, &nullinfo, (*queryValue)(&s.Query))
		if err != nil {
			log.WithFields(logrus.Fields{
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
		log.WithFields(logrus.Fields{
			"userID":   db.userID,
			"deviceID": deviceID,
			"err":      rows.Err(),
		}).Errorln("failed to scan subscriptions by device id")

		return nil
	}

	log.Debug(subscriptions)
	return subscriptions
}

func (db *database) GetMatchingSubscriptions(record *skydb.Record) (subscriptions []skydb.Subscription) {
	if db.DatabaseType() == skydb.UnionDatabase {
		log.WithFields(logrus.Fields{
			"user_id": db.userID,
		}).Errorln("GetMatchingSubscriptions on union database is not implemented")
		return nil
	}
	builder := psql.Select("id", "device_id", "type", "notification_info", "query").
		From(db.TableName("_subscription")).
		Where(`user_id = ? AND query @> ?::jsonb`, db.userID, fmt.Sprintf(`{"Type":"%s"}`, record.ID.Type))

	rows, err := db.c.QueryWith(builder)
	if err != nil {
		log.WithFields(logrus.Fields{
			"record": record,
			"userID": db.userID,
			"err":    err,
		}).Errorln("failed to select subscriptions")

		return nil
	}

	var s skydb.Subscription
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
		if predMatchRecord(&(subscription.Query.Predicate), record) {
			matchingSubs = append(matchingSubs, subscription)
		}
	}

	if rows.Err() != nil {
		log.WithFields(logrus.Fields{
			"record": record,
			"userID": db.userID,
			"err":    rows.Err(),
		}).Errorln("failed to scan matching subscriptions")

		return nil
	}

	return matchingSubs
}

func predMatchRecord(p *skydb.Predicate, record *skydb.Record) (b bool) {
	if p == nil || p.IsEmpty() {
		return true
	}

	switch p.Operator {
	case skydb.And:
		b = true
		for _, childPred := range p.GetSubPredicates() {
			if !predMatchRecord(&childPred, record) {
				b = false
				break
			}
		}
	case skydb.Or:
		for _, childPred := range p.GetSubPredicates() {
			if predMatchRecord(&childPred, record) {
				b = true
				break
			}
		}
	case skydb.Not:
		b = !predMatchRecord(&p.GetSubPredicates()[0], record)
	case skydb.Equal:
		lv, rv := extractBinaryOperands(p.GetExpressions(), record)
		return reflect.DeepEqual(lv, rv)
	// case skydb.GreaterThan:
	// case skydb.LessThan:
	// case skydb.GreaterThanOrEqual:
	// case skydb.LessThanOrEqual:
	case skydb.NotEqual:
		lv, rv := extractBinaryOperands(p.GetExpressions(), record)
		return !reflect.DeepEqual(lv, rv)
	case skydb.In:
		lv, rv := extractBinaryOperands(p.GetExpressions(), record)
		haystack, ok := rv.([]interface{})
		if !ok {
			log.Panicf("unknown value in right hand side of `In` operand = %v", rv)
		}

		return deepEqualIn(lv, haystack)
	// case skydb.Like:
	// case skydb.ILike:
	default:
		log.Panicf("unknown Predicate.Operator = %v", p.Operator)
	}

	return
}

func extractBinaryOperands(exprs []skydb.Expression, record *skydb.Record) (lv interface{}, rv interface{}) {
	lv = extractValue(exprs[0], record)
	rv = extractValue(exprs[1], record)
	return
}

func extractValue(expr skydb.Expression, record *skydb.Record) interface{} {
	switch expr.Type {
	case skydb.Literal:
		switch expr.Value.(type) {
		case bool, float64, string, time.Time, *skydb.Location, skydb.Reference, []interface{}:
			return expr.Value
		default:
			panic(fmt.Sprintf("unknown type %[1]T of Expression.Value = %[1]v", expr.Value))
		}
	case skydb.KeyPath:
		return record.Get(expr.Value.(string))
	case skydb.Function:
		panic("unsupported type of predicate expression = Function")
	}

	panic("unreachable code")
}

func deepEqualIn(needle interface{}, haystack []interface{}) bool {
	for _, hay := range haystack {
		if reflect.DeepEqual(needle, hay) {
			return true
		}
	}
	return false
}
