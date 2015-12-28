package push

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/skydb"
	"github.com/timehop/apns"
)

// GatewayType determine which kind of gateway should be used for APNS
type GatewayType string

// Available gateways
const (
	Sandbox    GatewayType = "sandbox"
	Production             = "production"
)

// private interface s.t. we can mock apns.Client in test
type apnsSender interface {
	Send(n apns.Notification) error
	FailedNotifs() chan apns.NotificationResult
}

// private interface to mock apns.Feedback in test
type feedbackReceiver interface {
	Receive() <-chan apns.FeedbackTuple
}

// APNSPusher pushes notification via apns
type APNSPusher struct {
	// Function to obtain a skydb connection
	connOpener func() (skydb.Conn, error)

	// we are directly coupling on apns as it seems redundant to duplicate
	// all the payload and client logic and interfaces.
	client apnsSender

	feedback feedbackReceiver
}

// NewAPNSPusher returns a new APNSPusher from content of certificate
// and private key as string
func NewAPNSPusher(connOpener func() (skydb.Conn, error), gwType GatewayType, cert string, key string) (*APNSPusher, error) {
	var gateway, fbGateway string
	switch gwType {
	case Sandbox:
		gateway = apns.SandboxGateway
		fbGateway = apns.SandboxFeedbackGateway
	case Production:
		gateway = apns.ProductionGateway
		fbGateway = apns.ProductionFeedbackGateway
	default:
		return nil, fmt.Errorf("unrecgonized GatewayType = %#v", gwType)
	}

	client, err := apns.NewClient(gateway, cert, key)
	if err != nil {
		return nil, err
	}

	fb, err := apns.NewFeedback(fbGateway, cert, key)
	if err != nil {
		return nil, err
	}

	return &APNSPusher{
		connOpener: connOpener,
		client:     &wrappedClient{&client},
		feedback:   fb,
	}, nil
}

// Run listens to the notification error channel
func (pusher *APNSPusher) Run() {
	for result := range pusher.client.FailedNotifs() {
		log.Errorf("Failed to send notification = %s: %v", result.Notif.ID, result.Err)
	}
}

// RunFeedback kicks start receiving from the Feedback Service.
//
// The checking behaviour is to:
//	1. Receive once on startup
//	2. Receive once at 00:00:00 everyday
func (pusher *APNSPusher) RunFeedback() {
	pusher.recvFeedback()

	for {
		now := time.Now()
		year, month, day := now.Date()
		nextDay := time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
		d := nextDay.Sub(now)

		log.Infof("apns/fb: next feedback scheduled after %v, at %v", d, nextDay)

		<-time.After(d)

		log.Infoln("apns/fb: going to query feedback service")
		pusher.recvFeedback()
	}
}

func (pusher *APNSPusher) recvFeedback() {
	conn, err := pusher.connOpener()
	if err != nil {
		log.Errorf("apns/fb: failed to open skydb.Conn, abort feedback retrival: %v\n", err)
		return
	}

	received := false
	for fb := range pusher.feedback.Receive() {
		log.Infof("apns/fb: got a feedback = %v", fb)

		received = true

		// NOTE(limouren): it might be more elegant in the future to extend
		// push.Sender as NotificationService and bridge over the differences
		// between gcm and apns on handling unregistered devices (probably
		// as an async channel)
		if err := conn.DeleteDeviceByToken(fb.DeviceToken, fb.Timestamp); err != nil {
			log.Errorf("apns/fb: failed to delete device token = %s: %v", fb.DeviceToken, err)
		}
	}

	if !received {
		log.Infoln("apns/fb: no feedback received")
	}
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
				aps.Badge.Set(uint(value))
			case float64:
				aps.Badge.Set(uint(value))
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

func setPayload(m map[string]interface{}, p *apns.Payload) {
	if apsValue, ok := m["aps"]; ok {
		if apsMap, ok := apsValue.(map[string]interface{}); ok {
			setPayloadAPS(apsMap, &p.APS)
		} else {
			log.Errorf("Want aps.(type) be map[string]interface{}, got %T", apsValue)
		}
	}

	// set custom values
	for key, value := range m {
		// the "aps" key is not a custom key
		if key == "aps" {
			continue
		}
		if err := p.SetCustomValue(key, value); err != nil {
			log.Errorf("Failed to set data[%v] = %v", key, value)
		}
	}
}

// Send sends a notification to the device identified by the
// specified device
func (pusher *APNSPusher) Send(m Mapper, device *skydb.Device) error {
	if m == nil {
		return nil
	}
	apnsMap, ok := m.Map()["apns"].(map[string]interface{})
	if !ok {
		return errors.New("push/apns: payload has no apns dictionary")
	}

	payload := apns.NewPayload()
	setPayload(apnsMap, payload)

	notification := apns.NewNotification()
	notification.Payload = payload
	notification.DeviceToken = device.Token
	notification.Priority = apns.PriorityImmediate

	if err := pusher.client.Send(notification); err != nil {
		log.Errorf("Failed to send APNS Notification: %v", err)
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
