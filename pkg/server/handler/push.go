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
	"github.com/skygeario/skygear-server/pkg/server/push"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

var sendPushNotification = func(sender push.Sender, device skydb.Device, m push.Mapper) {
	go func() {
		log.Debugf("Sending notification to device token = %s", device.Token)
		err := sender.Send(m, device)

		if err != nil {
			log.Warnf("Failed to send notification: %v\n", err)
		} else {
			log.Debugf("Sent notification to device token = %s", device.Token)
		}
	}()
}

type sendPushResponseItem struct {
	id  string
	err *error
}

func (e *sendPushResponseItem) MarshalJSON() ([]byte, error) {
	var err skyerr.Error
	if e.err != nil && *e.err == skydb.ErrDeviceNotFound {
		err = skyerr.NewErrorWithInfo(skyerr.ResourceNotFound, fmt.Sprintf(`cannot find device "%s"`, e.id), map[string]interface{}{"id": e.id})
	} else if e.err != nil && *e.err == skydb.ErrUserNotFound {
		err = skyerr.NewErrorWithInfo(skyerr.ResourceNotFound, fmt.Sprintf(`cannot find user "%s"`, e.id), map[string]interface{}{"id": e.id})
	} else if e.err != nil {
		err = skyerr.NewError(skyerr.UnexpectedError, fmt.Sprintf("unknown error occurred: %v", (*e.err).Error()))
	}
	if e.err != nil {
		return json.Marshal(&struct {
			ID       string                 `json:"_id"`
			ItemType string                 `json:"_type"`
			Message  string                 `json:"message"`
			Name     string                 `json:"name"`
			Code     skyerr.ErrorCode       `json:"code"`
			Info     map[string]interface{} `json:"info,omitempty"`
		}{e.id, "error", err.Message(), err.Name(), err.Code(), err.Info()})
	}
	return json.Marshal(&struct {
		ID string `json:"_id"`
	}{e.id})
}

type pushToUserPayload struct {
	UserIDs      []string               `mapstructure:"user_ids"`
	Notification map[string]interface{} `mapstructure:"notification"`
}

func (payload *pushToUserPayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *pushToUserPayload) Validate() skyerr.Error {
	if len(payload.UserIDs) == 0 {
		return skyerr.NewInvalidArgument("empty user ids", []string{"user_ids"})
	}
	if payload.Notification == nil {
		return skyerr.NewInvalidArgument("no notification specified", []string{"notification"})
	}
	return nil
}

type PushToUserHandler struct {
	NotificationSender push.Sender      `inject:"PushSender"`
	AccessKey          router.Processor `preprocessor:"accesskey"`
	DBConn             router.Processor `preprocessor:"dbconn"`
	InjectDB           router.Processor `preprocessor:"inject_db"`
	Notification       router.Processor `preprocessor:"notification"`
	preprocessors      []router.Processor
}

func (h *PushToUserHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.InjectDB,
		h.Notification,
	}
}

func (h *PushToUserHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *PushToUserHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := pushToUserPayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	conn := rpayload.DBConn
	resultItems := make([]sendPushResponseItem, len(payload.UserIDs))
	for i, userID := range payload.UserIDs {
		resultItems[i].id = userID
		devices, err := conn.QueryDevicesByUser(userID)
		if err != nil {
			resultItems[i].err = &err
		} else {
			// FIXME: The deduplication should be done at device register.
			deviceIDs := map[string]bool{}
			for i := range devices {
				device := devices[i]
				if _, ok := deviceIDs[device.Token]; !ok {
					deviceIDs[device.Token] = true
					pushMap := push.MapMapper(payload.Notification)
					sendPushNotification(h.NotificationSender, device, pushMap)
				}
			}
		}
	}
	response.Result = resultItems
}

type pushToDevicePayload struct {
	DeviceIDs    []string               `mapstructure:"device_ids"`
	Notification map[string]interface{} `mapstructure:"notification"`
}

func (payload *pushToDevicePayload) Decode(data map[string]interface{}) skyerr.Error {
	if err := mapstructure.Decode(data, payload); err != nil {
		return skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}
	return payload.Validate()
}

func (payload *pushToDevicePayload) Validate() skyerr.Error {
	if len(payload.DeviceIDs) == 0 {
		return skyerr.NewInvalidArgument("empty device ids", []string{"device_ids"})
	}
	if payload.Notification == nil {
		return skyerr.NewInvalidArgument("no notification specified", []string{"notification"})
	}
	return nil
}

type PushToDeviceHandler struct {
	NotificationSender push.Sender      `inject:"PushSender"`
	AccessKey          router.Processor `preprocessor:"accesskey"`
	DBConn             router.Processor `preprocessor:"dbconn"`
	InjectDB           router.Processor `preprocessor:"inject_db"`
	Notification       router.Processor `preprocessor:"notification"`
	preprocessors      []router.Processor
}

func (h *PushToDeviceHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
		h.DBConn,
		h.InjectDB,
		h.Notification,
	}
}

func (h *PushToDeviceHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *PushToDeviceHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := &pushToDevicePayload{}
	skyErr := payload.Decode(rpayload.Data)
	if skyErr != nil {
		response.Err = skyErr
		return
	}

	conn := rpayload.DBConn
	resultItems := make([]sendPushResponseItem, len(payload.DeviceIDs))
	for i, deviceID := range payload.DeviceIDs {
		device := skydb.Device{}
		resultItems[i].id = deviceID
		if err := conn.GetDevice(deviceID, &device); err != nil {
			resultItems[i].err = &err
		} else {
			pushMap := push.MapMapper(payload.Notification)
			sendPushNotification(h.NotificationSender, device, pushMap)
		}
	}
	response.Result = resultItems
}
