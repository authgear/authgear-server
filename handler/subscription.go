// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/skyconv"
	"github.com/oursky/skygear/skyerr"
)

type jsonSubscription skydb.Subscription

func (s jsonSubscription) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID               string                  `json:"id"`
		Type             string                  `json:"type"`
		DeviceID         string                  `json:"device_id"`
		NotificationInfo *skydb.NotificationInfo `json:"notification_info,omitempty"`
		Query            jsonQuery               `json:"query"`
	}{
		s.ID,
		s.Type,
		s.DeviceID,
		s.NotificationInfo,
		jsonQuery(s.Query),
	})
}

type jsonQuery skydb.Query

func (q jsonQuery) MarshalJSON() ([]byte, error) {
	var optionalPredicate *jsonPredicate
	if !q.Predicate.IsEmpty() {
		optionalPredicate = (*jsonPredicate)(&q.Predicate)
	}
	return json.Marshal(struct {
		Type         string                      `json:"record_type"`
		Predicate    *jsonPredicate              `json:"predicate,omitempty"`
		Sorts        []skydb.Sort                `json:"order,omitempty"`
		ComputedKeys map[string]skydb.Expression `json:"computed_keys,omitempty"`
		DesiredKeys  []string                    `json:"desired_keys,omitempty"`
		Limit        *uint64                     `json:"limit,omitempty"`
		Offset       uint64                      `json:"offset,omitempty"`
	}{
		q.Type,
		optionalPredicate,
		q.Sorts,
		q.ComputedKeys,
		q.DesiredKeys,
		q.Limit,
		q.Offset,
	})
}

type jsonPredicate skydb.Predicate

func (p *jsonPredicate) MarshalJSON() ([]byte, error) {
	if (*skydb.Predicate)(p).IsEmpty() {
		return []byte{}, nil
	}

	var results []interface{}
	if p.Operator.IsCompound() {
		results = append(results, opString(p.Operator))
		for i, child := range p.Children {
			childPred, ok := child.(skydb.Predicate)
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
			childExpr, ok := child.(skydb.Expression)
			if !ok {
				return nil, fmt.Errorf("got %s.Operand[%d] of type %T, want Expression",
					p.Operator, i, child)
			}
			results = append(results, jsonExpression(childExpr))
		}
	}

	return json.Marshal(results)
}

type jsonExpression skydb.Expression

func (expr jsonExpression) MarshalJSON() ([]byte, error) {
	var i interface{}
	switch expr.Type {
	case skydb.Literal:
		switch v := expr.Value.(type) {
		case skydb.Reference:
			i = skyconv.ToMap(skyconv.MapReference(v))
		default:
			i = expr.Value
		}
	case skydb.KeyPath:
		i = skyconv.ToMap(skyconv.MapKeyPath(expr.Value.(string)))
	case skydb.Function:
		i = funcSlice(expr.Value)
	default:
		return nil, fmt.Errorf("unrecgonized ExpressionType = %v", expr.Type)
	}

	return json.Marshal(i)
}

func opString(op skydb.Operator) string {
	switch op {
	case skydb.And:
		return "and"
	case skydb.Or:
		return "or"
	case skydb.Not:
		return "not"
	case skydb.Equal:
		return "eq"
	case skydb.GreaterThan:
		return "gt"
	case skydb.LessThan:
		return "lt"
	case skydb.GreaterThanOrEqual:
		return "gte"
	case skydb.LessThanOrEqual:
		return "lte"
	case skydb.NotEqual:
		return "neq"
	case skydb.Like:
		return "like"
	case skydb.ILike:
		return "ilike"
	case skydb.In:
		return "in"
	default:
		return "UNKNOWN_OPERATOR"
	}
}

