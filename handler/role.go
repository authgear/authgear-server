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
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"

	"github.com/skygeario/skygear-server/router"
	"github.com/skygeario/skygear-server/skyerr"
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
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

func (h *RoleDefaultHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
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
		response.Err = skyerr.NewUnknownErr(err)
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
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

func (h *RoleAdminHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
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
		response.Err = skyerr.NewUnknownErr(err)
	}
	response.Result = payload.Roles
}
