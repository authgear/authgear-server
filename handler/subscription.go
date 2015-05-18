package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

type subscriptionIDsPayload struct {
	DeviceID        string   `json:"device_id"`
	SubscriptionIDs []string `json:"subscription_ids"`
}

type subscriptionPayload struct {
	DeviceID      string              `json:"device_id"`
	Subscriptions []oddb.Subscription `json:"subscriptions"`
}

// FIXME(limouren): settle on a way to centralize error creation
type errorWithID struct {
	id  string
	err error
}

func newErrorWithID(id string, err error) *errorWithID {
	return &errorWithID{id, err}
}

func (e *errorWithID) MarshalJSON() ([]byte, error) {
	var (
		message string
		t       string
		code    uint
		info    map[string]interface{}
	)
	if e.err == oddb.ErrSubscriptionNotFound {
		message = fmt.Sprintf(`cannot find subscription "%s"`, e.id)
		t = "ResourceNotFound"
		code = 101
		info = map[string]interface{}{"id": e.id}
	} else {
		message = fmt.Sprintf("unknown error occurred: %v", e.err.Error())
		t = "UnknownError"
		code = 1
	}
	return json.Marshal(&struct {
		ID       string                 `json:"_id"`
		ItemType string                 `json:"_type"`
		Message  string                 `json:"message"`
		Type     string                 `json:"type"`
		Code     uint                   `json:"code"`
		Info     map[string]interface{} `json:"info,omitempty"`
	}{e.id, "error", message, t, code, info})
}

// SubscriptionFetchHandler fetches subscriptions from the specified Database.
//
// Example curl:
//	curl -X POST -H "Content-Type: application/json" \
//	  -d @- http://localhost:3000/ <<EOF
//	{
//	    "action": "subscription:fetch",
//	    "access_token": "ACCESS_TOKEN",
//	    "database_id": "_private",
//	    "device_id": "DEVICE_ID",
//	    "subscription_ids": ["SUBSCRIPTION_ID"]
//	}
//	EOF
func SubscriptionFetchHandler(rpayload *router.Payload, response *router.Response) {
	payload := subscriptionIDsPayload{}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &payload,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}
	if err := mapDecoder.Decode(rpayload.Data); err != nil {
		response.Err = oderr.NewRequestInvalidErr(err)
		return
	}

	if payload.DeviceID == "" {
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty device_id"))
		return
	}

	if len(payload.SubscriptionIDs) == 0 {
		response.Result = []interface{}{}
		return
	}

	db := rpayload.Database
	results := make([]interface{}, 0, len(payload.SubscriptionIDs))
	for _, id := range payload.SubscriptionIDs {
		var item interface{}

		subscription := oddb.Subscription{}
		if err := db.GetSubscription(id, payload.DeviceID, &subscription); err != nil {
			// handle err here
			item = newErrorWithID(id, err)
		} else {
			item = subscription
		}

		results = append(results, item)
	}

	response.Result = results
}

// SubscriptionFetchAllHandler fetches all subscriptions of a device
//
//	curl -X POST -H "Content-Type: application/json" \
//	  -d @- http://localhost:3000/ <<EOF
//	{
//	    "action": "subscription:fetch_all",
//	    "access_token": "ACCESS_TOKEN",
//	    "database_id": "_private",
//	    "device_id": "DEVICE_ID"
//	}
//	EOF
func SubscriptionFetchAllHandler(rpayload *router.Payload, response *router.Response) {
	var payload struct {
		DeviceID string `json:"device_id"`
	}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &payload,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}
	if err := mapDecoder.Decode(rpayload.Data); err != nil {
		response.Err = oderr.NewRequestInvalidErr(err)
		return
	}

	if payload.DeviceID == "" {
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty device id"))
		return
	}

	response.Result = rpayload.Database.GetSubscriptionsByDeviceID(payload.DeviceID)
}

// SubscriptionSaveHandler saves one or more subscriptions associate with
// a database.
//
// Example curl:
//	curl -X POST -H "Content-Type: application/json" \
//	  -d @- http://localhost:3000/ <<EOF
//	{
//	    "action": "subscription:save",
//	    "access_token": "ACCESS_TOKEN",
//	    "database_id": "_private",
//	    "device_id": "DEVICE_ID",
//	    "subscriptions": [
//	        {
//	            "id": "SUBSCRIPTION_ID",
//	            "notification_info": {
//	                "aps": {
//	                    "alert": {
//	                        "body": "BODY_TEXT",
//	                        "action-loc-key": "ACTION_LOC_KEY",
//	                        "loc-key": "LOC_KEY",
//	                        "loc-args": ["LOC_ARGS"],
//	                        "launch-image": "LAUNCH_IMAGE"
//	                    },
//	                    "sound": "SOUND_NAME",
//	                    "should-badge": true,
//	                    "should-send-content-available": true
//	                }
//	            },
//	            "type": "query",
//	            "query": {
//	                "record_type": "RECORD_TYPE",
//	                "predicate": {}
//	            }
//	        }
//	    ]
//	}
//	EOF
func SubscriptionSaveHandler(rpayload *router.Payload, response *router.Response) {
	payload := subscriptionPayload{}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &payload,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}
	if err := mapDecoder.Decode(rpayload.Data); err != nil {
		response.Err = oderr.NewRequestInvalidErr(err)
	}

	subscriptions := payload.Subscriptions
	if len(subscriptions) == 0 {
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty subscriptions"))
		return
	}

	if payload.DeviceID == "" {
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty device_id"))
		return
	}

	for i := range subscriptions {
		subscriptions[i].DeviceID = payload.DeviceID
	}

	db := rpayload.Database
	results := make([]interface{}, 0, len(subscriptions))
	var (
		subscription *oddb.Subscription
		item         interface{}
	)
	for i := range subscriptions {
		subscription = &subscriptions[i]
		if err := db.SaveSubscription(subscription); err != nil {
			item = newErrorWithID(subscription.ID, err)
		} else {
			item = subscription
		}
		results = append(results, item)
	}

	response.Result = results
}

// SubscriptionDeleteHandler deletes subscriptions from the specified Database.
//
// Example curl:
//	curl -X POST -H "Content-Type: application/json" \
//	  -d @- http://localhost:3000/ <<EOF
//	{
//	    "action": "subscription:delete",
//	    "access_token": "ACCESS_TOKEN",
//	    "database_id": "_private",
//	    "subscription_ids": ["SUBSCRIPTION_ID"]
//	}
//	EOF
func SubscriptionDeleteHandler(rpayload *router.Payload, response *router.Response) {
	payload := subscriptionIDsPayload{}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &payload,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}
	if err := mapDecoder.Decode(rpayload.Data); err != nil {
		response.Err = oderr.NewRequestInvalidErr(err)
		return
	}

	if payload.DeviceID == "" {
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty device_id"))
		return
	}

	if len(payload.SubscriptionIDs) == 0 {
		response.Result = []interface{}{}
		return
	}

	db := rpayload.Database
	results := make([]interface{}, 0, len(payload.SubscriptionIDs))
	for _, id := range payload.SubscriptionIDs {
		var item interface{}

		if err := db.DeleteSubscription(id, payload.DeviceID); err != nil {
			item = newErrorWithID(id, err)
		} else {
			item = struct {
				ID string `json:"id"`
			}{id}
		}

		results = append(results, item)
	}

	response.Result = results
}
