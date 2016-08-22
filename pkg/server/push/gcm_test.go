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
	"errors"
	"testing"

	"github.com/google/go-gcm"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGCMSend(t *testing.T) {
	Convey("GCMPusher", t, func() {
		pusher := GCMPusher{
			APIKey: "apiKey",
		}
		device := skydb.Device{
			Token: "deviceToken",
		}

		var (
			apiKey     string
			gcmMessage gcm.HttpMessage
		)
		gcmSendHTTP = func(k string, m gcm.HttpMessage) (*gcm.HttpResponse, error) {
			apiKey = k
			gcmMessage = m
			return &gcm.HttpResponse{}, nil
		}
		defer func() {
			gcmSendHTTP = gcm.SendHttp
		}()

		Convey("sends notification", func() {
			err := pusher.Send(MapMapper{
				"gcm": map[string]interface{}{
					"content_available": true,
					"notification": map[string]interface{}{
						"title": "You have got a message",
						"body":  "This is a message.",
						"icon":  "myicon",
						"sound": "default",
						"badge": "5",
					},
					"data": map[string]interface{}{
						"string":  "value",
						"integer": 1,
						"nested": map[string]interface{}{
							"should": "correct",
						},
					},
				},
			}, device)

			So(err, ShouldBeNil)
			So(apiKey, ShouldEqual, "apiKey")
			So(gcmMessage, ShouldResemble, gcm.HttpMessage{
				To:               "deviceToken",
				ContentAvailable: true,
				Data: gcm.Data{
					"string":  "value",
					"integer": 1,
					"nested": map[string]interface{}{
						"should": "correct",
					},
				},
				Notification: gcm.Notification{
					Title: "You have got a message",
					Body:  "This is a message.",
					Icon:  "myicon",
					Sound: "default",
					Badge: "5",
				},
			})
		})

		Convey("propagates error from gcm.SendHttp", func() {
			gcmSendHTTP = func(string, gcm.HttpMessage) (*gcm.HttpResponse, error) {
				return nil, errors.New("gcm_test: some error")
			}

			err := pusher.Send(EmptyMapper, device)
			So(err, ShouldResemble, errors.New("gcm_test: some error"))
		})
	})

}
