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
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func addDevice(t *testing.T, c *conn, userID string, deviceID string) {
	_, err := c.Exec("INSERT INTO app_com_oursky_skygear._device (id, user_id, type, token, last_registered_at) VALUES ($1, $2, '', $3, $4)", deviceID, userID, randHex(64), time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
}

func randHex(n int) string {
	const hexStr = "0123456789abcef"
	buf := bytes.Buffer{}
	buf.Grow(n)

	for i := 0; i < n; i++ {
		buf.WriteByte(hexStr[rand.Intn(len(hexStr))])
	}

	return buf.String()
}

func TestSubscriptionCRUD(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PrivateDB("userid")

		// fixture
		addUser(t, c, "userid")
		addDevice(t, c, "userid", "deviceid")

		notificationInfo := skydb.NotificationInfo{
			APS: skydb.APSSetting{
				Alert: &skydb.AppleAlert{
					Body:                  "somebody",
					LocalizationKey:       "somelocalizationkey",
					LocalizationArgs:      []string{"arg0", "arg1"},
					LaunchImage:           "somelaunchimage",
					ActionLocalizationKey: "someactionlocalizationkey",
				},
				SoundName:                  "somesoundname",
				ShouldBadge:                true,
				ShouldSendContentAvailable: true,
			},
		}
		query := skydb.Query{
			Type: "recordtype",
			Predicate: skydb.Predicate{
				Operator: skydb.Equal,
				Children: []interface{}{
					skydb.Expression{
						Type:  skydb.KeyPath,
						Value: "_id",
					},
					skydb.Expression{
						Type:  skydb.Literal,
						Value: "RECORD_ID",
					},
				},
			},
		}
		subscription := skydb.Subscription{
			ID:               "subscriptionid",
			Type:             "query",
			DeviceID:         "deviceid",
			NotificationInfo: &notificationInfo,
			Query:            query,
		}

		Convey("get an existing subscription", func() {
			So(db.SaveSubscription(&subscription), ShouldBeNil)

			resultSubscription := skydb.Subscription{}
			err := db.GetSubscription("subscriptionid", "deviceid", &resultSubscription)
			So(err, ShouldBeNil)
			So(subscription, ShouldResemble, resultSubscription)
		})

		Convey("returns ErrSubscriptionNotFound while trying to get a non-existing subscription ", func() {
			resultSubscription := skydb.Subscription{}
			err := db.GetSubscription("notexistsubscriptionid", "deviceid", &resultSubscription)
			So(err, ShouldEqual, skydb.ErrSubscriptionNotFound)
		})

		Convey("create new subscription", func() {
			err := db.SaveSubscription(&subscription)
			So(err, ShouldBeNil)

			var (
				deviceID, queryType    string
				resultNotificationInfo nullNotificationInfo
				resultQuery            skydb.Query
			)
			err = c.QueryRowx(`
				SELECT device_id, type, notification_info, query FROM app_com_oursky_skygear._subscription
				WHERE id = $1 AND user_id = $2`, "subscriptionid", "userid").
				Scan(&deviceID, &queryType, &resultNotificationInfo, (*queryValue)(&resultQuery))
			So(err, ShouldBeNil)

			So(deviceID, ShouldEqual, "deviceid")
			So(queryType, ShouldEqual, "query")
			So(resultNotificationInfo.NotificationInfo, ShouldResemble, notificationInfo)
			So(resultQuery, ShouldResemble, query)
		})

		Convey("modify existing subscription", func() {
			So(db.SaveSubscription(&subscription), ShouldBeNil)

			subscription.Query.Type = "otherrecordtype"
			err := db.SaveSubscription(&subscription)
			So(err, ShouldBeNil)

			var (
				deviceID, queryType    string
				resultNotificationInfo nullNotificationInfo
				resultQuery            skydb.Query
			)
			err = c.QueryRowx(`
				SELECT device_id, type, notification_info, query FROM app_com_oursky_skygear._subscription
				WHERE id = $1 AND user_id = $2`, "subscriptionid", "userid").
				Scan(&deviceID, &queryType, &resultNotificationInfo, (*queryValue)(&resultQuery))
			So(err, ShouldBeNil)

			So(deviceID, ShouldEqual, "deviceid")
			So(queryType, ShouldEqual, "query")
			So(resultNotificationInfo.NotificationInfo, ShouldResemble, notificationInfo)

			query.Type = "otherrecordtype"
			So(resultQuery, ShouldResemble, query)
		})

		Convey("save subscription with the same name in the same database with dfference device id", func() {
			addDevice(t, c, "userid", "device0")
			addDevice(t, c, "userid", "device1")

			subscription.DeviceID = "device0"
			So(db.SaveSubscription(&subscription), ShouldBeNil)

			subscription.DeviceID = "device1"
			So(db.SaveSubscription(&subscription), ShouldBeNil)
		})

		Convey("cannot save subscription with empty id", func() {
			subscription.ID = ""
			err := db.SaveSubscription(&subscription)
			So(err.Error(), ShouldEqual, "empty id")
		})

		Convey("cannot save subscription with empty type", func() {
			subscription.Type = ""
			err := db.SaveSubscription(&subscription)
			So(err.Error(), ShouldEqual, "empty type")
		})

		Convey("cannot save subscription with empty query type", func() {
			subscription.Query.Type = ""
			err := db.SaveSubscription(&subscription)
			So(err.Error(), ShouldEqual, "empty query type")
		})

		Convey("cannot save subscription with empty device id", func() {
			subscription.DeviceID = ""
			err := db.SaveSubscription(&subscription)
			So(err.Error(), ShouldEqual, "empty device id")
		})

		Convey("returns ErrDeviceNotFound if a subscription with non-exist device is saved", func() {
			subscription.DeviceID = "notexistdeviceid"
			err := db.SaveSubscription(&subscription)
			So(err, ShouldEqual, skydb.ErrDeviceNotFound)
		})

		Convey("delets an existing subscription", func() {
			So(db.SaveSubscription(&subscription), ShouldBeNil)

			err := db.DeleteSubscription("subscriptionid", "deviceid")
			So(err, ShouldBeNil)

			var count int
			err = c.QueryRowx(
				`SELECT COUNT(*) FROM app_com_oursky_skygear._subscription
				WHERE id = $1 AND user_id = $2`,
				"subscriptionid", "userid").
				Scan(&count)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
		})

		Convey("returns ErrSubscriptionNotFound while deleting a non-exist subscription", func() {
			err := db.DeleteSubscription("notexistsubscriptionid", "deviceid")
			So(err, ShouldEqual, skydb.ErrSubscriptionNotFound)
		})
	})
}

