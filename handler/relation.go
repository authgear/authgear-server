package handler

import (
	log "github.com/Sirupsen/logrus"

	"github.com/mitchellh/mapstructure"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
)

type relationPayload struct {
	Name      string   `json:"name"`
	Direction string   `json:"direction"`
	Target    []string `json:"targets"`

	Limit  uint64 `json:"limit"`
	Offset uint64 `json:"offset"`
}

func relationColander(data map[string]interface{}, result *relationPayload) skyerr.Error {
	mapDecoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  result,
		TagName: "json",
	})
	if err := mapDecoder.Decode(data); err != nil {
		return skyerr.NewError(skyerr.BadRequest, err.Error())
	}
	relationMap := map[string]string{
		"friend":  "_friend",
		"_friend": "_friend",
		"follow":  "_follow",
		"_follow": "_follow",
	}
	relationName, ok := relationMap[result.Name]
	if !ok {
		return skyerr.NewError(skyerr.NotSupported, "Only friend and follow relation is supported")
	}
	result.Name = relationName
	if result.Direction != "" {
		if result.Direction != "outward" && result.Direction != "inward" && result.Direction != "mutual" {
			return skyerr.NewError(skyerr.InvalidArgument, "Only outward, inward and mutual direction is allowed")
		}
	}
	return nil
}

// RelationQueryHandler query user from current users' relation
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "relation:query",
//     "access_token": "ACCESS_TOKEN",
//     "name": "follow",
//     "direction": "outward"
//	   "limit": 2
//	   "offset": 0
// }
// EOF
//
// {
//     "request_id": "REQUEST_ID",
//     "result": [
//         {
//             "id": "1001",
//             "type": "user",
//             "data": {
//                 "_id": "1001",
//                 "username": "user1001",
//                 "email": "user1001@skygear.io"
//             }
//         },
//         {
//             "id": "1002",
//             "type": "user",
//             "data": {
//                 "_id": "1002",
//                 "username": "user1002",
//                 "email": "user1001@skygear.io"
//             }
//         }
//     ],
//     "info": {
//         "count": 2
//     }
// }
type RelationQueryHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *RelationQueryHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
	}
}

func (h *RelationQueryHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RelationQueryHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debug("RelationQueryHandler")
	payload := relationPayload{}
	if err := relationColander(rpayload.Data, &payload); err != nil {
		response.Err = err
		return
	}
	result := rpayload.DBConn.QueryRelation(
		rpayload.UserInfoID, payload.Name, payload.Direction, skydb.QueryConfig{
			Limit:  payload.Limit,
			Offset: payload.Offset,
		})
	resultList := make([]interface{}, 0, len(result))
	for _, userinfo := range result {
		resultList = append(resultList, struct {
			ID   string      `json:"id"`
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}{userinfo.ID, "user", userinfo})
	}
	response.Result = resultList
	count, countErr := rpayload.DBConn.QueryRelationCount(
		rpayload.UserInfoID, payload.Name, payload.Direction)
	if countErr != nil {
		log.WithFields(log.Fields{
			"err": countErr,
		}).Warnf("Relation Count Query fails")
		count = 0
	}
	response.Info = struct {
		Count uint64 `json:"count"`
	}{
		count,
	}
}

// RelationAddHandler add current user relation
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "relation:add",
//     "access_token": "ACCESS_TOKEN",
//     "name": "follow",
//     "targets": [
//         "1001",
//         "1002"
//     ]
// }
// EOF
//
// {
//     "request_id": "REQUEST_ID",
//     "result": [
//         {
//             "id": "1001",
//             "type": "user",
//             "data": {
//                 "_id": "1001",
//                 "username": "user1001",
//                 "email": "user1001@skygear.io"
//             }
//         },
//         {
//             "id": "1002",
//             "type": "error",
//             "data": {
//                 "type": "ResourceFetchFailure",
//                 "code": 101,
//                 "message": "failed to fetch user id = 1002"
//             }
//         }
//     ]
// }
type RelationAddHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *RelationAddHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
	}
}

func (h *RelationAddHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RelationAddHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debug("RelationAddHandler")
	payload := relationPayload{}
	if err := relationColander(rpayload.Data, &payload); err != nil {
		response.Err = err
		return
	}
	results := make([]interface{}, 0, len(payload.Target))
	for s := range payload.Target {
		target := payload.Target[s]
		err := rpayload.DBConn.AddRelation(rpayload.UserInfoID, payload.Name, target)
		if err != nil {
			log.WithFields(log.Fields{
				"target": target,
				"err":    err,
			}).Debugln("failed to add relation")
			results = append(results, struct {
				ID   string       `json:"id"`
				Type string       `json:"type"`
				Data skyerr.Error `json:"data"`
			}{target, "error", skyerr.NewResourceFetchFailureErr("user", target)})
		} else {
			userinfo := skydb.UserInfo{}
			rpayload.DBConn.GetUser(target, &userinfo)
			userinfo.HashedPassword = []byte{}
			results = append(results, struct {
				ID   string      `json:"id"`
				Type string      `json:"type"`
				Data interface{} `json:"data"`
			}{target, "user", userinfo})
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
//     "name": "follow",
//     "targets": [
//         "1001",
//         "1002"
//     ]
// }
// EOF
type RelationRemoveHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	preprocessors []router.Processor
}

func (h *RelationRemoveHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
	}
}

func (h *RelationRemoveHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RelationRemoveHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debug("RelationRemoveHandler")
	payload := relationPayload{}
	if err := relationColander(rpayload.Data, &payload); err != nil {
		response.Err = err
		return
	}
	results := make([]interface{}, 0, len(payload.Target))
	for s := range payload.Target {
		target := payload.Target[s]
		err := rpayload.DBConn.RemoveRelation(rpayload.UserInfoID, payload.Name, target)
		if err != nil {
			log.WithFields(log.Fields{
				"target": target,
				"err":    err,
			}).Debugln("failed to remmove user")
			results = append(results, struct {
				ID   string      `json:"id"`
				Type string      `json:"type"`
				Data interface{} `json:"data"`
			}{target, "error", err})
		} else {
			results = append(results, struct {
				ID string `json:"id"`
			}{target})
		}
	}
	response.Result = results
}
