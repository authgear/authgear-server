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
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/RobotsAndPencils/buford/push"
	"github.com/skygeario/skygear-server/skydb"
	. "github.com/skygeario/skygear-server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

type naiveServiceNotification struct {
	DeviceToken string
	Headers     *push.Headers
	Payload     []byte
}

type naiveService struct {
	Sent []naiveServiceNotification
	Err  *push.Error
}

func (s *naiveService) Push(deviceToken string, headers *push.Headers, payload []byte) (string, error) {
	s.Sent = append(s.Sent, naiveServiceNotification{deviceToken, headers, payload})
	if s.Err != nil {
		return "", s.Err
	}
	return "77BAF428-6FD8-42DB-8D1E-6F14A36C0863", nil
}

func TestAPNSSend(t *testing.T) {
	Convey("APNSPusher", t, func() {
		service := naiveService{}
		pusher := APNSPusher{
			service: &service,
			failed:  make(chan failedNotification, 10),
		}
		device := skydb.Device{
			Token: "deviceToken",
		}
		secondDevice := skydb.Device{
			Token: "deviceToken2",
		}

		Convey("pushes notification", func() {
			customMap := MapMapper{

				"apns": map[string]interface{}{
					"aps": map[string]interface{}{
						"content-available": 1,
						"sound":             "sosumi.mp3",
						"badge":             5,
						"alert":             "This is a message.",
					},
					"string":  "value",
					"integer": 1,
					"nested": map[string]interface{}{
						"should": "correct",
					},
				},
			}

			So(pusher.Send(customMap, device), ShouldBeNil)
			So(pusher.Send(customMap, secondDevice), ShouldBeNil)
			So(len(service.Sent), ShouldEqual, 2)
			So(service.Sent[0].DeviceToken, ShouldEqual, "deviceToken")
			So(service.Sent[1].DeviceToken, ShouldEqual, "deviceToken2")

			for i := range service.Sent {
				n := service.Sent[i]
				So(string(n.Payload), ShouldEqualJSON, `{
					"aps": {
						"content-available": 1,
						"sound": "sosumi.mp3",
						"badge": 5,
						"alert": "This is a message."
					},
					"string": "value",
					"integer": 1,
					"nested": {
						"should": "correct"
					}
				}`)
			}
		})

		Convey("returns error when missing apns dictionary", func() {
			err := pusher.Send(EmptyMapper, device)
			So(err, ShouldResemble, errors.New("push/apns: payload has no apns dictionary"))
		})

		Convey("returns error returned from Service.Push (BadMessageId)", func() {
			service.Err = &push.Error{
				Reason:    errors.New("BadMessageId"),
				Status:    http.StatusBadRequest,
				Timestamp: time.Time{},
			}
			err := pusher.Send(MapMapper{
				"apns": map[string]interface{}{},
			}, device)
			So(err, ShouldResemble, service.Err)
		})

		Convey("returns error returned from Service.Push (Unregistered)", func() {
			pushError := push.Error{
				Reason:    errors.New("Unregistered"),
				Status:    http.StatusGone,
				Timestamp: time.Now(),
			}
			service.Err = &pushError
			err := pusher.Send(MapMapper{
				"apns": map[string]interface{}{},
			}, device)
			So(err, ShouldResemble, &pushError)
			So(<-pusher.failed, ShouldResemble, failedNotification{
				deviceToken: device.Token,
				err:         pushError,
			})
		})

		Convey("pushes with custom alert", func() {
			customMap := MapMapper{
				"apns": map[string]interface{}{
					"aps": map[string]interface{}{
						"alert": map[string]interface{}{
							"body":           "Acme message received from Johnny Appleseed",
							"action-loc-key": "VIEW",
						},
					},
				},
			}

			err := pusher.Send(customMap, device)

			So(err, ShouldBeNil)

			n := service.Sent[0]
			So(n.DeviceToken, ShouldEqual, "deviceToken")

			So(string(n.Payload), ShouldEqualJSON, `{
				"aps": {
					"alert": {
						"body": "Acme message received from Johnny Appleseed",
						"action-loc-key": "VIEW"
					}
				}
			}`)
		})
	})
}

type deleteCall struct {
	token string
	t     time.Time
}

type mockConn struct {
	calls []deleteCall
	err   error
	skydb.Conn
}

