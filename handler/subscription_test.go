package handler

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/oddbtest"
	"github.com/oursky/ourd/router"
)

func newFetchSubscription(id string) oddb.Subscription {
	return oddb.Subscription{
		ID:       id,
		Type:     "query",
		DeviceID: "deviceid",
		Query: oddb.Query{
			Type: "recordtype",
		},
	}
}

func TestSubscriptionFetchHandler(t *testing.T) {
	Convey("SubscriptionFetchHandler", t, func() {
		sub0 := newFetchSubscription("0")
		sub1 := newFetchSubscription("1")

		db := oddbtest.NewMapDB()
		db.SaveSubscription(&sub0)
		db.SaveSubscription(&sub1)

		r := newSingleRouteRouter(SubscriptionFetchHandler, func(p *router.Payload) {
			p.Database = db
		})

		Convey("fetchs multiple subscriptions", func() {
			resp := r.POST(`{"subscription_ids": ["0", "1"]}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), shouldEqualJSON, `{
	"result": [
		{
			"id": "0",
			"type": "query",
			"device_id": "deviceid",
			"query": {"record_type": "recordtype"}
		},
		{
			"id": "1",
			"type": "query",
			"device_id": "deviceid",
			"query": {"record_type": "recordtype"}
		}
	]
}`)
		})

		Convey("fetchs not existed subscriptions", func() {
			resp := r.POST(`{"subscription_ids": ["notexistid"]}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), shouldEqualJSON, `{
	"result": [{
		"_id": "notexistid",
		"_type": "error",
		"message": "cannot find subscription \"notexistid\"",
		"type": "ResourceNotFound",
		"code": 101,
		"info": {"id": "notexistid"}
	}]
}`)
		})

		Convey("fetchs without subscription_ids", func() {
			resp := r.POST(`{}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), shouldEqualJSON, `{"result": []}`)
		})
	})
}

func TestSubscriptionSaveHandler(t *testing.T) {
	Convey("SubscriptionSaveHandler", t, func() {
		db := oddbtest.NewMapDB()
		r := newSingleRouteRouter(SubscriptionSaveHandler, func(p *router.Payload) {
			p.Database = db
		})

		Convey("smoke test", func() {
			resp := r.POST(`
{
	"device_id": "somedeviceid",
	"subscriptions": [{
		"id": "subscription_id",
		"notification_info": {
			"aps": {
				"alert": {
					"body": "BODY_TEXT",
					"action-loc-key": "ACTION_LOC_KEY",
					"loc-key": "LOC_KEY",
					"loc-args": ["LOC_ARGS"],
					"launch-image": "LAUNCH_IMAGE"
				},
				"sound": "SOUND_NAME",
				"should-badge": true,
				"should-send-content-available": true
			}
		},
		"type": "query",
		"query": {
			"record_type": "RECORD_TYPE"
		}
	}]
}`)

			So(resp.Code, ShouldEqual, 200)
			// FIXME(limouren): The following JSON output is wrong by duplicating
			// subscription id in "_id" and "id"
			So(resp.Body.Bytes(), shouldEqualJSON, `
{
	"result": [{
		"_id": "subscription_id",
		"_type": "subscription",
		"id": "subscription_id",
		"device_id": "somedeviceid",
		"notification_info": {
			"aps": {
				"alert": {
					"body": "BODY_TEXT",
					"action-loc-key": "ACTION_LOC_KEY",
					"loc-key": "LOC_KEY",
					"loc-args": ["LOC_ARGS"],
					"launch-image": "LAUNCH_IMAGE"
				},
				"sound": "SOUND_NAME",
				"should-badge": true,
				"should-send-content-available": true
			}
		},
		"type": "query",
		"query": {
			"record_type": "RECORD_TYPE"
		}
	}]
}`)

			actualSubscription := oddb.Subscription{}
			So(db.GetSubscription("subscription_id", &actualSubscription), ShouldBeNil)
			So(actualSubscription, ShouldResemble, oddb.Subscription{
				ID:       "subscription_id",
				DeviceID: "somedeviceid",
				Type:     "query",
				NotificationInfo: &oddb.NotificationInfo{
					APS: oddb.APSSetting{
						Alert: &oddb.AppleAlert{
							Body:                  "BODY_TEXT",
							LocalizationKey:       "LOC_KEY",
							LocalizationArgs:      []string{"LOC_ARGS"},
							LaunchImage:           "LAUNCH_IMAGE",
							ActionLocalizationKey: "ACTION_LOC_KEY",
						},
						SoundName:                  "SOUND_NAME",
						ShouldBadge:                true,
						ShouldSendContentAvailable: true,
					},
				},
				Query: oddb.Query{
					Type: "RECORD_TYPE",
				},
			})
		})
	})
}
