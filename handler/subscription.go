package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

type subscriptionIDsPayload struct {
	SubscriptionIDs []string `json:"subscription_ids"`
}

type subscriptionPayload struct {
	DeviceID      string              `json:"device_id"`
	Subscriptions []oddb.Subscription `json:"subscriptions"`
}

type subscriptionItem struct {
	id           string
	subscription *oddb.Subscription
	err          oderr.Error
}

func newSubscriptionResponseItem(subscription *oddb.Subscription) subscriptionItem {
	return subscriptionItem{
		id:           subscription.ID,
		subscription: subscription,
	}
}

func newSubscriptionResponseItemErr(id string, err oderr.Error) subscriptionItem {
	return subscriptionItem{
		id:  id,
		err: err,
	}
}

func (item subscriptionItem) MarshalJSON() ([]byte, error) {
	var (
		buf bytes.Buffer
		i   interface{}
	)
	buf.Write([]byte(`{"_id":"`))
	buf.WriteString(item.id)
	buf.Write([]byte(`","_type":"`))
	if item.err != nil {
		buf.Write([]byte(`error",`))
		i = item.err
	} else if item.subscription != nil {
		buf.Write([]byte(`subscription",`))
		i = item.subscription
	} else {
		panic("inconsistent state: both err and subscription is nil")
	}

	bodyBytes, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	if bodyBytes[0] != '{' {
		return nil, fmt.Errorf("first char of embedded json != {: %v", string(bodyBytes))
	} else if bodyBytes[len(bodyBytes)-1] != '}' {
		return nil, fmt.Errorf("last char of embedded json != }: %v", string(bodyBytes))
	}
	buf.Write(bodyBytes[1:])
	return buf.Bytes(), nil
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

// SubscriptionFetchHandler fetchs subscriptions from the specified Database.
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

	if len(payload.SubscriptionIDs) == 0 {
		response.Result = []interface{}{}
		return
	}

	db := rpayload.Database
	results := make([]interface{}, 0, len(payload.SubscriptionIDs))
	for _, id := range payload.SubscriptionIDs {
		var item interface{}

		subscription := oddb.Subscription{}
		if err := db.GetSubscription(id, &subscription); err != nil {
			// handle err here
			item = newErrorWithID(id, err)
		} else {
			item = subscription
		}

		results = append(results, item)
	}

	response.Result = results
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
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty subsciptions"))
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
	results := make([]interface{}, len(subscriptions), len(subscriptions))
	for i := range subscriptions {
		if err := db.SaveSubscription(&subscriptions[i]); err != nil {
			results[i] = newSubscriptionResponseItemErr(
				subscriptions[i].ID,
				oderr.NewResourceSaveFailureErrWithStringID("subscription", subscriptions[i].ID),
			)
		} else {
			results[i] = newSubscriptionResponseItem(&subscriptions[i])
		}
	}

	response.Result = results
}
