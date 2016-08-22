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
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"

	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

type deviceRegisterPayload struct {
	ID          string
	Type        string
	DeviceToken string `mapstructure:"device_token"`
}

func (payload *deviceRegisterPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *deviceRegisterPayload) Validate() skyerr.Error {
	if payload.Type == "" {
		return skyerr.NewInvalidArgument("empty device type", []string{"type"})
	} else if payload.Type != "ios" && payload.Type != "android" {
		return skyerr.NewInvalidArgument(fmt.Sprintf("unknown device type = %v", payload.Type), []string{"type"})
	}

	return nil
}

// DeviceReigsterResult is the result put onto response.Result on
// successful call of DeviceRegisterHandler
type DeviceReigsterResult struct {
	ID string `json:"id"`
}

// DeviceRegisterHandler creates or updates a device and associates it to a user
//
// Example to create a new device:
//
//	curl -X POST -H "Content-Type: application/json" \
//	  -d @- http://localhost:3000/ <<EOF
//	{
//		"action": "device:register",
//		"access_token": "some-access-token",
//		"type": "ios",
//		"device_token": "some-device-token"
//	}
//	EOF
//
// Example to update an existing device:
//
//	curl -X POST -H "Content-Type: application/json" \
//	  -d @- http://localhost:3000/ <<EOF
//	{
//		"action": "device:register",
//		"access_token": "some-access-token",
//		"id": "existing-device-id",
//		"type": "ios",
//		"device_token": "new-device-token"
//	}
//	EOF
//
type DeviceRegisterHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireUser   router.Processor `preprocessor:"require_user"`
	preprocessors []router.Processor
}

func (h *DeviceRegisterHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
		h.InjectDB,
		h.RequireUser,
	}
}

func (h *DeviceRegisterHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *DeviceRegisterHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := deviceRegisterPayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	conn := rpayload.DBConn

	device := skydb.Device{}
	deviceID := payload.ID
	if deviceID == "" { // new device
		device.ID = uuid.New()
	} else { // update device
		if err := conn.GetDevice(deviceID, &device); err != nil {
			var errToReturn skyerr.Error
			if err == skydb.ErrDeviceNotFound {
				errToReturn = skyerr.NewError(skyerr.ResourceNotFound, "device not found")
			} else {
				log.WithFields(logrus.Fields{
					"deviceID": deviceID,
					"device":   device,
					"err":      err,
				}).Errorln("Failed to get device")

				errToReturn = skyerr.NewResourceFetchFailureErr("device", deviceID)
			}
			response.Err = errToReturn
			return
		}
	}

	// delete all all devices with the same token
	if err := conn.DeleteDevicesByToken(payload.DeviceToken, skydb.ZeroTime); err != nil {
		if err != skydb.ErrDeviceNotFound {
			response.Err = skyerr.NewResourceDeleteFailureErrWithStringID("device", "")
			return
		}
	}

	userinfoID := rpayload.UserInfoID

	device.Type = payload.Type
	device.Token = payload.DeviceToken
	device.UserInfoID = userinfoID
	device.LastRegisteredAt = timeNow()

	if err := conn.SaveDevice(&device); err != nil {
		log.WithFields(logrus.Fields{
			"deviceID": deviceID,
			"device":   device,
			"err":      err,
		}).Errorln("Failed to save device")

		response.Err = skyerr.NewResourceSaveFailureErrWithStringID("device", deviceID)
	} else {
		response.Result = DeviceReigsterResult{device.ID}
	}
}
