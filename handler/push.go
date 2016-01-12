package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/oursky/skygear/push"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
)

type pushToDevicePayload struct {
	DeviceIDs    []string               `mapstructure:"device_ids"`
	Notification map[string]interface{} `mapstructure:"notification"`
}

func (p *pushToDevicePayload) Validate() error {
	if len(p.DeviceIDs) == 0 {
		return errors.New("empty device ids")
	}
	if p.Notification == nil {
		return errors.New("no notification specified")
	}
	return nil
}

type pushToUserPayload struct {
	UserIDs      []string               `mapstructure:"user_ids"`
	Notification map[string]interface{} `mapstructure:"notification"`
}

func (p *pushToUserPayload) Validate() error {
	if len(p.UserIDs) == 0 {
		return errors.New("empty user ids")
	}
	if p.Notification == nil {
		return errors.New("no notification specified")
	}
	return nil
}

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

type PushToUserHandler struct {
	NotificationSender push.Sender `inject:"PushSender"`
}

func (h *PushToUserHandler) Setup() {
	return
}

func (h *PushToUserHandler) GetPreprocessors() []router.Processor {
	return nil
}

func (h *PushToUserHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := pushToUserPayload{}
	if err := mapstructure.Decode(rpayload.Data, &payload); err != nil {
		response.Err = skyerr.NewError(skyerr.BadRequest, err.Error())
		return
	}
	if err := payload.Validate(); err != nil {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, err.Error())
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

type PushToDeviceHandler struct {
	NotificationSender push.Sender `inject:"PushSender"`
}

func (h *PushToDeviceHandler) Setup() {
	return
}

func (h *PushToDeviceHandler) GetPreprocessors() []router.Processor {
	return nil
}

func (h *PushToDeviceHandler) Handle(rpayload *router.Payload, response *router.Response) {
	payload := pushToDevicePayload{}
	if err := mapstructure.Decode(rpayload.Data, &payload); err != nil {
		response.Err = skyerr.NewError(skyerr.BadRequest, err.Error())
		return
	}
	if err := payload.Validate(); err != nil {
		response.Err = skyerr.NewError(skyerr.InvalidArgument, err.Error())
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
