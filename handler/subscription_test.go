package handler

import (
	"testing"

	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/skydbtest"
	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func newFetchSubscription(id string) skydb.Subscription {
	return skydb.Subscription{
		ID:       id,
		Type:     "query",
		DeviceID: "deviceid",
		Query: skydb.Query{
			Type: "recordtype",
		},
	}
}

func TestSubscriptionFetchHandler(t *testing.T) {
	Convey("SubscriptionFetchHandler", t, func() {
		sub0 := newFetchSubscription("0")
		sub1 := newFetchSubscription("1")

		db := skydbtest.NewMapDB()
		db.SaveSubscription(&sub0)
		db.SaveSubscription(&sub1)

		r := handlertest.NewSingleRouteRouter(&SubscriptionFetchHandler{}, func(p *router.Payload) {
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
		"code": 110,
		"name": "ResourceNotFound",
		"info": {"id": "notexistid"}
	}]
}`)
		})

		Convey("fetches without device_id", func() {
			resp := r.POST(`{}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"code":108,"info":{"arguments":["device_id"]},"name":"InvalidArgument","message":"empty device_id"}}`)
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
	subscriptions []skydb.Subscription
	lastDeviceID  string
	skydb.Database
}

func newFetchallDB(subscriptions ...skydb.Subscription) *fetchallDB {
	return &fetchallDB{subscriptions: subscriptions}
}

func (db *fetchallDB) GetSubscriptionsByDeviceID(deviceID string) []skydb.Subscription {
	db.lastDeviceID = deviceID
	return db.subscriptions
}

func TestSubscriptionFetchAllHandler(t *testing.T) {
	Convey("SubscriptionFetchAllHandler", t, func() {
		subscriptions := []skydb.Subscription{
			newFetchSubscription("0"),
			newFetchSubscription("1"),
			newFetchSubscription("2"),
		}
		db := newFetchallDB(subscriptions...)

		r := handlertest.NewSingleRouteRouter(&SubscriptionFetchAllHandler{}, func(p *router.Payload) {
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
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error": {"code": 108, "message": "empty device_id", "info": {"arguments": ["device_id"]}, "name": "InvalidArgument"}}`)
		})
	})
}

func TestSubscriptionSaveHandler(t *testing.T) {
	Convey("SubscriptionSaveHandler", t, func() {
		db := skydbtest.NewMapDB()
		r := handlertest.NewSingleRouteRouter(&SubscriptionSaveHandler{}, func(p *router.Payload) {
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

			actualSubscription := skydb.Subscription{}
			So(db.GetSubscription("subscription_id", "somedeviceid", &actualSubscription), ShouldBeNil)
			So(actualSubscription, ShouldResemble, skydb.Subscription{
				ID:       "subscription_id",
				DeviceID: "somedeviceid",
				Type:     "query",
				NotificationInfo: &skydb.NotificationInfo{
					APS: skydb.APSSetting{
						Alert: &skydb.AppleAlert{
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
				Query: skydb.Query{
					Type: "RECORD_TYPE",
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

			var sub0, sub1 skydb.Subscription
			So(db.GetSubscription("sub0", "somedeviceid", &sub0), ShouldBeNil)
			So(db.GetSubscription("sub1", "somedeviceid", &sub1), ShouldBeNil)

			So(sub0, ShouldResemble, skydb.Subscription{
				ID:       "sub0",
				DeviceID: "somedeviceid",
				Type:     "query",
				Query: skydb.Query{
					Type: "recordtype0",
				},
			})
			So(sub1, ShouldResemble, skydb.Subscription{
				ID:       "sub1",
				DeviceID: "somedeviceid",
				Type:     "query",
				Query: skydb.Query{
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
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"code":108,"message":"empty device_id","name":"InvalidArgument","info":{"arguments":["device_id"]}}}`)
		})

		Convey("errors without subscriptions", func() {
			resp := r.POST(`{"device_id":"somedeviceid"}`)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"code":108,"message":"empty subscriptions","name":"InvalidArgument","info":{"arguments":["subscriptions"]}}}`)
		})
	})
}

func TestSubscriptionDeleteHandler(t *testing.T) {
	Convey("SubscriptionFetchHandler", t, func() {
		sub0 := newFetchSubscription("0")
		sub1 := newFetchSubscription("1")

		db := skydbtest.NewMapDB()
		db.SaveSubscription(&sub0)
		db.SaveSubscription(&sub1)

		r := handlertest.NewSingleRouteRouter(&SubscriptionDeleteHandler{}, func(p *router.Payload) {
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
		"name": "ResourceNotFound",
		"code": 110,
		"info": {"id": "notexistid"}
	}]
}`)
		})

		Convey("deletes without device_id", func() {
			resp := r.POST(`{}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"code":108,"message":"empty device_id","info":{"arguments":["device_id"]},"name":"InvalidArgument"}}`)
			So(resp.Code, ShouldEqual, 400)
		})

		Convey("deletes without subscription_ids", func() {
			resp := r.POST(`{"device_id": "deviceid"}`)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result": []}`)
		})
	})
}
