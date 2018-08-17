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

package push

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/buford/push"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

// GatewayType determine which kind of gateway should be used for APNS
type GatewayType string

// Available gateways
const (
	Sandbox    GatewayType = "sandbox"
	Production GatewayType = "production"
)

// private interface s.t. we can mock push.Service in test
type pushService interface {
	Push(deviceToken string, headers *push.Headers, payload []byte) (string, error)
}

// APNSPusher defines an interface of APNS Pusher
type APNSPusher interface {
	Sender

	Start()
	Stop()

	// for enqueuing the failed notification and delete device token if it is
	// reported as invalid
	getFailedNotificationChannel() chan failedNotification
	deleteDeviceToken(token string, beforeTime time.Time) error
}

type failedNotification struct {
	deviceToken string
	err         push.Error
}

// TODO: make something like APNSFeedbackHandler to handle the feedback from APNS so that the pusher will not have the reference to conn.
func newPushService(client *http.Client, gatewayType GatewayType) (*push.Service, error) {
	switch gatewayType {
	case Sandbox:
		return push.NewService(client, push.Development), nil
	case Production:
		return push.NewService(client, push.Production), nil
	default:
		return nil, fmt.Errorf("push/apns: unrecognized gateway type %s", gatewayType)
	}
}

func checkFailedNotifications(pusher APNSPusher) {
	for failedNotification := range pusher.getFailedNotificationChannel() {
		handleFailedNotification(pusher, failedNotification)
	}
}

func handleFailedNotification(pusher APNSPusher, failedNote failedNotification) {
	if failedNote.err.Status == http.StatusGone ||
		failedNote.err.Reason.Error() == "BadDeviceToken" {
		unregisterDevice(pusher, failedNote.deviceToken, failedNote.err.Timestamp)
	}
}

func unregisterDevice(pusher APNSPusher, deviceToken string, timestamp time.Time) {
	logger := log.WithFields(logrus.Fields{
		"deviceToken": deviceToken,
	})

	defer func() {
		if r := recover(); r != nil {
			logger.Panicf("Panic occurred while unregistering device: %s", r)
		}
	}()

	if err := pusher.deleteDeviceToken(deviceToken, timestamp); err != nil && err != skydb.ErrDeviceNotFound {
		logger.Errorf("apns/fb: failed to delete device token = %s: %v", deviceToken, err)
		return
	}

	logger.Info("Unregistered device from skydb")
}

func queueFailedNotification(pusher APNSPusher, deviceToken string, err push.Error) bool {
	logger := log.WithFields(logrus.Fields{
		"deviceToken": deviceToken,
	})

	failed := pusher.getFailedNotificationChannel()
	if failed == nil {
		logger.Warn("Unable to queue failed notification for error handling because the pusher is not running")
		return false
	}
	failed <- failedNotification{
		deviceToken: deviceToken,
		err:         err,
	}
	logger.Debug("Queued failed notification for error handling")
	return true
}
