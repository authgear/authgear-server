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
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/authtoken"
	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// Define the playload that disable user handler will process
type setDisableUserPayload struct {
	AuthInfoID   string `mapstructure:"auth_id"`
	Disabled     bool   `mapstructure:"disabled"`
	Message      string `mapstructure:"message"`
	ExpiryString string `mapstructure:"expiry"`
	expiry       *time.Time
}

func (payload *setDisableUserPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	if !payload.Disabled {
		payload.Message = ""
		payload.ExpiryString = ""
		payload.expiry = nil
		return nil
	}

	if payload.ExpiryString != "" {
		if expiry, err := time.Parse(time.RFC3339, payload.ExpiryString); err == nil {
			payload.expiry = &expiry
		} else {
			return skyerr.NewInvalidArgument("invalid expiry", []string{"expiry"})
		}
	}
	return payload.Validate()
}

func (payload *setDisableUserPayload) Validate() skyerr.Error {
	return nil
}

// SetDisableUserHandler set disabled flag for the specified user
//
// SetDisableUserHandler receives these parameters:
//
// * auth_id (string, required)
// * disabled (boolean, required)
// * message (string, optional)
// * expiry (date/time, optional)
//
// Current implementation:
//
// ```
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "auth:disable:set",
//     "auth_id": "77FA8BCF-CD6C-4A22-A170-CECC2667654F"
// }
// EOF
// ```
//
// Response:
// * success response
type SetDisableUserHandler struct {
	TokenStore    authtoken.Store  `inject:"TokenStore"`
	AssetStore    asset.Store      `inject:"AssetStore"`
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectAuth    router.Processor `preprocessor:"inject_auth"`
	RequireAdmin  router.Processor `preprocessor:"require_admin"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *SetDisableUserHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectAuth,
		h.RequireAdmin,
		h.PluginReady,
	}
}

func (h *SetDisableUserHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *SetDisableUserHandler) Handle(payload *router.Payload, response *router.Response) {
	logger := logging.CreateLogger(payload.Context(), "handler")
	p := &setDisableUserPayload{}
	skyErr := p.Decode(payload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"auth_id": p.AuthInfoID,
	})
	logger.Debug("Handler called to set disabled user status")

	authinfo := skydb.AuthInfo{}
	if err := payload.DBConn.GetAuth(p.AuthInfoID, &authinfo); err != nil {
		if err == skydb.ErrUserNotFound {
			logger.Info("Auth info not found when setting disabled user status")
			response.Err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
			return
		}
		logger.WithError(err).Error("Unable to get auth info when setting disabled user status")
		response.Err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
		return
	}

	authinfo.Disabled = p.Disabled
	authinfo.DisabledMessage = p.Message
	authinfo.DisabledExpiry = p.expiry

	logger.WithFields(logrus.Fields{
		"disabled": authinfo.Disabled,
		"message":  authinfo.DisabledMessage,
		"expiry":   authinfo.DisabledExpiry,
	}).Debug("Will set disabled user status")

	if err := payload.DBConn.UpdateAuth(&authinfo); err != nil {
		logger.WithError(err).Error("Unable to update auth info when setting disabled user status")
		response.Err = skyerr.MakeError(err)
		return
	}

	logger.Info("Successfully set disabled user status")

	h.logAuditTrail(payload, p)

	response.Result = statusResponse{
		Status: "OK",
	}
}

func (h *SetDisableUserHandler) logAuditTrail(payload *router.Payload, p *setDisableUserPayload) {
	var event audit.Event
	if p.Disabled {
		event = audit.EventDisableUser
	} else {
		event = audit.EventEnableUser
	}

	audit.Trail(audit.Entry{
		AuthID: p.AuthInfoID,
		Event:  event,
	}.WithRouterPayload(payload))
}
