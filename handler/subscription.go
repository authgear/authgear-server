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

type subscriptionPayload struct {
	Subscriptions []oddb.Subscription
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
		response.Err = oderr.NewRequestInvalidErr(err)
	}

	subscriptions := payload.Subscriptions
	if len(subscriptions) == 0 {
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty subsciptions"))
		return
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
