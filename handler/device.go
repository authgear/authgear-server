package handler

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
	"github.com/oursky/skygear/uuid"
)

type deviceRegisterPayload struct {
	ID          string
	Type        string
	DeviceToken string `mapstructure:"device_token"`
}

func (p *deviceRegisterPayload) Validate() error {
	if p.Type == "" {
		return errors.New("empty device type")
	} else if p.Type != "ios" && p.Type != "android" {
		return fmt.Errorf("unknown device type = %v", p.Type)
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
func DeviceRegisterHandler(rpayload *router.Payload, response *router.Response) {
	payload := deviceRegisterPayload{}
	if err := mapstructure.Decode(rpayload.Data, &payload); err != nil {
		response.Err = skyerr.NewRequestInvalidErr(err)
		return
	}
	if err := payload.Validate(); err != nil {
		response.Err = skyerr.NewRequestInvalidErr(err)
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
				errToReturn = skyerr.ErrDeviceNotFound
			} else {
				log.WithFields(log.Fields{
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

	userinfoID := rpayload.UserInfoID

	device.Type = payload.Type
	device.Token = payload.DeviceToken
	device.UserInfoID = userinfoID
	device.LastRegisteredAt = timeNow()

	if err := conn.SaveDevice(&device); err != nil {
		log.WithFields(log.Fields{
			"deviceID": deviceID,
			"device":   device,
			"err":      err,
		}).Errorln("Failed to save device")

		response.Err = skyerr.NewResourceSaveFailureErrWithStringID("device", deviceID)
	} else {
		response.Result = DeviceReigsterResult{device.ID}
	}
}
