package handler

import (
	"testing"

	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/oddb"
	"github.com/oursky/skygear/oddb/oddbtest"
	. "github.com/oursky/skygear/ourtest"
	"github.com/oursky/skygear/router"
	. "github.com/smartystreets/goconvey/convey"
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

		r := handlertest.NewSingleRouteRouter(SubscriptionFetchHandler, func(p *router.Payload) {
			p.Database = db
		})

		Convey("fetches multiple subscriptions", func() {
			resp := r.POST(`{"device_id": "deviceid", "subscription_ids": ["0", "1"]}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
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

		Convey("fetches not existed subscriptions", func() {
			resp := r.POST(`{"device_id": "deviceid", "subscription_ids": ["notexistid"]}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
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

		Convey("fetches without device_id", func() {
			resp := r.POST(`{}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"type":"RequestInvalid","code":101,"message":"empty device_id"}}`)
			So(resp.Code, ShouldEqual, 400)
		})

		Convey("fetches without subscription_ids", func() {
			resp := r.POST(`{"device_id": "deviceid"}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result": []}`)
			So(resp.Code, ShouldEqual, 200)
		})
	})
}

type fetchallDB struct {
	subscriptions []oddb.Subscription
	lastDeviceID  string
	oddb.Database
}

func newFetchallDB(subscriptions ...oddb.Subscription) *fetchallDB {
	return &fetchallDB{subscriptions: subscriptions}
}

func (db *fetchallDB) GetSubscriptionsByDeviceID(deviceID string) []oddb.Subscription {
	db.lastDeviceID = deviceID
	return db.subscriptions
}

func TestSubscriptionFetchAllHandler(t *testing.T) {
	Convey("SubscriptionFetchAllHandler", t, func() {
		subscriptions := []oddb.Subscription{
			newFetchSubscription("0"),
			newFetchSubscription("1"),
			newFetchSubscription("2"),
		}
		db := newFetchallDB(subscriptions...)

		r := handlertest.NewSingleRouteRouter(SubscriptionFetchAllHandler, func(p *router.Payload) {
			p.Database = db
		})

		Convey("fetches all subscriptions", func() {
			resp := r.POST(`{
	"device_id": "deviceid"
}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"id": "0",
		"type": "query",
		"device_id": "deviceid",
		"query": {"record_type": "recordtype"}
	}, {
		"id": "1",
		"type": "query",
		"device_id": "deviceid",
		"query": {"record_type": "recordtype"}
	}, {
		"id": "2",
		"type": "query",
		"device_id": "deviceid",
		"query": {"record_type": "recordtype"}
	}]
}`)

			So(db.lastDeviceID, ShouldEqual, "deviceid")
		})

		Convey("errors with empty device id", func() {
			resp := r.POST(`{}`)
			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error": {"code": 101, "message": "empty device id", "type": "RequestInvalid"}}`)
		})
	})
}

func TestSubscriptionSaveHandler(t *testing.T) {
	Convey("SubscriptionSaveHandler", t, func() {
		db := oddbtest.NewMapDB()
		r := handlertest.NewSingleRouteRouter(SubscriptionSaveHandler, func(p *router.Payload) {
			p.Database = db
		})

		Convey("saves one subscription", func() {
			resp := r.POST(`{
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
						"record_type": "RECORD_TYPE",
						"predicate": [
							"eq",
							{
								"$val": "_id",
								"$type": "keypath"
							},
							"RECORD_ID"
						]
					}
				}]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": [{
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
						"record_type": "RECORD_TYPE",
						"predicate": [
							"eq",
							{
								"$val": "_id",
								"$type": "keypath"
							},
							"RECORD_ID"
						]
					}
				}]
			}`)
			So(resp.Code, ShouldEqual, 200)

			actualSubscription := oddb.Subscription{}
			So(db.GetSubscription("subscription_id", "somedeviceid", &actualSubscription), ShouldBeNil)
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
				},
			})
		})

		Convey("saves two subscriptions", func() {
			resp := r.POST(`
{
	"device_id": "somedeviceid",
	"subscriptions": [{
		"id": "sub0",
		"type": "query",
		"query": {
			"record_type": "recordtype0"
		}
	}, {
		"id": "sub1",
		"type": "query",
		"query": {
			"record_type": "recordtype1"
		}
	}]
}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [{
		"id": "sub0",
		"device_id": "somedeviceid",
		"type": "query",
		"query": {
			"record_type": "recordtype0"
		}
	}, {
		"id": "sub1",
		"device_id": "somedeviceid",
		"type": "query",
		"query": {
			"record_type": "recordtype1"
		}
	}]
}`)

			var sub0, sub1 oddb.Subscription
			So(db.GetSubscription("sub0", "somedeviceid", &sub0), ShouldBeNil)
			So(db.GetSubscription("sub1", "somedeviceid", &sub1), ShouldBeNil)

			So(sub0, ShouldResemble, oddb.Subscription{
				ID:       "sub0",
				DeviceID: "somedeviceid",
				Type:     "query",
				Query: oddb.Query{
					Type: "recordtype0",
				},
			})
			So(sub1, ShouldResemble, oddb.Subscription{
				ID:       "sub1",
				DeviceID: "somedeviceid",
				Type:     "query",
				Query: oddb.Query{
					Type: "recordtype1",
				},
			})
		})

		Convey("errors without device_id", func() {
			resp := r.POST(`
{
	"subscriptions": [{
		"id": "subscription_id",
		"type": "query",
		"query": {
			"record_type": "RECORD_TYPE"
		}
	}]
}`)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"code":101,"message":"empty device_id","type":"RequestInvalid"}}`)
		})

		Convey("errors without subscriptions", func() {
			resp := r.POST(`{"device_id":"somedeviceid"}`)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"code":101,"message":"empty subscriptions","type":"RequestInvalid"}}`)
		})
	})
}

func TestSubscriptionDeleteHandler(t *testing.T) {
	Convey("SubscriptionFetchHandler", t, func() {
		sub0 := newFetchSubscription("0")
		sub1 := newFetchSubscription("1")

		db := oddbtest.NewMapDB()
		db.SaveSubscription(&sub0)
		db.SaveSubscription(&sub1)

		r := handlertest.NewSingleRouteRouter(SubscriptionDeleteHandler, func(p *router.Payload) {
			p.Database = db
		})

		Convey("deletes multiple subscriptions", func() {
			resp := r.POST(`{"device_id": "deviceid", "subscription_ids": ["0", "1"]}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{"id": "0"},
		{"id": "1"}
	]
}`)
		})

		Convey("deletes not existed subscriptions", func() {
			resp := r.POST(`{"device_id": "deviceid", "subscription_ids": ["notexistid"]}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
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

		Convey("deletes without device_id", func() {
			resp := r.POST(`{}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"type":"RequestInvalid","code":101,"message":"empty device_id"}}`)
			So(resp.Code, ShouldEqual, 400)
		})

		Convey("deletes without subscription_ids", func() {
			resp := r.POST(`{"device_id": "deviceid"}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result": []}`)
		})
	})
}
