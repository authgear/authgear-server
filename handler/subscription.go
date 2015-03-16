package handler

import (
	"github.com/mitchellh/mapstructure"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

type subscriptionPayload struct {
	Subscriptions []oddb.Subscription
}

// SubscriptionSaveHandler saves one or more subscriptions associate with
// a database.
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
		response.Err = oderr.New(oderr.RequestInvalidErr, "invalid request: "+err.Error())
	}

	subscriptions := payload.Subscriptions
	if len(subscriptions) == 0 {
		response.Err = oderr.New(oderr.RequestInvalidErr, "empty subsciptions")
		return
	}

	db := rpayload.Database
	results := make([]interface{}, len(subscriptions), len(subscriptions))
	for i := range subscriptions {
		if err := db.SaveSubscription(&subscriptions[i]); err != nil {
			results[i] = oderr.New(oderr.PersistentStorageErr, "persistent: failed to save subscription")
		} else {
			results[i] = subscriptions[i]
		}
	}

	response.Result = results
}
