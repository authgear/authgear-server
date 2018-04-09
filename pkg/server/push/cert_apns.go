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
	"net/http"
	"time"

	"github.com/SkygearIO/buford/push"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type certBasedAPNSPusher struct {
	APNSPusher

	// Function to obtain a skydb connection
	connOpener func() (skydb.Conn, error)

	topic   string
	conn    skydb.Conn
	service pushService

	failed chan failedNotification
}

// parseCertificateLeaf parses the provided TLS certificate for its
// leaf certificate. Returns an error if the leaf certificate cannot be found.
func parseCertificateLeaf(certificate *tls.Certificate) error {
	if certificate.Leaf != nil {
		return nil
	}

	if len(certificate.Certificate) > 0 {
		x509Cert, err := x509.ParseCertificate(certificate.Certificate[0])
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
// Returns the topic name, and an error if an error occurring finding the topic
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

// NewCertBasedAPNSPusher returns a new APNSPusher from content of certificate
// and private key as string
func NewCertBasedAPNSPusher(
	connOpener func() (skydb.Conn, error),
	gatewayType GatewayType,
	cert string,
	key string,
) (APNSPusher, error) {
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

	service, err := newPushService(client, gatewayType)
	if err != nil {
		return nil, err
	}

	return &certBasedAPNSPusher{
		connOpener: connOpener,
		service:    service,
		topic:      topic,
	}, nil
}

// Start setups the pusher and starts it
func (pusher *certBasedAPNSPusher) Start() {
	conn, err := pusher.connOpener()
	if err != nil {
		log.Errorf("push/apns: failed to open skydb.Conn, abort feedback retrival: %v\n", err)
		return
	}

	pusher.conn = conn
	pusher.failed = make(chan failedNotification)

	go func() {
		checkFailedNotifications(pusher)
	}()
}

// Stop stops and cleans up the pusher
func (pusher *certBasedAPNSPusher) Stop() {
	close(pusher.failed)
	pusher.failed = nil
}

// Send sends a notification to the device identified by the
// specified device
func (pusher *certBasedAPNSPusher) Send(m Mapper, device skydb.Device) error {
	logger := log.WithFields(logrus.Fields{
		"deviceToken": device.Token,
		"deviceID":    device.ID,
		"deviceTopic": device.Topic,
	})

	if m == nil {
		logger.Warn("Cannot send push notification with nil data.")
		return errors.New("push/apns: push notification has no data")
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

	apnsid, err := pusher.service.Push(device.Token, &headers, serializedPayload)
	if err != nil {
		if pushError, ok := err.(*push.Error); ok && pushError != nil {
			// We recognize the error, and that error comes from APNS
			pushLogger := logger.WithFields(logrus.Fields{
				"apnsErrorReason":    pushError.Reason,
				"apnsErrorStatus":    pushError.Status,
				"apnsErrorTimestamp": pushError.Timestamp,
			})
			if pushError.Status == http.StatusGone ||
				pushError.Reason.Error() == "BadDeviceToken" {
				pushLogger.Info("push/apns: device token is no longer valid")
				queueFailedNotification(pusher, device.Token, *pushError)
			} else {
				pushLogger.Error("push/apns: failed to send push notification")
			}
			return err
		}

		logger.Errorf("Failed to send push notification: %s", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"apnsID": apnsid,
	}).Info("push/apns: push notification is sent")

	return nil
}

func (pusher certBasedAPNSPusher) getFailedNotificationChannel() chan failedNotification {
	return pusher.failed
}

func (pusher certBasedAPNSPusher) deleteDeviceToken(token string, beforeTime time.Time) error {
	return pusher.conn.DeleteDevicesByToken(token, beforeTime)
}
