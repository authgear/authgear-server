package push

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/timehop/apns"
)

// private interface s.t. we can mock apns.Client in test
type apnsSender interface {
	Send(n apns.Notification) error
	FailedNotifs() chan apns.NotificationResult
}

// APNSPusher pushes notification via apns
type APNSPusher struct {
	// we are directly coupling on apns as it seems redundant to duplicate
	// all the payload and client logic and interfaces.
	Client apnsSender
}

// NewAPNSPusher returns a new APNSPusher from content of certificate
// and private key as string
func NewAPNSPusher(gateway string, cert string, key string) (*APNSPusher, error) {
	client, err := apns.NewClient(gateway, cert, key)
	if err != nil {
		return nil, err
	}

	return &APNSPusher{Client: &wrappedClient{&client}}, nil
}

// NewAPNSPusherFromFiles returns a new APNSPusher from certificate and
// private key file
func NewAPNSPusherFromFiles(gateway string, certPath string, keyPath string) (*APNSPusher, error) {
	client, err := apns.NewClientWithFiles(gateway, certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &APNSPusher{Client: &wrappedClient{&client}}, nil
}

// Init set up the notification error channel
func (pusher *APNSPusher) Init() error {
	go func() {
		for result := range pusher.Client.FailedNotifs() {
			log.Errorf("Failed to send notification = %s: %v", result.Notif.ID, result.Err)
		}
	}()

	return nil
}

func setPayloadAPS(apsMap map[string]interface{}, aps *apns.APS) {
	for key, value := range apsMap {
		switch key {
		case "content-available":
			switch value := value.(type) {
			case int:
				aps.ContentAvailable = value
			case float64:
				aps.ContentAvailable = int(value)
			}
		case "sound":
			if sound, ok := value.(string); ok {
				aps.Sound = sound
			}
		case "badge":
			switch value := value.(type) {
			case int:
				aps.Badge = &value
			case float64:
				badge := int(value)
				aps.Badge = &badge
			}
		case "alert":
			if body, ok := value.(string); ok {
				aps.Alert.Body = body
			} else if alertMap, ok := value.(map[string]interface{}); ok {
				jsonbytes, err := json.Marshal(&alertMap)
				if err != nil {
					panic("Unable to convert alert to json.")
				}

				err = json.Unmarshal(jsonbytes, &aps.Alert)
				if err != nil {
					panic("Unable to convert json back to Alert struct.")
				}
			}
		}
	}
}

func setPayload(m Mapper, p *apns.Payload) {
	customMap := m.Map()
	for key, value := range customMap {
		if key == "aps" {
			if apsMap, ok := value.(map[string]interface{}); ok {
				setPayloadAPS(apsMap, &p.APS)
				log.Errorf("Failed to set key = %v", p.APS.ContentAvailable)
			} else {
				log.Errorf("Failed to set key = %v, value = %v", key, value)
			}
		} else if err := p.SetCustomValue(key, value); err != nil {
			log.Errorf("Failed to set key = %v, value = %v", key, value)
		}
	}
}

// Send sends a notification to the device identified by the
// specified deviceToken
func (pusher *APNSPusher) Send(m Mapper, deviceToken string) error {
	payload := apns.NewPayload()
	if m != nil {
		setPayload(m, payload)
	}

	notification := apns.NewNotification()
	notification.Payload = payload
	notification.DeviceToken = deviceToken
	notification.Priority = apns.PriorityImmediate

	if err := pusher.Client.Send(notification); err != nil {
		log.Errorf("Failed to send Push Notification: %v", err)
		return err
	}

	return nil
}

// wrapper of apns.Client which implement apnsSender
type wrappedClient struct {
	ci *apns.Client
}

func (c *wrappedClient) Send(n apns.Notification) error {
	return c.ci.Send(n)
}

func (c *wrappedClient) FailedNotifs() chan apns.NotificationResult {
	return c.ci.FailedNotifs
}