func funcSlice(i interface{}) []interface{} {
	switch f := i.(type) {
	case skydb.DistanceFunc:
		return []interface{}{
			"func",
			"distance",
			skyconv.ToMap(skyconv.MapKeyPath(f.Field)),
			skyconv.ToMap(skyconv.MapLocation(f.Location)),
		}
	default:
		panic(fmt.Errorf("got unrecgonized skydb.Func = %T", i))
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
	var err skyerr.Error
	if e.err == skydb.ErrSubscriptionNotFound {
		err = skyerr.NewErrorWithInfo(skyerr.ResourceNotFound, fmt.Sprintf(`cannot find subscription "%s"`, e.id), map[string]interface{}{"id": e.id})
	} else {
		err = skyerr.NewError(skyerr.UnexpectedError, fmt.Sprintf("unknown error occurred: %v", e.err.Error()))
	}
	return json.Marshal(&struct {
		ID       string                 `json:"_id"`
		ItemType string                 `json:"_type"`
		Message  string                 `json:"message"`
		Name     string                 `json:"name"`
		Code     skyerr.ErrorCode       `json:"code"`
		Info     map[string]interface{} `json:"info,omitempty"`
	}{e.id, "error", err.Message(), err.Name(), err.Code(), err.Info()})
}

// subscriptionPayload is shared by SubscriptionFetchHandler and SubscriptionDeleteHandler.
type subscriptionPayload struct {
	DeviceID        string   `mapstructure:"device_id"`
	SubscriptionIDs []string `mapstructure:"subscription_ids"`
}

func (payload *subscriptionPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *subscriptionPayload) Validate() skyerr.Error {
	if payload.DeviceID == "" {
		return skyerr.NewInvalidArgument("empty device_id", []string{"device_id"})
	}

	return nil
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
type SubscriptionFetchHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SubscriptionFetchHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SubscriptionFetchHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SubscriptionFetchHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := &subscriptionPayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
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

		subscription := skydb.Subscription{}
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
type SubscriptionFetchAllHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SubscriptionFetchAllHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SubscriptionFetchAllHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SubscriptionFetchAllHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := &subscriptionPayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	subscriptions := rpayload.Database.GetSubscriptionsByDeviceID(payload.DeviceID)

	results := []jsonSubscription{}
	for _, sub := range subscriptions {
		results = append(results, jsonSubscription(sub))
	}

	if len(results) > 0 {
		response.Result = results
	}
}

type subscriptionSavePayload struct {
	DeviceID      string               `json:"device_id"`
	Subscriptions []skydb.Subscription `json:"subscriptions"`
}

func (payload *subscriptionSavePayload) Decode(data map[string]interface{}, parser *QueryParser) skyerr.Error {
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     payload,
		TagName:    "json",
		DecodeHook: mapToQueryHookFunc(parser),
	})
	if err != nil {
		panic(err)
	}
	if err := mapDecoder.Decode(data); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *subscriptionSavePayload) Validate() skyerr.Error {
	if len(payload.Subscriptions) == 0 {
		return skyerr.NewInvalidArgument("empty subscriptions", []string{"subscriptions"})
	}

	if payload.DeviceID == "" {
		return skyerr.NewInvalidArgument("empty device_id", []string{"device_id"})
	}

	// Reset the device ID for individual subscription to the device ID
	// specified in the top-level of the payload.
	for i := range payload.Subscriptions {
		subscription := &payload.Subscriptions[i]
		subscription.DeviceID = payload.DeviceID
	}

	return nil
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
type SubscriptionSaveHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SubscriptionSaveHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SubscriptionSaveHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SubscriptionSaveHandler) Handle(rpayload *router.Payload, response *router.Response) {
	parser := QueryParser{UserID: rpayload.UserInfoID}
	payload := &subscriptionSavePayload{}
	skyErr := payload.Decode(rpayload.Data, &parser)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	db := rpayload.Database
	results := make([]interface{}, 0, len(payload.Subscriptions))
	var (
		subscription *skydb.Subscription
		item         interface{}
	)
	for i := range payload.Subscriptions {
		subscription = &payload.Subscriptions[i]
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
type SubscriptionDeleteHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *SubscriptionDeleteHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *SubscriptionDeleteHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SubscriptionDeleteHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := &subscriptionPayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
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
