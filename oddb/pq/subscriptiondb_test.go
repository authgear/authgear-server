package pq

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/ourd/oddb"
)

func addDevice(t *testing.T, c *conn, userID string, deviceID string) {
	_, err := c.Db.Exec("INSERT INTO app_com_oursky_ourd._device (id, user_id, type, token) VALUES ($1, $2, '', '')", deviceID, userID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSubscriptionCRUD(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		db := c.PrivateDB("userid")

		// fixture
		addUser(t, c, "userid")
		addDevice(t, c, "userid", "deviceid")

		notificationInfo := oddb.NotificationInfo{
			APS: oddb.APSSetting{
				Alert: oddb.AppleAlert{
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
		}
		subscription := oddb.Subscription{
			ID:               "subscriptionid",
			Type:             "query",
			DeviceID:         "deviceid",
			NotificationInfo: notificationInfo,
			Query:            query,
		}

		Convey("get an existing subscription", func() {
			So(db.SaveSubscription(&subscription), ShouldBeNil)

			resultSubscription := oddb.Subscription{}
			err := db.GetSubscription("subscriptionid", &resultSubscription)
			So(err, ShouldBeNil)
			So(subscription, ShouldResemble, resultSubscription)
		})

		Convey("returns ErrSubscriptionNotFound while trying to get a non-existing subscription ", func() {
			resultSubscription := oddb.Subscription{}
			err := db.GetSubscription("notexistsubscriptionid", &resultSubscription)
			So(err, ShouldEqual, oddb.ErrSubscriptionNotFound)
		})

		Convey("create new subscription", func() {
			err := db.SaveSubscription(&subscription)
			So(err, ShouldBeNil)

			var (
				deviceID, queryType    string
				resultNotificationInfo oddb.NotificationInfo
				resultQuery            oddb.Query
			)
			err = c.Db.QueryRow(`
				SELECT device_id, type, notification_info, query FROM app_com_oursky_ourd._subscription
				WHERE id = $1 AND user_id = $2`, "subscriptionid", "userid").
				Scan(&deviceID, &queryType, (*notificationInfoValue)(&resultNotificationInfo), (*queryValue)(&resultQuery))
			So(err, ShouldBeNil)

			So(deviceID, ShouldEqual, "deviceid")
			So(queryType, ShouldEqual, "query")
			So(resultNotificationInfo, ShouldResemble, notificationInfo)
			So(resultQuery, ShouldResemble, query)
		})

		Convey("modify existing subscription", func() {
			So(db.SaveSubscription(&subscription), ShouldBeNil)

			subscription.Query.Type = "otherrecordtype"
			err := db.SaveSubscription(&subscription)
			So(err, ShouldBeNil)

			var (
				deviceID, queryType    string
				resultNotificationInfo oddb.NotificationInfo
				resultQuery            oddb.Query
			)
			err = c.Db.QueryRow(`
				SELECT device_id, type, notification_info, query FROM app_com_oursky_ourd._subscription
				WHERE id = $1 AND user_id = $2`, "subscriptionid", "userid").
				Scan(&deviceID, &queryType, (*notificationInfoValue)(&resultNotificationInfo), (*queryValue)(&resultQuery))
			So(err, ShouldBeNil)

			So(deviceID, ShouldEqual, "deviceid")
			So(queryType, ShouldEqual, "query")
			So(resultNotificationInfo, ShouldResemble, notificationInfo)

			query.Type = "otherrecordtype"
			So(resultQuery, ShouldResemble, query)
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

		cleanupDB(t, c)
	})
}
