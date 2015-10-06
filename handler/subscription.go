package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/oddbconv"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

type subscriptionIDsPayload struct {
	DeviceID        string   `json:"device_id"`
	SubscriptionIDs []string `json:"subscription_ids"`
}

type subscriptionPayload struct {
	DeviceID      string `json:"device_id"`
	Subscriptions []struct {
		ID               string                 `json:"id"`
		Type             string                 `json:"type"`
		DeviceID         string                 `json:"device_id"`
		NotificationInfo *oddb.NotificationInfo `json:"notification_info,omitempty"`
		Query            map[string]interface{} `json:"query"`
	} `json:"subscriptions"`
}

type jsonSubscription oddb.Subscription

func (s *jsonSubscription) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID               string                 `json:"id"`
		Type             string                 `json:"type"`
		DeviceID         string                 `json:"device_id"`
		NotificationInfo *oddb.NotificationInfo `json:"notification_info,omitempty"`
		Query            jsonQuery              `json:"query"`
	}{
		s.ID,
		s.Type,
		s.DeviceID,
		s.NotificationInfo,
		jsonQuery(s.Query),
	})
}

type jsonQuery oddb.Query

func (q jsonQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type         string                     `json:"record_type"`
		Predicate    *jsonPredicate             `json:"predicate,omitempty"`
		Sorts        []oddb.Sort                `json:"order,omitempty"`
		ReadableBy   string                     `json:"readable_by,omitempty"`
		ComputedKeys map[string]oddb.Expression `json:"computed_keys,omitempty"`
		DesiredKeys  []string                   `json:"desired_keys,omitempty"`
		Limit        uint64                     `json:"limit,omitempty"`
		Offset       uint64                     `json:"offset,omitempty"`
	}{
		q.Type,
		(*jsonPredicate)(q.Predicate),
		q.Sorts,
		q.ReadableBy,
		q.ComputedKeys,
		q.DesiredKeys,
		q.Limit,
		q.Offset,
	})
}

type jsonPredicate oddb.Predicate

func (p *jsonPredicate) MarshalJSON() ([]byte, error) {
	var results []interface{}
	if p.Operator.IsCompound() {
		results = append(results, opString(p.Operator))
		for i, child := range p.Children {
			childPred, ok := child.(oddb.Predicate)
			if !ok {
				return nil, fmt.Errorf("got %s.Operand[%d] of type %T, want Predicate",
					p.Operator, i, child)
			}
			results = append(results, jsonPredicate(childPred))
		}
	} else {
		operandLen := 1
		if p.Operator.IsBinary() {
			operandLen = 2
		}

		if operandLen != len(p.Children) {
			return nil, fmt.Errorf("got len(operand) = %d, want %d", len(p.Children), operandLen)
		}

		results = append(results, opString(p.Operator))
		for i := 0; i < operandLen; i++ {
			child := p.Children[i]
			childExpr, ok := child.(oddb.Expression)
			if !ok {
				return nil, fmt.Errorf("got %s.Operand[%d] of type %T, want Expression",
					p.Operator, i, child)
			}
			results = append(results, jsonExpression(childExpr))
		}
	}

	return json.Marshal(results)
}

type jsonExpression oddb.Expression

func (expr jsonExpression) MarshalJSON() ([]byte, error) {
	var i interface{}
	switch expr.Type {
	case oddb.Literal:
		switch v := expr.Value.(type) {
		case oddb.Reference:
			i = oddbconv.ToMap(oddbconv.MapReference(v))
		default:
			i = expr.Value
		}
	case oddb.KeyPath:
		i = oddbconv.ToMap(oddbconv.MapKeyPath(expr.Value.(string)))
	case oddb.Function:
		i = funcSlice(expr.Value)
	default:
		return nil, fmt.Errorf("unrecgonized ExpressionType = %v", expr.Type)
	}

	return json.Marshal(i)
}

func opString(op oddb.Operator) string {
	switch op {
	case oddb.And:
		return "and"
	case oddb.Or:
		return "or"
	case oddb.Not:
		return "not"
	case oddb.Equal:
		return "eq"
	case oddb.GreaterThan:
		return "gt"
	case oddb.LessThan:
		return "lt"
	case oddb.GreaterThanOrEqual:
		return "gte"
	case oddb.LessThanOrEqual:
		return "lte"
	case oddb.NotEqual:
		return "neq"
	case oddb.Like:
		return "like"
	case oddb.ILike:
		return "ilike"
	default:
		return "UNKNOWN_OPERATOR"
	}
}

func funcSlice(i interface{}) []interface{} {
	switch f := i.(type) {
	case *oddb.DistanceFunc:
		return []interface{}{
			"func",
			"distance",
			oddbconv.ToMap(oddbconv.MapKeyPath(f.Field)),
			oddbconv.ToMap((*oddbconv.MapLocation)(f.Location)),
		}
	default:
		panic(fmt.Errorf("got unrecgonized oddb.Func = %T", i))
	}
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
			item = jsonSubscription(subscription)
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
		return
	}

	rawSubs := payload.Subscriptions
	if len(rawSubs) == 0 {
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty subscriptions"))
		return
	}

	if payload.DeviceID == "" {
		response.Err = oderr.NewRequestInvalidErr(errors.New("empty device_id"))
		return
	}

	subscriptions := make([]oddb.Subscription, len(rawSubs), len(rawSubs))
	for i, rawSub := range rawSubs {
		sub := &subscriptions[i]
		sub.ID = rawSub.ID
		sub.Type = rawSub.Type
		sub.DeviceID = rawSub.DeviceID
		sub.NotificationInfo = rawSub.NotificationInfo
		sub.DeviceID = payload.DeviceID
		if err := queryFromRaw(rawSub.Query, &sub.Query); err != nil {
			response.Err = oderr.NewRequestInvalidErr(fmt.Errorf(
				"failed to parse subscriptions: %v", err))
			return
		}
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
			item = (*jsonSubscription)(subscription)
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
