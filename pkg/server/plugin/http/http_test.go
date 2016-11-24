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

package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"golang.org/x/net/context"

	skyplugin "github.com/skygeario/skygear-server/pkg/server/plugin"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func createTestServerClient(handlerFunc http.HandlerFunc) (*httptest.Server, http.Client) {
	server := httptest.NewServer(http.HandlerFunc(handlerFunc))
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	client := http.Client{Transport: transport}
	return server, client
}

func TestRun(t *testing.T) {
	Convey("test call plugin", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		appconfig := skyconfig.Configuration{}
		appconfig.App.Name = "hello-world"
		transport := &httpTransport{
			Path:   "http://localhost:8000",
			Args:   []string{},
			state:  skyplugin.TransportStateReady,
			config: appconfig,
		}

		Convey("send event", func() {
			Convey("success case", func() {
				data := []byte(`{"data": "hello-world"}`)
				httpmock.RegisterResponder(
					"POST",
					"http://localhost:8000",
					func(req *http.Request) (*http.Response, error) {
						bodyBytes, _ := ioutil.ReadAll(req.Body)
						So(
							bodyBytes,
							ShouldEqualJSON,
							`{"kind":"event","name":"foo","param":{"data":"hello-world"}}`,
						)

						return httpmock.NewJsonResponse(200, map[string]interface{}{
							"result": map[string]interface{}{"data": "hello-world-resp"},
						})
					},
				)

				out, err := transport.SendEvent("foo", data)
				So(out, ShouldEqualJSON, `{"data": "hello-world-resp"}`)
				So(err, ShouldBeNil)
			})

			Convey("fail case", func() {
				data := []byte(`{"data": "hello-world"}`)
				httpmock.RegisterResponder(
					"POST",
					"http://localhost:8000",
					func(req *http.Request) (*http.Response, error) {
						bodyBytes, _ := ioutil.ReadAll(req.Body)
						So(
							bodyBytes,
							ShouldEqualJSON,
							`{"kind":"event","name":"foo2","param":{"data":"hello-world"}}`,
						)

						return httpmock.NewJsonResponse(500, map[string]interface{}{
							"error": map[string]interface{}{
								"code":    skyerr.UnexpectedError,
								"message": "test error",
							},
						})
					},
				)

				out, err := transport.SendEvent("foo2", data)
				So(out, ShouldBeNil)
				So(err.Error(), ShouldEqual, "UnexpectedError: test error")
			})
		})

		Convey("run lambda", func() {
			ctx := context.WithValue(context.Background(), router.UserIDContextKey, "user")
			ctx = context.WithValue(ctx, router.AccessKeyTypeContextKey, router.MasterAccessKey)
			data := `{"data": "bye"}`
			httpmock.RegisterResponder("POST", "http://localhost:8000",
				func(req *http.Request) (*http.Response, error) {
					out, _ := ioutil.ReadAll(req.Body)
					So(out, ShouldEqualJSON, `{"context":{"user_id":"user","access_key_type":"master"},"kind":"op","name":"john","param":{"data":"bye"}}`)
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"result": map[string]interface{}{"data": "hello"},
					})
				},
			)

			out, err := transport.RunLambda(ctx, "john", []byte(data))
			So(out, ShouldEqualJSON, `{"data": "hello"}`)
			So(err, ShouldBeNil)
		})

		Convey("run hook", func() {
			ctx := context.WithValue(context.Background(), router.UserIDContextKey, "user")
			ctx = context.WithValue(ctx, router.AccessKeyTypeContextKey, router.ClientAccessKey)

			recordin := skydb.Record{
				ID:      skydb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: skydb.RecordACL{
					skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
					skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "some note content",
					"noteOrder": float64(1),
					"tags":      []interface{}{"test", "unimportant"},
					"date":      time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC),
					"ref":       skydb.NewReference("category", "1"),
					"asset":     &skydb.Asset{Name: "asset-name"},
				},
			}

			recordold := skydb.Record{
				ID:      skydb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: skydb.RecordACL{
					skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
					skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "original content",
					"noteOrder": float64(1),
					"tags":      []interface{}{},
					"date":      time.Date(2017, 7, 21, 19, 30, 24, 0, time.UTC),
				},
			}

			httpmock.RegisterResponder("POST", "http://localhost:8000",
				func(req *http.Request) (*http.Response, error) {
					in, _ := ioutil.ReadAll(req.Body)
					So(in, ShouldEqualJSON, `{
						"context":{"user_id":"user","access_key_type":"client"},
						"kind":"hook",
						"name":"beforeSave",
						"param":{
							"record": {
								"_id": "note/id",
								"_type": "record",
								"_ownerID": "john.doe@example.com",
								"content": "some note content",
								"noteOrder": 1,
								"tags": ["test", "unimportant"],
								"date": {
									"$type": "date",
									"$date": "2017-07-23T19:30:24Z"
								},
								"ref": {
									"$type": "ref",
									"$id": "category/1"
								},
								"asset":{
									"$type": "asset",
									"$name": "asset-name"
								},
								"_access": [{
									"relation": "friend",
									"level": "write"
								}, {
									"relation": "$direct",
									"level": "read",
									"user_id": "user_id"
								}]
							},
							"original": {
								"_id": "note/id",
								"_type": "record",
								"_ownerID": "john.doe@example.com",
								"content": "original content",
								"noteOrder": 1,
								"tags": [],
								"date": {
									"$type": "date",
									"$date": "2017-07-21T19:30:24Z"
								},
								"_access": [{
									"relation": "friend",
									"level": "write"
								}, {
									"relation": "$direct",
									"level": "read",
									"user_id": "user_id"
								}]
							}
						}
					}`)

					return httpmock.NewStringResponse(200, `{
						"result": {
							"_id": "note/id",
							"_type": "record",
							"_ownerID": "john.doe@example.com",
							"content": "content has been modified",
							"noteOrder": 1,
							"tags": ["test", "unimportant"],
							"date": {
								"$type": "date",
								"$date": "2017-07-23T19:30:24Z"
							},
							"ref": {
								"$type": "ref",
								"$id": "category/1"
							},
							"asset":{
								"$type": "asset",
								"$name": "asset-name"
							},
							"_access": [{
								"relation": "friend",
								"level": "write"
							}, {
								"relation": "$direct",
								"level": "read",
								"user_id": "user_id"
							}]
						}
					}`), nil
				},
			)

			recordout, err := transport.RunHook(ctx, "beforeSave", &recordin, &recordold)
			So(err, ShouldBeNil)

			datein := recordin.Data["date"].(time.Time)
			delete(recordin.Data, "date")
			So(recordin, ShouldResemble, skydb.Record{
				ID:      skydb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: skydb.RecordACL{
					skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
					skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "some note content",
					"noteOrder": float64(1),
					"tags":      []interface{}{"test", "unimportant"},
					"ref":       skydb.NewReference("category", "1"),
					"asset":     &skydb.Asset{Name: "asset-name"},
				},
			})
			// GoConvey's bug, ShouldEqual and ShouldResemble doesn't work on time.Time
			So(datein == time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC), ShouldBeTrue)

			dateout := recordout.Data["date"].(time.Time)
			delete(recordout.Data, "date")
			So(*recordout, ShouldResemble, skydb.Record{
				ID:      skydb.NewRecordID("note", "id"),
				OwnerID: "john.doe@example.com",
				ACL: skydb.RecordACL{
					skydb.NewRecordACLEntryRelation("friend", skydb.WriteLevel),
					skydb.NewRecordACLEntryDirect("user_id", skydb.ReadLevel),
				},
				Data: map[string]interface{}{
					"content":   "content has been modified",
					"noteOrder": float64(1),
					"tags":      []interface{}{"test", "unimportant"},
					"ref":       skydb.NewReference("category", "1"),
					"asset":     &skydb.Asset{Name: "asset-name"},
				},
			})
			So(dateout == time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC), ShouldBeTrue)
		})

		Convey("run timer", func() {
			httpmock.RegisterResponder("POST", "http://localhost:8000",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"result": map[string]interface{}{"data": "hello"},
					})
				},
			)

			out, err := transport.RunTimer("john", []byte{})
			So(out, ShouldEqualJSON, `{"data": "hello"}`)
			So(err, ShouldBeNil)
		})

		Convey("run provider", func() {
			authData := map[string]interface{}{"data": "bye"}

			httpmock.RegisterResponder("POST", "http://localhost:8000",
				func(req *http.Request) (*http.Response, error) {
					out, _ := ioutil.ReadAll(req.Body)
					So(out, ShouldEqualJSON, `{"kind":"provider","name":"com.example","param":{"action":"login","auth_data":{"data":"bye"}}}`)

					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"result": map[string]interface{}{
							"principal_id": "hello",
							"auth_data":    map[string]interface{}{"hello": "world"},
						},
					})
				},
			)

			authReq := &skyplugin.AuthRequest{"com.example", "login", authData}
			authResp, err := transport.RunProvider(authReq)
			So(authResp, ShouldResemble, &skyplugin.AuthResponse{
				PrincipalID: "hello",
				AuthData:    map[string]interface{}{"hello": "world"},
			})
			So(err, ShouldBeNil)
		})
	})
}