func TestMatchingSubscriptions(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()

		// fixture
		addUser(t, c, "userid")
		addDevice(t, c, "userid", "device0")
		addDevice(t, c, "userid", "device1")

		sub00 := subscriptionForTest("device0", "00", "type0")
		sub01 := subscriptionForTest("device0", "01", "type0")
		sub10 := subscriptionForTest("device1", "10", "type0")
		sub11 := subscriptionForTest("device1", "11", "type1")

		So(db.SaveSubscription(&sub00), ShouldBeNil)
		So(db.SaveSubscription(&sub01), ShouldBeNil)
		So(db.SaveSubscription(&sub10), ShouldBeNil)
		So(db.SaveSubscription(&sub11), ShouldBeNil)

		Convey("fetch matching subscription for a record", func() {
			record := skydb.Record{ID: skydb.NewRecordID("type0", "recordid")}
			subscriptions := db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []skydb.Subscription{sub00, sub01, sub10})
		})

		Convey("fetch no subscription for a not matching record", func() {
			record := skydb.Record{ID: skydb.NewRecordID("notexisttype", "recordid")}
			subscriptions := db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldBeEmpty)
		})

		Convey("match subscription with predicate eq", func() {
			record := skydb.Record{ID: skydb.NewRecordID("record", "id")}
			subeq := subscriptionForTest("device0", "eq", "record")
			subeq.Query.Predicate = skydb.Predicate{
				Operator: skydb.Equal,
				Children: []interface{}{
					skydb.Expression{
						Type:  skydb.Literal,
						Value: "id",
					},
					skydb.Expression{
						Type:  skydb.KeyPath,
						Value: "_id",
					},
				},
			}
			So(db.SaveSubscription(&subeq), ShouldBeNil)

			subscriptions := db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []skydb.Subscription{subeq})
		})

		Convey("match subscription with compound predicates", func() {
			binaryPred := func(op skydb.Operator, k string, v interface{}) skydb.Predicate {
				return skydb.Predicate{
					Operator: skydb.Equal,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: k,
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: v,
						},
					},
				}
			}

			record := skydb.Record{ID: skydb.NewRecordID("record", "id")}

			suband := subscriptionForTest("device0", "and", "record")
			suband.Query.Predicate = skydb.Predicate{
				Operator: skydb.And,
				Children: []interface{}{
					binaryPred(skydb.Equal, "falsy", false),
					binaryPred(skydb.Equal, "truthy", true),
				},
			}
			So(db.SaveSubscription(&suband), ShouldBeNil)

			subor := subscriptionForTest("device0", "or", "record")
			subor.Query.Predicate = skydb.Predicate{
				Operator: skydb.Or,
				Children: []interface{}{
					binaryPred(skydb.Equal, "truthy", false),
					binaryPred(skydb.Equal, "falsy", false),
				},
			}
			So(db.SaveSubscription(&subor), ShouldBeNil)

			subnot := subscriptionForTest("device0", "not", "record")
			subnot.Query.Predicate = skydb.Predicate{
				Operator: skydb.Not,
				Children: []interface{}{
					binaryPred(skydb.Equal, "truthy", false),
				},
			}
			So(db.SaveSubscription(&subnot), ShouldBeNil)

			record.Data = map[string]interface{}{
				"truthy": true,
				"falsy":  false,
			}
			subscriptions := db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []skydb.Subscription{suband, subor, subnot})

			record.Data = map[string]interface{}{
				"truthy": false,
				"falsy":  true,
			}
			subscriptions = db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []skydb.Subscription{subor})

			record.Data = map[string]interface{}{
				"truthy": true,
				"falsy":  true,
			}
			subscriptions = db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []skydb.Subscription{subnot})

			record.Data = map[string]interface{}{
				"truthy": false,
				"falsy":  false,
			}
			subscriptions = db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []skydb.Subscription{subor})
		})
	})
}

