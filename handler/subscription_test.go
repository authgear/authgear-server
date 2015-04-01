package handler

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"encoding/json"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/oddbtest"
	"github.com/oursky/ourd/router"
)

func TestSubscriptionSaveHandler(t *testing.T) {
	Convey("SubscriptionSaveHandler", t, func() {
		db := oddbtest.NewMapDB()
		payload := router.Payload{
			Data:     map[string]interface{}{},
			Database: db,
		}
		response := router.Response{}

		Convey("smoke test", func() {
			if err := json.Unmarshal([]byte(`
{
	"Subscriptions": [{
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
			"predicate": {}
		}
	}]
}
			`), &payload.Data); err != nil {
				panic(err)
			}
			SubscriptionSaveHandler(&payload, &response)

			expectedSubscription := oddb.Subscription{
				ID:   "subscription_id",
				Type: "query",
				NotificationInfo: oddb.NotificationInfo{
					APS: oddb.APSSetting{
						Alert: oddb.AppleAlert{
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
			}
			So(response.Result, ShouldResemble, []interface{}{
				newSubscriptionResponseItem(&expectedSubscription),
			})
		})
	})
}
