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

	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type rolePayload struct {
	Roles []string `mapstructure:"roles"`
}

func (payload *rolePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *rolePayload) Validate() skyerr.Error {
	if payload.Roles == nil {
		return skyerr.NewInvalidArgument("unspecified roles in request", []string{"roles"})
	}
	return nil
}

// RoleDefaultHandler enable system administrator to set default user role
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "role:default",
//     "master_key": "MASTER_KEY",
//     "access_token": "ACCESS_TOKEN",
//     "roles": [
//        "writer",
//        "user"
//     ]
// }
// EOF
//
// {
//     "result": [
//        "writer",
//        "user"
//     ]
// }
type RoleDefaultHandler struct {
	AccessKey     router.Processor `preprocessor:"accesskey"`
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *RoleDefaultHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DevOnly,
		h.DBConn,
		h.PluginReady,
	}
}

func (h *RoleDefaultHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RoleDefaultHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debugf("RoleDefaultHandler %v", h)
	payload := &rolePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	err := rpayload.DBConn.SetDefaultRoles(payload.Roles)
	if err != nil {
		response.Err = skyerr.MakeError(err)
	}
	response.Result = payload.Roles
}

// RoleAdminHandler enable system administrator to set which roles can perform
// administrative action, like change others user role.
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "role:default",
//     "master_key": "MASTER_KEY",
//     "access_token": "ACCESS_TOKEN",
//     "roles": [
//        "admin",
//        "moderator"
//     ]
// }
// EOF
//
// {
//     "result": [
//        "admin",
//        "moderator"
//     ]
// }
type RoleAdminHandler struct {
	AccessKey     router.Processor `preprocessor:"accesskey"`
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *RoleAdminHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DevOnly,
		h.DBConn,
		h.PluginReady,
	}
}

func (h *RoleAdminHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RoleAdminHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debugf("RoleAdminHandler %v", h)
	payload := &rolePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	err := rpayload.DBConn.SetAdminRoles(payload.Roles)
	if err != nil {
		response.Err = skyerr.MakeError(err)
	}
	response.Result = payload.Roles
}

type roleBatchPayload struct {
	Roles   []string `mapstructure:"roles"`
	UserIDs []string `mapstructure:"users"`
}

func (payload *roleBatchPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *roleBatchPayload) Validate() skyerr.Error {
	if payload.Roles == nil {
		return skyerr.NewInvalidArgument("unspecified roles in request", []string{"roles"})
	}
	if payload.UserIDs == nil {
		return skyerr.NewInvalidArgument("unspecified users in request", []string{"users"})
	}
	return nil
}

// RoleAssignHandler allow system administractor to batch assign roles to
// users
//
// RoleAssignHandler required user with admin role.
// All specified users will assign to all roles specified. Roles not already
// exisited in DB will be created. Users not already existed will be ignored.
//
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "role:assign",
//     "master_key": "MASTER_KEY",
//     "access_token": "ACCESS_TOKEN",
//     "roles": [
//        "writer",
//        "user"
//     ],
//     "users": [
//        "95db1e34-0cc0-47b0-8a97-3948633ce09f",
//        "3df4b52b-bd58-4fa2-8aee-3d44fd7f974d"
//     ]
// }
// EOF
//
// {
//     "result": "OK"
// }
type RoleAssignHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	RequireAdmin  router.Processor `preprocessor:"require_admin"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *RoleAssignHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.RequireAdmin,
		h.PluginReady,
	}
}

func (h *RoleAssignHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RoleAssignHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debugf("RoleAssignHandler %v", h)
	payload := &roleBatchPayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if err := rpayload.DBConn.AssignRoles(payload.UserIDs, payload.Roles); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}
	response.Result = "OK"
}

// RoleRevokeHandler allow system administractor to batch revoke roles from
// users
//
// RoleRevokeHandler required user with admin role.
// All specified users will have all specified roles revoked. Roles or users
// not already exisited in DB will be ignored.
//
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "role:revoke",
//     "master_key": "MASTER_KEY",
//     "access_token": "ACCESS_TOKEN",
//     "roles": [
//        "writer",
//        "user"
//     ],
//     "users": [
//        "95db1e34-0cc0-47b0-8a97-3948633ce09f",
//        "3df4b52b-bd58-4fa2-8aee-3d44fd7f974d"
//     ]
// }
// EOF
//
// {
//     "result": "OK"
// }
type RoleRevokeHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	RequireAdmin  router.Processor `preprocessor:"require_admin"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *RoleRevokeHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.RequireAdmin,
		h.PluginReady,
	}
}

func (h *RoleRevokeHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RoleRevokeHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debugf("RoleRevokeHandler %v", h)
	payload := &roleBatchPayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	if err := rpayload.DBConn.RevokeRoles(payload.UserIDs, payload.Roles); err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}
	response.Result = "OK"
}
