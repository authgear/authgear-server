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
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/RobotsAndPencils/buford/push"
	log "github.com/Sirupsen/logrus"
	"github.com/skygeario/skygear-server/skydb"
)

// GatewayType determine which kind of gateway should be used for APNS
type GatewayType string

// Available gateways
const (
	Sandbox    GatewayType = "sandbox"
	Production             = "production"
)

// private interface s.t. we can mock push.Service in test
type pushService interface {
	Push(deviceToken string, headers *push.Headers, payload []byte) (string, error)
}

// APNSPusher pushes notification via apns
type APNSPusher struct {
	// Function to obtain a skydb connection
	connOpener func() (skydb.Conn, error)

	conn    skydb.Conn
	service pushService
	failed  chan failedNotification
	topic   string
}

type failedNotification struct {
	deviceToken string
	err         push.Error
}

// parseCertificateLeaf parse the provided TLS certificate for its
// leaf certificate. Returns an error if the leaf certificate cannot be found.
func parseCertificateLeaf(certificate *tls.Certificate) error {
	if certificate.Leaf != nil {
		return nil
	}

	for _, cert := range certificate.Certificate {
		x509Cert, err := x509.ParseCertificate(cert)
		if err != nil {
			return err
		}
		certificate.Leaf = x509Cert
		return nil
	}
	return errors.New("push/apns: provided APNS certificate does not contain leaf")
}

// findDefaultAPNSTopic returns the APNS topic in the TLS certificate.
//
// The Subject of leaf certificate should contains the UID, which we can
// use as the topic for APNS. The topic is usually the same as the
// application bundle // identifier.
//
// Returns the topic name, and an error if an error occuring finding the topic
// name.
func findDefaultAPNSTopic(certificate tls.Certificate) (string, error) {
	if certificate.Leaf == nil {
		err := parseCertificateLeaf(&certificate)
		if err != nil {
			return "", err
		}
	}

	// Loop over the subject names array to look for UID
	uidObjectIdentifier := asn1.ObjectIdentifier([]int{0, 9, 2342, 19200300, 100, 1, 1})
	for _, attr := range certificate.Leaf.Subject.Names {
		if uidObjectIdentifier.Equal(attr.Type) {
			switch value := attr.Value.(type) {
			case string:
				return value, nil
			}
			break
		}
	}

	return "", errors.New("push/apns: cannot find UID in APNS certificate subject name")
}

// NewAPNSPusher returns a new APNSPusher from content of certificate
// and private key as string
func NewAPNSPusher(connOpener func() (skydb.Conn, error), gwType GatewayType, cert string, key string) (*APNSPusher, error) {
	certificate, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return nil, err
	}

	topic, err := findDefaultAPNSTopic(certificate)
	if err != nil {
		return nil, err
	}

	client, err := push.NewClient(certificate)
	if err != nil {
		return nil, err
	}

	var service *push.Service
	switch gwType {
	case Sandbox:
		service = push.NewService(client, push.Development)
	case Production:
		service = push.NewService(client, push.Production)
	default:
		return nil, fmt.Errorf("push/apns: unrecognized gateway type %s", gwType)
	}

	return &APNSPusher{
		connOpener: connOpener,
		service:    service,
		topic:      topic,
	}, nil
}

// Run listens to the notification error channel
func (pusher *APNSPusher) Run() {
	pusher.failed = make(chan failedNotification)
	conn, err := pusher.connOpener()
	if err != nil {
		log.Errorf("apns/fb: failed to open skydb.Conn, abort feedback retrival: %v\n", err)
		return
	}

	pusher.conn = conn

	go func() {
		pusher.checkFailedNotifications()
	}()
}

func (pusher *APNSPusher) Stop() {
	close(pusher.failed)
}

func (pusher *APNSPusher) checkFailedNotifications() {
	for failedNote := range pusher.failed {
		pusher.handleFailedNotification(failedNote)
	}
}

func (pusher *APNSPusher) queueFailedNotification(deviceToken string, err push.Error) bool {
	logger := log.WithFields(log.Fields{
		"deviceToken": deviceToken,
	})
	failed := pusher.failed
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

func shouldUnregisterDevice(failedNote failedNotification) bool {
	return failedNote.err.Status == http.StatusGone || failedNote.err.Reason.Error() == "BadDeviceToken"
}

func (pusher *APNSPusher) handleFailedNotification(failedNote failedNotification) {
	if shouldUnregisterDevice(failedNote) {
		pusher.unregisterDevice(failedNote.deviceToken, failedNote.err.Timestamp)
	}
}

func (pusher *APNSPusher) unregisterDevice(deviceToken string, timestamp time.Time) {
	logger := log.WithFields(log.Fields{
		"deviceToken": deviceToken,
	})

	defer func() {
		if r := recover(); r != nil {
			logger.Panicf("Panic occurred while unregistering device: %s", r)
		}
	}()

	if err := pusher.conn.DeleteDevicesByToken(deviceToken, timestamp); err != nil && err != skydb.ErrDeviceNotFound {
		logger.Errorf("apns/fb: failed to delete device token = %s: %v", deviceToken, err)
		return
	}

	logger.Info("Unregistered device from skydb")
}

// Send sends a notification to the device identified by the
// specified device
func (pusher *APNSPusher) Send(m Mapper, device skydb.Device) error {
	logger := log.WithFields(log.Fields{
		"deviceToken": device.Token,
		"deviceID":    device.ID,
		"apnsTopic":   pusher.topic,
	})

	if m == nil {
		logger.Warn("Cannot send push notification with nil data.")
		return nil
	}

	apnsMap, ok := m.Map()["apns"].(map[string]interface{})
	if !ok {
		return errors.New("push/apns: payload has no apns dictionary")
	}

	serializedPayload, err := json.Marshal(apnsMap)
	if err != nil {
		return err
	}

	headers := push.Headers{
		Topic: pusher.topic,
	}

	// push the notification:
	apnsid, err := pusher.service.Push(device.Token, &headers, serializedPayload)
	if err != nil {
		if pushError, ok := err.(*push.Error); ok && pushError != nil {
			// We recognize the error, and that error comes from APNS
			logger.WithFields(log.Fields{
				"apnsErrorReason":    pushError.Reason,
				"apnsErrorStatus":    pushError.Status,
				"apnsErrorTimestamp": pushError.Timestamp,
			}).Error("Failed to send push notification to APNS")
			pusher.queueFailedNotification(device.Token, *pushError)
			return err
		}

		logger.Errorf("Failed to send push notification: %s", err)
		return err
	}

	logger.WithFields(log.Fields{
		"apnsID": apnsid,
	}).Info("Sent push notification to APNS")
	return nil
}
