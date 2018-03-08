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
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// Define the playload that set verify user handler will process
type setVerifyUserPayload struct {
	AuthInfoID string `mapstructure:"auth_id"`
	Verified   bool   `mapstructure:"verified"`
}

func (payload *setVerifyUserPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *setVerifyUserPayload) Validate() skyerr.Error {
	return nil
}

// SetVerifyUserHandler set verified flag for the specified user
//
// SetVerifyUserHandler receives one parameter:
//
// * auth_id (string, required)
// * verified (bool)
//
// Current implementation:
//
// ```
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "auth:verify:set",
//     "auth_id": "77FA8BCF-CD6C-4A22-A170-CECC2667654F",
//     "verified": true
// }
// EOF
// ```
//
// Response:
// * success response
type SetVerifyUserHandler struct {
	TokenStore    authtoken.Store  `inject:"TokenStore"`
	AssetStore    asset.Store      `inject:"AssetStore"`
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectAuth    router.Processor `preprocessor:"inject_auth"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	RequireAdmin  router.Processor `preprocessor:"require_admin"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *SetVerifyUserHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectAuth,
		h.InjectUser,
		h.RequireAdmin,
		h.PluginReady,
	}
}

func (h *SetVerifyUserHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SetVerifyUserHandler) Handle(payload *router.Payload, response *router.Response) {
	p := &setVerifyUserPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	log = log.WithFields(logrus.Fields{
		"auth_id":  p.AuthInfoID,
		"verified": p.Verified,
	})
	log.Debug("Handler called to set verified user status")

	authinfo := skydb.AuthInfo{}
	if err := payload.DBConn.GetAuth(p.AuthInfoID, &authinfo); err != nil {
		if err == skydb.ErrUserNotFound {
			log.Info("Auth info not found when setting verified user status")
			response.Err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
			return
		}
		log.WithError(err).Error("Unable to get auth info when setting verified user status")
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
		return
	}

	authinfo.Verified = p.Verified

	log.WithFields(logrus.Fields{
		"verified": p.Verified,
	}).Debug("Will set verified user status")

	if err := payload.DBConn.UpdateAuth(&authinfo); err != nil {
		log.WithError(err).Error("Unable to update auth info when set verified user status")
		response.Err = skyerr.MakeError(err)
		return
	}

	log.Info("Successfully set verified user status")

	h.logAuditTrail(payload, p)

	response.Result = statusResponse{
		Status: "OK",
	}
}

func (h *SetVerifyUserHandler) logAuditTrail(payload *router.Payload, p *setVerifyUserPayload) {
	var event audit.Event
	if p.Verified {
		event = audit.EventVerifyUser
	} else {
		event = audit.EventUnverifyUser
	}

	audit.Trail(audit.Entry{
		AuthID: p.AuthInfoID,
		Event:  event,
	}.WithRouterPayload(payload))

}
