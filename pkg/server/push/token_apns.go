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
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/SkygearIO/buford/push"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

const tokenRefreshInterval = 30 * time.Minute

type tokenBasedAPNSPusher struct {
	APNSPusher

	// Function to obtain a skydb connection
	connOpener func() (skydb.Conn, error)

	conn    skydb.Conn
	service pushService

	teamID     string
	keyID      string
	privateKey *ecdsa.PrivateKey

	tokenRefreshTimer *time.Ticker
	tokenMutex        *sync.RWMutex
	token             token

	failed chan failedNotification
}

type token struct {
	value     string
	expiredAt time.Time
}

// NewTokenBasedAPNSPusher creates a new APNSPusher from the content of auth key
func NewTokenBasedAPNSPusher(
	connOpener func() (skydb.Conn, error),
	gatewayType GatewayType,
	teamID string,
	keyID string,
	key string,
) (APNSPusher, error) {
	keyBlock, _ := pem.Decode([]byte(key))
	if keyBlock == nil {
		return nil, errors.New("APNS Auth Key is malformed")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	client, err := push.NewTLSClient(&tls.Config{})
	if err != nil {
		return nil, err
	}

	service, err := newPushService(client, gatewayType)
	if err != nil {
		return nil, err
	}

	switch typedKey := privateKey.(type) {
	case *ecdsa.PrivateKey:
		return &tokenBasedAPNSPusher{
			connOpener: connOpener,
			service:    service,
			teamID:     teamID,
			keyID:      keyID,
			privateKey: typedKey,
			tokenMutex: &sync.RWMutex{},
		}, nil
	default:
		return nil, errors.New("Unknown APNS Auth Key type")
	}
}

func (pusher *tokenBasedAPNSPusher) refreshToken() {
	claims := jwt.StandardClaims{
		Issuer:   pusher.teamID,
		IssuedAt: time.Now().Unix(),
	}
	method := jwt.SigningMethodES256
	jwtToken := &jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": method.Alg(),
			"kid": pusher.keyID,
		},
		Claims: claims,
		Method: method,
	}

	signedToken, err := jwtToken.SignedString(pusher.privateKey)
	if err != nil {
		log.Warnf("Failed to signed the apns auth token: %v\n", err)
		return
	}

	pusher.updateToken(token{
		value:     signedToken,
		expiredAt: time.Now().Add(1 * time.Hour),
	})
}

func (pusher *tokenBasedAPNSPusher) updateToken(newToken token) {
	pusher.tokenMutex.Lock()
	defer pusher.tokenMutex.Unlock()

	pusher.token = newToken
}

func (pusher tokenBasedAPNSPusher) getToken() token {
	pusher.tokenMutex.RLock()
	defer pusher.tokenMutex.RUnlock()

	return pusher.token
}

// Start setups the pusher and starts it
func (pusher *tokenBasedAPNSPusher) Start() {
	conn, err := pusher.connOpener()
	if err != nil {
		log.Errorf("push/apns: failed to open skydb.Conn, abort feedback retrival: %v\n", err)
		return
	}

	pusher.conn = conn
	pusher.failed = make(chan failedNotification)
	pusher.tokenRefreshTimer = time.NewTicker(tokenRefreshInterval)

	go func() {
		checkFailedNotifications(pusher)
	}()

	go func() {
		pusher.refreshToken()
		for tickerTime := range pusher.tokenRefreshTimer.C {
			log.
				WithField("time", tickerTime).
				Info("Refreshing APNS Token")
			pusher.refreshToken()
		}
	}()
}

// Stop stops and cleans up the pusher
func (pusher *tokenBasedAPNSPusher) Stop() {
	close(pusher.failed)
	pusher.failed = nil

	pusher.tokenRefreshTimer.Stop()
}

// Send sends a notification to the device identified by the
// specified device
func (pusher *tokenBasedAPNSPusher) Send(m Mapper, device skydb.Device) error {
	logger := log.WithFields(logrus.Fields{
		"deviceToken": device.Token,
		"deviceID":    device.ID,
		"deviceTopic": device.Topic,
	})

	if m == nil {
		logger.Warn("Cannot send push notification with nil data.")
		return errors.New("push/apns: push notification has no data")
	}

	if device.Topic == "" {
		logger.Print("Found device with null topic field, ignored.")
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
		Topic:         device.Topic,
		Authorization: pusher.getToken().value,
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

func (pusher tokenBasedAPNSPusher) getFailedNotificationChannel() chan failedNotification {
	return pusher.failed
}

func (pusher tokenBasedAPNSPusher) deleteDeviceToken(token string, beforeTime time.Time) error {
	return pusher.conn.DeleteDevicesByToken(token, beforeTime)
}
