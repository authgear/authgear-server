package pq

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/oursky/skygear/oddb"
	. "github.com/smartystreets/goconvey/convey"
)

func addDevice(t *testing.T, c *conn, userID string, deviceID string) {
	_, err := c.Db.Exec("INSERT INTO app_com_oursky_skygear._device (id, user_id, type, token, last_registered_at) VALUES ($1, $2, '', $3, $4)", deviceID, userID, randHex(64), time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
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
		defer cleanupDB(t, c.Db)

		db := c.PrivateDB("userid")

		// fixture
		addUser(t, c, "userid")
		addDevice(t, c, "userid", "deviceid")

		notificationInfo := oddb.NotificationInfo{
			APS: oddb.APSSetting{
				Alert: &oddb.AppleAlert{
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
		query := oddb.Query{
			Type: "recordtype",
			Predicate: &oddb.Predicate{
				Operator: oddb.Equal,
				Children: []interface{}{
					oddb.Expression{
						Type:  oddb.KeyPath,
						Value: "_id",
					},
					oddb.Expression{
						Type:  oddb.Literal,
						Value: "RECORD_ID",
					},
				},
			},
		}
		subscription := oddb.Subscription{
			ID:               "subscriptionid",
			Type:             "query",
			DeviceID:         "deviceid",
			NotificationInfo: &notificationInfo,
			Query:            query,
		}

		Convey("get an existing subscription", func() {
			So(db.SaveSubscription(&subscription), ShouldBeNil)

			resultSubscription := oddb.Subscription{}
			err := db.GetSubscription("subscriptionid", "deviceid", &resultSubscription)
			So(err, ShouldBeNil)
			So(subscription, ShouldResemble, resultSubscription)
		})

		Convey("returns ErrSubscriptionNotFound while trying to get a non-existing subscription ", func() {
			resultSubscription := oddb.Subscription{}
			err := db.GetSubscription("notexistsubscriptionid", "deviceid", &resultSubscription)
			So(err, ShouldEqual, oddb.ErrSubscriptionNotFound)
		})

		Convey("create new subscription", func() {
			err := db.SaveSubscription(&subscription)
			So(err, ShouldBeNil)

			var (
				deviceID, queryType    string
				resultNotificationInfo nullNotificationInfo
				resultQuery            oddb.Query
			)
			err = c.Db.QueryRow(`
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
				resultQuery            oddb.Query
			)
			err = c.Db.QueryRow(`
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
			So(err, ShouldEqual, oddb.ErrDeviceNotFound)
		})

		Convey("delets an existing subscription", func() {
			So(db.SaveSubscription(&subscription), ShouldBeNil)

			err := db.DeleteSubscription("subscriptionid", "deviceid")
			So(err, ShouldBeNil)

			var count int
			err = c.Db.QueryRow(
				`SELECT COUNT(*) FROM app_com_oursky_skygear._subscription
				WHERE id = $1 AND user_id = $2`,
				"subscriptionid", "userid").
				Scan(&count)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
		})

		Convey("returns ErrSubscriptionNotFound while deleting a non-exist subscription", func() {
			err := db.DeleteSubscription("notexistsubscriptionid", "deviceid")
			So(err, ShouldEqual, oddb.ErrSubscriptionNotFound)
		})
	})
}

func TestMatchingSubscriptions(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

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
			record := oddb.Record{ID: oddb.NewRecordID("type0", "recordid")}
			subscriptions := db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []oddb.Subscription{sub00, sub01, sub10})
		})

		Convey("fetch no subscription for a not matching record", func() {
			record := oddb.Record{ID: oddb.NewRecordID("notexisttype", "recordid")}
			subscriptions := db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldBeEmpty)
		})

		Convey("match subscription with predicate eq", func() {
			record := oddb.Record{ID: oddb.NewRecordID("record", "id")}
			subeq := subscriptionForTest("device0", "eq", "record")
			subeq.Query.Predicate = &oddb.Predicate{
				Operator: oddb.Equal,
				Children: []interface{}{
					oddb.Expression{
						Type:  oddb.Literal,
						Value: "id",
					},
					oddb.Expression{
						Type:  oddb.KeyPath,
						Value: "_id",
					},
				},
			}
			So(db.SaveSubscription(&subeq), ShouldBeNil)

			subscriptions := db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []oddb.Subscription{subeq})
		})

		Convey("match subscription with compound predicates", func() {
			binaryPred := func(op oddb.Operator, k string, v interface{}) oddb.Predicate {
				return oddb.Predicate{
					Operator: oddb.Equal,
					Children: []interface{}{
						oddb.Expression{
							Type:  oddb.KeyPath,
							Value: k,
						},
						oddb.Expression{
							Type:  oddb.Literal,
							Value: v,
						},
					},
				}
			}

			record := oddb.Record{ID: oddb.NewRecordID("record", "id")}

			suband := subscriptionForTest("device0", "and", "record")
			suband.Query.Predicate = &oddb.Predicate{
				Operator: oddb.And,
				Children: []interface{}{
					binaryPred(oddb.Equal, "falsy", false),
					binaryPred(oddb.Equal, "truthy", true),
				},
			}
			So(db.SaveSubscription(&suband), ShouldBeNil)

			subor := subscriptionForTest("device0", "or", "record")
			subor.Query.Predicate = &oddb.Predicate{
				Operator: oddb.Or,
				Children: []interface{}{
					binaryPred(oddb.Equal, "truthy", false),
					binaryPred(oddb.Equal, "falsy", false),
				},
			}
			So(db.SaveSubscription(&subor), ShouldBeNil)

			subnot := subscriptionForTest("device0", "not", "record")
			subnot.Query.Predicate = &oddb.Predicate{
				Operator: oddb.Not,
				Children: []interface{}{
					binaryPred(oddb.Equal, "truthy", false),
				},
			}
			So(db.SaveSubscription(&subnot), ShouldBeNil)

			record.Data = map[string]interface{}{
				"truthy": true,
				"falsy":  false,
			}
			subscriptions := db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []oddb.Subscription{suband, subor, subnot})

			record.Data = map[string]interface{}{
				"truthy": false,
				"falsy":  true,
			}
			subscriptions = db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []oddb.Subscription{subor})

			record.Data = map[string]interface{}{
				"truthy": true,
				"falsy":  true,
			}
			subscriptions = db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []oddb.Subscription{subnot})

			record.Data = map[string]interface{}{
				"truthy": false,
				"falsy":  false,
			}
			subscriptions = db.GetMatchingSubscriptions(&record)
			So(subscriptions, ShouldResemble, []oddb.Subscription{subor})
		})
	})
}

func TestGetSubscriptionsByDeviceID(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

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
			So(subscriptions, ShouldResemble, []oddb.Subscription{
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

func subscriptionForTest(deviceID, id, queryRecordType string) oddb.Subscription {
	return oddb.Subscription{
		ID:       id,
		Type:     "query",
		DeviceID: deviceID,
		Query: oddb.Query{
			Type: queryRecordType,
		},
	}
}
