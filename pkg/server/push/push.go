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

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

var log = logging.LoggerEntry("push")

// EmptyMapper is a Mapper which always returns a empty map.
const EmptyMapper = emptyMapper(0)

type emptyMapper int

func (m emptyMapper) Map() map[string]interface{} {
	return map[string]interface{}{}
}

// Mapper defines a single method Map()
type Mapper interface {
	// Implementor of Map should return a string-interface map which
	// all values are JSON-marshallable
	Map() map[string]interface{}
}

// MapMapper is a string-interface map that implemented the Mapper
// interface.
type MapMapper map[string]interface{}

// Map returns the map itself.
func (m MapMapper) Map() map[string]interface{} {
	return map[string]interface{}(m)
}

// Sender defines the methods that a push service should support.
type Sender interface {
	Send(m Mapper, device skydb.Device) error
}

// RouteSender routes notifications to registered senders that is capable of
// sending them. RouteSender itself doesn't send notifications.
type RouteSender struct {
	senders map[string]Sender
}

// NewRouteSender return a new RouteSender.
func NewRouteSender() RouteSender {
	return RouteSender{
		senders: map[string]Sender{},
	}
}

// Route registers a sender to handle notifications sent via a certain
// Push Notification Service.
func (s RouteSender) Route(service string, sender Sender) {
	s.senders[service] = sender
}

// Len returns the number of services registered with sender.
func (s RouteSender) Len() int {
	return len(s.senders)
}

// Send inspects device and route notification (m) to corresponding sender.
func (s RouteSender) Send(m Mapper, device skydb.Device) error {
	sender, ok := s.senders[device.Type]
	if !ok {
		log.WithFields(logrus.Fields{
			"device":  device,
			"message": m,
		}).Errorln("No sender can send device of the Type")

		return fmt.Errorf("cannot find sender with type = %s", device.Type)
	}

	return sender.Send(m, device)
}
