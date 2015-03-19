package push

import (
	"github.com/anachronistic/apns"
	"log"
)

// APNSPusher pushes notification via apns
type APNSPusher struct {
	// we are directly coupling on apns as it seems redundant to duplicate
	// all the payload and client logic and interfaces.
	Client apns.APNSClient
}

// Send sends a notification to the device identified by the
// specified deviceToken
func (pusher *APNSPusher) Send(m Mapper, deviceToken string) error {
	payload := apns.NewPayload()
	payload.ContentAvailable = 1

	notification := apns.NewPushNotification()
	notification.DeviceToken = deviceToken

	// the use of Set instead of AddPayload is intentional to avoid
	// badge being reset. See the following link for details:
	//	https://github.com/anachronistic/apns/blob/master/push_notification.go#L94
	notification.Set("aps", payload)

	customMap := m.Map()
	for key, value := range customMap {
		notification.Set(key, value)
	}

	resp := pusher.Client.Send(notification)
	if !resp.Success {
		log.Printf("Failed to send Push Notification: %v\nPayload:\n%#v\n\n", resp.Error, notification)
		return resp.Error
	}

	return nil
}
