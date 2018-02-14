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
	"github.com/google/go-gcm"
	"github.com/mitchellh/mapstructure"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

var gcmSendHTTP = gcm.SendHttp

// GCMPusher sends push notifications via GCM.
type GCMPusher struct {
	APIKey string
}

// Send sends the dictionary represented by m to device.
func (p *GCMPusher) Send(m Mapper, device skydb.Device) error {
	message := gcm.HttpMessage{}

	if err := mapGCMMessage(m, &message); err != nil {
		log.Errorf("Failed to convert gcm message: %v", err)
		return err
	}

	message.To = device.Token
	message.RegistrationIds = nil

	// NOTE(limouren): might need to check response for deleted / invalid
	// device here
	if _, err := gcmSendHTTP(p.APIKey, message); err != nil {
		log.Errorf("Failed to send GCM Notification: %v", err)
		return err
	}

	return nil
}

func mapGCMMessage(mapper Mapper, msg *gcm.HttpMessage) error {
	m := mapper.Map()
	if gcmMap, ok := m["gcm"].(map[string]interface{}); ok {
		config := mapstructure.DecoderConfig{
			TagName: "json",
			Result:  msg,
		}
		// NewDecoder only returns error when DecoderConfig.Result
		// is not a pointer.
		decoder, err := mapstructure.NewDecoder(&config)
		if err != nil {
			panic(err)
		}

		if err := decoder.Decode(gcmMap); err != nil {
			return err
		}
	}

	return nil
}
