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
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/recordutil"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func canonicalRelationName(name string) (string, skyerr.Error) {
	relationMap := map[string]string{
		"friend":  "_friend",
		"_friend": "_friend",
		"follow":  "_follow",
		"_follow": "_follow",
	}
	relationName, ok := relationMap[name]
	if !ok {
		return "", skyerr.NewError(skyerr.NotSupported, "Only friend and follow relation is supported")
	}
	return relationName, nil
}

type relationQueryPayload struct {
	Name      string `mapstructure:"name"`
	Direction string `mapstructure:"direction"`

	Limit  uint64 `mapstructure:"limit"`
	Offset uint64 `mapstructure:"offset"`
}

func (payload *relationQueryPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *relationQueryPayload) Validate() skyerr.Error {
	relationName, err := canonicalRelationName(payload.Name)
	if err != nil {
		return err
	}
	payload.Name = relationName

	if payload.Direction != "" && payload.Direction != "outward" && payload.Direction != "inward" && payload.Direction != "mutual" {
		return skyerr.NewInvalidArgument("only outward, inward and mutual direction is allowed", []string{"direction"})
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
//                 "_type": "record",
//                 "_id": "user/1001",
//                 "_access": null,
//                 "username": "user1001",
//                 "email": "user1001@skygear.io"
//             }
//         },
//         {
//             "id": "1002",
//             "type": "user",
//             "data": {
//                 "_type": "record",
//                 "_id": "1002",
//                 "_access": null,
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
	AssetStore    asset.Store      `inject:"AssetStore"`
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectAuth    router.Processor `preprocessor:"inject_auth"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *RelationQueryHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectAuth,
		h.InjectDB,
		h.PluginReady,
	}
}

func (h *RelationQueryHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RelationQueryHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debug("RelationQueryHandler")
	payload := &relationQueryPayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	result := rpayload.DBConn.QueryRelation(
		rpayload.AuthInfoID, payload.Name, payload.Direction, skydb.QueryConfig{
			Limit:  payload.Limit,
			Offset: payload.Offset,
		})
	resultList := make([]interface{}, 0, len(result))
	for _, authinfo := range result {
		user, err := fetchUser(
			rpayload.Database,
			rpayload.DBConn,
			h.AssetStore,
			*rpayload.AuthInfo,
			authinfo.ID,
			rpayload.HasMasterKey(),
		)
		if err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}

		resultList = append(resultList, struct {
			ID   string      `json:"id"`
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}{authinfo.ID, "user", &user})
	}
	response.Result = resultList
	count, countErr := rpayload.DBConn.QueryRelationCount(
		rpayload.AuthInfoID, payload.Name, payload.Direction)
	if countErr != nil {
		log.WithFields(logrus.Fields{
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

// relationChangePayload is shared by RelationAddHandler and RelationRemoveHandler
type relationChangePayload struct {
	Name   string   `mapstructure:"name"`
	Target []string `mapstructure:"targets"`
}

func (payload *relationChangePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *relationChangePayload) Validate() skyerr.Error {
	relationName, err := canonicalRelationName(payload.Name)
	if err != nil {
		return err
	}
	payload.Name = relationName
	return nil
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
//                 "_type": "record",
//                 "_id": "1001",
//                 "_access": null,
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
	AssetStore    asset.Store      `inject:"AssetStore"`
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectAuth    router.Processor `preprocessor:"inject_auth"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *RelationAddHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectAuth,
		h.InjectDB,
		h.PluginReady,
	}
}

func (h *RelationAddHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RelationAddHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debug("RelationAddHandler")
	payload := relationChangePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	results := make([]interface{}, 0, len(payload.Target))
	for s := range payload.Target {
		target := payload.Target[s]
		err := rpayload.DBConn.AddRelation(rpayload.AuthInfoID, payload.Name, target)
		if err != nil {
			log.WithFields(logrus.Fields{
				"target": target,
				"err":    err,
			}).Debugln("failed to add relation")
			results = append(results, struct {
				ID   string       `json:"id"`
				Type string       `json:"type"`
				Data skyerr.Error `json:"data"`
			}{target, "error", skyerr.NewResourceFetchFailureErr("user", target)})
		} else {
			user, err := fetchUser(
				rpayload.Database,
				rpayload.DBConn,
				h.AssetStore,
				*rpayload.AuthInfo,
				target,
				rpayload.HasMasterKey(),
			)
			if err != nil {
				response.Err = skyerr.MakeError(err)
				return
			}

			results = append(results, struct {
				ID   string      `json:"id"`
				Type string      `json:"type"`
				Data interface{} `json:"data"`
			}{target, "user", &user})
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
	InjectAuth    router.Processor `preprocessor:"inject_auth"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *RelationRemoveHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectAuth,
		h.InjectDB,
		h.PluginReady,
	}
}

func (h *RelationRemoveHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RelationRemoveHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debug("RelationRemoveHandler")
	payload := relationChangePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	results := make([]interface{}, 0, len(payload.Target))
	for s := range payload.Target {
		target := payload.Target[s]
		err := rpayload.DBConn.RemoveRelation(rpayload.AuthInfoID, payload.Name, target)
		if err != nil {
			log.WithFields(logrus.Fields{
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

func fetchUser(db skydb.Database, conn skydb.Conn, assetStore asset.Store, info skydb.AuthInfo, userID string, hasMasterKey bool) (jsonUser skyconv.JSONRecord, err error) {
	user := skydb.Record{}
	if err = db.Get(skydb.NewRecordID("user", userID), &user); err != nil {
		// TODO: handle auth without user in a better way
		return
	}

	filter, err := recordutil.NewRecordResultFilter(conn, assetStore, &info, hasMasterKey)
	if err != nil {
		return
	}

	result := filter.JSONResult(&user)
	jsonUser = *result
	return
}