func (c *mockConn) DeleteDeviceByToken(token string, t time.Time) error {
	c.calls = append(c.calls, deleteCall{token, t})
	return c.err
}

func (c *mockConn) Open() (skydb.Conn, error) {
	return c, nil
}

func TestAPNSFeedback(t *testing.T) {
	Convey("APNSPusher", t, func() {
		conn := &mockConn{}
		pusher := APNSPusher{
			connOpener: conn.Open,
			conn:       conn,
		}

		Convey("unregister device", func() {
			pusher.unregisterDevice("devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})

		Convey("unregister device with error", func() {
			conn.err = errors.New("apns/test: unknown error")
			pusher.unregisterDevice("devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})

		Convey("handle unregistered notification", func() {
			pusher.handleFailedNotification(failedNotification{"devicetoken0", push.Error{
				Reason:    errors.New("Unregistered"),
				Status:    http.StatusGone,
				Timestamp: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			}})
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})

		Convey("check failed notifications", func() {
			pusher.failed = make(chan failedNotification)
			go func() {
				pushError1 := push.Error{
					Reason:    errors.New("Unregistered"),
					Status:    http.StatusGone,
					Timestamp: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				}
				pusher.failed <- failedNotification{"devicetoken0", pushError1}

				pushError2 := push.Error{
					Reason:    errors.New("Unregistered"),
					Status:    http.StatusGone,
					Timestamp: time.Date(2046, 1, 2, 15, 4, 5, 0, time.UTC),
				}
				pusher.failed <- failedNotification{"devicetoken1", pushError2}
				close(pusher.failed)
			}()

			pusher.checkFailedNotifications()
			So(conn.calls, ShouldResemble, []deleteCall{
				{"devicetoken0", time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)},
				{"devicetoken1", time.Date(2046, 1, 2, 15, 4, 5, 0, time.UTC)},
			})
		})
	})
}

func TestAPNSCertificate(t *testing.T) {
	APNSCert := []byte(`-----BEGIN CERTIFICATE-----
MIICvjCCAaYCCQDrgz2ANVkBfTANBgkqhkiG9w0BAQsFADAhMR8wHQYKCZImiZPy
LGQBAQwPY29tLmV4YW1wbGUuQXBwMB4XDTE2MDcwODA3NTAzN1oXDTE3MDcwODA3
NTAzN1owITEfMB0GCgmSJomT8ixkAQEMD2NvbS5leGFtcGxlLkFwcDCCASIwDQYJ
KoZIhvcNAQEBBQADggEPADCCAQoCggEBAKd+cBNZ9lyKb81WGbNU4Vh4Z5TJB0tI
V7m3Iohl0sd8zbUhL8ISXPpKnPvwo2DZScs6Y4hJsZxQenNm5ll4cFAgcNHbu0I6
T1VzLSnxtpgvDOdxOBN5Nw1syKfzMUJw8o8RMtRt9cYVBwlKvOI92agFqZVYCIA3
4T/531f/VejIFd8wzp8fLMS+A8dJ+Run9Z4r4KZu8VhtKUP8GAFZ0pt9PL4Rm4Rl
/J/FZi5EmCE9Ms1RZoLuwO/IKPuGIY5rRi1c0kbYL3+QPxlkJa9DGW/61mDEkPkx
l6MvrbBFQIDSRUVQ97a8RNk/5tBwsAnyyqYxx9i9wjudNCv5YQs/QL8CAwEAATAN
BgkqhkiG9w0BAQsFAAOCAQEAb1mm4+6B+8YFqagAQ18I8EzOIHqrceDHj+v3PAh7
jD+orKmbFnq6kbzEj9AHOp+A5EjSLIBIarXFJIsbRXenYwDLF+0dwFkXzzhLSsAO
kvsTQPFaQC/h3mV8stx2SLxTDpWMPaaNCOlPkTmEtqXA3fes/1hF6TYalYO6kHe8
47iuzxKNjgfjjYeK3o4ccFS0+29WVoU5t+wuZ0Ha27PPNOFHLvn9TI9A5L+8ujgr
oyxSZaLz1oPX7aCcC847s+a73+K4V4QwdvKhxEN2McdZqv1h1Ha5zptt4kniBv6X
OFCnXiurw3uY37eBckl/JR++IkUekyIq1EJ0vfWyW/mhPQ==
-----END CERTIFICATE-----`)
	APNSKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAp35wE1n2XIpvzVYZs1ThWHhnlMkHS0hXubciiGXSx3zNtSEv
whJc+kqc+/CjYNlJyzpjiEmxnFB6c2bmWXhwUCBw0du7QjpPVXMtKfG2mC8M53E4
E3k3DWzIp/MxQnDyjxEy1G31xhUHCUq84j3ZqAWplVgIgDfhP/nfV/9V6MgV3zDO
nx8sxL4Dx0n5G6f1nivgpm7xWG0pQ/wYAVnSm308vhGbhGX8n8VmLkSYIT0yzVFm
gu7A78go+4YhjmtGLVzSRtgvf5A/GWQlr0MZb/rWYMSQ+TGXoy+tsEVAgNJFRVD3
trxE2T/m0HCwCfLKpjHH2L3CO500K/lhCz9AvwIDAQABAoIBAQCgZYyee4BZjpkS
YmmqOpaySlunN/wsM9MOnjoLtLbtIq87zdQWXc98QQeknQVYMb1hSUEXurrDnq4k
5V2iQJwNn4Nq9KmW+pAOnIWbrUXW5vfMi7fPrjzyNkLR0ypRHiiqqSWsGMFMN8bN
Ny0621Acf4+u3OcHInwq7/baJkL271g1m0hX/7TJ/nv+SlO00IkB6tm8iU0LWant
4fSvhV2ULxa0fF5XXx74jqDFF+NzJ45XMUDbbe72RKpydMsdprqp4BgaV7fbtHIx
xjG/9z6KqM3v0bMkJFV9BnmWzNe7vrI96dizVR3w3Yygul7sNC2iZfe17ulaXXKA
n3X+PBTBAoGBANkwAMW+dQw9x/dDIQgfkg27I7qgANdYyri/lJ6+GvMdcCd4mzQ+
UnJzhPNJyPZSUZeHOthShbHvBcoYNi4AwEcor2bXJI3+jW3/Wqbpq87+9x3Hkk69
tkKvESKy8IABOTntFD3/VGFjYiLXVpg0huvkRSlSv70gMusmTMxisvWNAoGBAMVt
Bx9IVkZ4uFl565WmigUuO8ICgHeRovCqeF3g0YBlS/x9Jznd9QkMYT0hy9cE4b7Z
GRm89mEOUclWgdsibkyEdD1qOF6xF+fi4xKaes8Vc3vuuhg+UqpYvbIxKFAT81A2
adyL6npDJDqQ0DcRjMng8V/77ktLOe/g9HFLX957AoGAOIlKai9eAMXEXBVZb+fn
+TMR5e7oySYP/2+/nGMYWNj87Ql0PXFLvQddQIegjJ55JtzI8K7qppr2AtmyoN8J
Lnzky/yNQ3lUD6I9Ut3ZH5U3dsUQzPaNj2ZLK6ExAeFPqEiS0GC68m8QiMlNfWmP
BbDyYANubikHmDbsHvhCZbECgYA0hx62/wsdcu8xt1OsHIRqfnOd2gaOSax9tg2S
hMeZDtqZ0j7GkbypbKbOmhhfHEhn++FGzNUM2798/0xLnqyUJUW8NW/MGfhPVTmv
cHSudnmkhs7ytlpOQpAuQhAExlodhGzEJmH7p7OS9YbAsCWybOwr6p7rX5eJsGO5
ZSGb0wKBgQCvYtVVUGPBbmp94yoRgQjPj+iZtSjmVA9LaNoX+g4UWxKTU3y+r9kg
SZVlxzxtvFvO+LcxmP6wBSX30HmhtIFLxmOyySl6BjfbtE0uFnIxxrKhb2L0gqoN
g723fJntDb71I1IS31Vd2wqqpVB4kDp8OiPnPp8ats/cNUFk77Jhxw==
-----END RSA PRIVATE KEY-----`)

	Convey("find apns topic", t, func() {
		certificate, err := tls.X509KeyPair(APNSCert, APNSKey)
		So(err, ShouldBeNil)
		topic, err := findDefaultAPNSTopic(certificate)
		So(err, ShouldBeNil)
		So(topic, ShouldEqual, "com.example.App")
	})
}
