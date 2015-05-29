package handler

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

type relationPayload struct {
	Name      string   `json:"name"`
	Direction string   `json:"direction"`
	Target    []string `json:"targets"`
}

func relationColander(data map[string]interface{}, result *relationPayload) error {
	mapDecoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  result,
		TagName: "json",
	})
	if err := mapDecoder.Decode(data); err != nil {
		return oderr.NewRequestJSONInvalidErr(err)
	}
	if result.Name != "friend" && result.Name != "follow" {
		return oderr.NewRequestInvalidErr(
			errors.New("Only friend and follow relation is supported"))
	}
	if result.Direction != "" {
		if result.Direction != "active" && result.Direction != "passive" && result.Direction != "mutual" {
			return oderr.NewRequestInvalidErr(
				errors.New("Only active, passive and mutual direction is supported"))
		}
	}
	for i, s := range result.Target {
		log.Debug(s)
		ss := strings.SplitN(s, "/", 2)
		if len(ss) == 1 {
			return oderr.NewRequestInvalidErr(fmt.Errorf(
				`"targets" should be of format 'user/{id}', got %#v`, s))
		}
		if ss[0] != "user" {
			return oderr.NewRequestInvalidErr(fmt.Errorf(
				`"targets" should be of format 'user/{id}', got %#v`, s))
		}
		result.Target[i] = ss[1]
	}
	return nil
}

// DeviceReigsterResult is the result put onto response.Result on
// successful call of DeviceRegisterHandler
// type DeviceReigsterResult struct {
// 	ID string `json:"id"`
// }

// RelationQueryHandler query user from current users' relation
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "relation:query",
//     "access_token": "ACCESS_TOKEN",
//     "type": "follow",
//     "direction": "active"
// }
// EOF
func RelationQueryHandler(rpayload *router.Payload, response *router.Response) {
	log.Debug("RelationQueryHandler")
	payload := relationPayload{}
	if err := relationColander(rpayload.Data, &payload); err != nil {
		response.Err = err
		return
	}
	result := rpayload.DBConn.QueryRelation(
		rpayload.UserInfoID, payload.Name, payload.Direction)
	response.Result = result
}

// RelationAddHandler add current user relation
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "relation:add",
//     "access_token": "ACCESS_TOKEN",
//     "type": "follow",
//     "targets": [
//         "user/1001",
//         "user/1002"
//     ]
// }
// EOF
//
// {
//     "request_id": "REQUEST_ID",
//     "result": [
//         {
//             "_id": "user/1001",
//         },
//         {
//             "_id": "user/1002",
//             "_type": "error",
//             "message": "cannot find user"
//         }
//     ]
// }
func RelationAddHandler(rpayload *router.Payload, response *router.Response) {
	log.Debug("RelationAddHandler")
	payload := relationPayload{}
	if err := relationColander(rpayload.Data, &payload); err != nil {
		response.Err = err
		return
	}
	results := []interface{}{}
	for s := range payload.Target {
		target := payload.Target[s]
		err := rpayload.DBConn.AddRelation(rpayload.UserInfoID, payload.Name, target)
		if err != nil {
			log.WithFields(log.Fields{
				"target": target,
				"err":    err,
			}).Debugln("failed to add relation")
			results = append(results, map[string]interface{}{
				"_id":     target,
				"_type":   "error",
				"message": err.Error(),
			})
		} else {
			results = append(results, struct {
				ID string `json:"_id"`
			}{"user/" + target})
		}
	}
	response.Result = results
}

// RelationRemoveHandler remove a users' relation to other users
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "relation:remove",
//     "access_token": "ACCESS_TOKEN",
//     "type": "follow",
//     "targets": [
//         "user/1001",
//         "user/1002"
//     ]
// }
// EOF
func RelationRemoveHandler(rpayload *router.Payload, response *router.Response) {
	log.Debug("RelationRemoveHandler")
	payload := relationPayload{}
	if err := relationColander(rpayload.Data, &payload); err != nil {
		response.Err = err
		return
	}
	results := []interface{}{}
	for s := range payload.Target {
		target := payload.Target[s]
		err := rpayload.DBConn.RemoveRelation(rpayload.UserInfoID, payload.Name, target)
		if err != nil {
			log.WithFields(log.Fields{
				"target": target,
				"err":    err,
			}).Debugln("failed to remmove user")
			results = append(results, map[string]interface{}{
				"_id":     target,
				"_type":   "error",
				"message": err.Error(),
			})
		} else {
			results = append(results, struct {
				ID string `json:"_id"`
			}{"user/" + target})
		}
	}
	response.Result = results
}