func TestGetSubscriptionsByDeviceID(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		db := c.PublicDB()

		// fixture
		addUser(t, c, "userid")
		addDevice(t, c, "userid", "device0")
		addDevice(t, c, "userid", "device1")
		addDevice(t, c, "userid", "device2")

		sub00 := subscriptionForTest("device0", "00", "type0")
		sub01 := subscriptionForTest("device0", "01", "type1")
		sub10 := subscriptionForTest("device1", "10", "type0")

		So(db.SaveSubscription(&sub00), ShouldBeNil)
		So(db.SaveSubscription(&sub01), ShouldBeNil)
		So(db.SaveSubscription(&sub10), ShouldBeNil)

		Convey("fetches subscriptions by device_id", func() {
			subscriptions := db.GetSubscriptionsByDeviceID("device0")
			So(subscriptions, ShouldResemble, []skydb.Subscription{
				sub00,
				sub01,
			})
		})

		Convey("fetches no subscriptions by device_id", func() {
			subscriptions := db.GetSubscriptionsByDeviceID("device2")
			So(subscriptions, ShouldBeEmpty)
		})

		Convey("fetches no subscriptions by non-exist device_id", func() {
			subscriptions := db.GetSubscriptionsByDeviceID("notexistdeviceid")
			So(subscriptions, ShouldBeEmpty)
		})
	})
}

func subscriptionForTest(deviceID, id, queryRecordType string) skydb.Subscription {
	return skydb.Subscription{
		ID:       id,
		Type:     "query",
		DeviceID: deviceID,
		Query: skydb.Query{
			Type: queryRecordType,
		},
	}
}

func TestPredicateMatchRecord(t *testing.T) {
	Convey("Records", t, func() {
		record1 := skydb.Record{ID: skydb.NewRecordID("record", "id")}
		record1.Data = map[string]interface{}{
			"category": "recipe",
		}
		Convey("Match record with predicate in", func() {

			predicate := skydb.Predicate{
				Operator: skydb.In,
				Children: []interface{}{
					skydb.Expression{
						Type:  skydb.KeyPath,
						Value: "category",
					},
					skydb.Expression{
						Type:  skydb.Literal,
						Value: []interface{}{"recipe", "fiction"},
					},
				},
			}

			So(predMatchRecord(&predicate, &record1), ShouldBeTrue)
		})

		Convey("Not match record with predicate in", func() {
			predicate := skydb.Predicate{
				Operator: skydb.In,
				Children: []interface{}{
					skydb.Expression{
						Type:  skydb.KeyPath,
						Value: "category",
					},
					skydb.Expression{
						Type:  skydb.Literal,
						Value: []interface{}{"utility", "fiction"},
					},
				},
			}

			So(predMatchRecord(&predicate, &record1), ShouldBeFalse)
		})
	})
}
