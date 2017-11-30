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

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

type deviceRegisterPayload struct {
	ID          string
	Type        string
	Topic       string
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

type deviceUnregisterPayload struct {
	ID string
}

func (payload *deviceUnregisterPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *deviceUnregisterPayload) Validate() skyerr.Error {
	if payload.ID == "" {
		return skyerr.NewInvalidArgument("Missing device id", []string{"id"})
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
//		"topic": "io.skygear.sample.topic",
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
	InjectAuth    router.Processor `preprocessor:"inject_auth"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireAuth   router.Processor `preprocessor:"require_auth"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *DeviceRegisterHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectAuth,
		h.InjectDB,
		h.RequireAuth,
		h.PluginReady,
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
			if err == skydb.ErrDeviceNotFound {
				response.Err = skyerr.NewError(skyerr.ResourceNotFound, "Device not found")
				return
			}

			log.WithFields(logrus.Fields{
				"deviceID": deviceID,
				"err":      err,
			}).Errorln("Fail to get device")

			response.Err = skyerr.NewResourceFetchFailureErr("device", deviceID)
			return
		}
	}

	// delete all devices with the same token
	if err := conn.DeleteDevicesByToken(payload.DeviceToken, skydb.ZeroTime); err != nil {
		if err != skydb.ErrDeviceNotFound {
			response.Err = skyerr.NewResourceDeleteFailureErrWithStringID("device", "")
			return
		}
	}

	device.Type = payload.Type
	device.Token = payload.DeviceToken
	device.Topic = payload.Topic
	device.AuthInfoID = rpayload.AuthInfoID
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

// DeviceUnregisterHandler removes user id from a device
//
// Example to unregister a device:
//
//	curl -X POST -H "Content-Type: application/json" \
//	  -d @- http://localhost:3000/ <<EOF
//	{
//		"action": "device:unregister",
//		"access_token": "some-access-token",
//		"type": "ios",
//		"device_token": "some-device-token"
//	}
//	EOF
//
type DeviceUnregisterHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectAuth    router.Processor `preprocessor:"inject_auth"`
	InjectDB      router.Processor `preprocessor:"inject_db"`
	RequireAuth   router.Processor `preprocessor:"require_auth"`
	PluginReady   router.Processor `preprocessor:"plugin_ready"`
	preprocessors []router.Processor
}

func (h *DeviceUnregisterHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectAuth,
		h.InjectDB,
		h.RequireAuth,
		h.PluginReady,
	}
}

func (h *DeviceUnregisterHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *DeviceUnregisterHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := deviceUnregisterPayload{}
	if err := payload.Decode(rpayload.Data); err != nil {
		response.Err = err
		return
	}

	conn := rpayload.DBConn

	device := skydb.Device{}
	if err := conn.GetDevice(payload.ID, &device); err != nil {
		if err == skydb.ErrDeviceNotFound {
			response.Err = skyerr.NewError(skyerr.ResourceNotFound, "Device not found")
			return
		}

		log.WithFields(logrus.Fields{
			"deviceID": payload.ID,
			"err":      err,
		}).Errorln("Fail to get device")

		response.Err = skyerr.NewResourceFetchFailureErr("device", payload.ID)
		return
	}

	// delete all devices with the same token
	if err := conn.DeleteDevicesByToken(device.Token, skydb.ZeroTime); err != nil {
		if err != skydb.ErrDeviceNotFound {
			response.Err = skyerr.NewResourceDeleteFailureErrWithStringID("device", "")
			return
		}
	}

	device.AuthInfoID = ""
	if err := conn.SaveDevice(&device); err != nil {
		log.WithFields(logrus.Fields{
			"deviceID": payload.ID,
			"device":   device,
			"err":      err,
		}).Errorln("Fail to save device")

		response.Err = skyerr.NewResourceSaveFailureErrWithStringID("device", payload.ID)
		return
	}

	response.Result = DeviceReigsterResult{device.ID}
}
