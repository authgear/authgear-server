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

	skyplugin "github.com/oursky/skygear/plugin"
	"github.com/oursky/skygear/plugin/common"
	"github.com/oursky/skygear/skydb"
	. "github.com/oursky/skygear/skytest"
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

		transport := &httpTransport{
			Path:  "http://localhost:8000",
			Args:  []string{},
			state: skyplugin.TransportStateReady,
		}

		Convey("run init", func() {
			httpmock.RegisterResponder("POST", "http://localhost:8000/init",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"data": "hello",
					})
				},
			)

			out, err := transport.RunInit()
			So(out, ShouldEqualJSON, `{"data": "hello"}`)
			So(err, ShouldBeNil)
		})

		Convey("run lambda", func() {
			ctx := context.WithValue(context.Background(), "UserID", "user")
			data := []byte(`{"data": "bye"}`)
			httpmock.RegisterResponder("POST", "http://localhost:8000/op/john",
				func(req *http.Request) (*http.Response, error) {
					encodedCtx := req.Header.Get("X-Skygear-Plugin-Context")
					decodedCtx := map[string]interface{}{}
					common.DecodeBase64JSON(encodedCtx, &decodedCtx)
					So(decodedCtx, ShouldResemble, map[string]interface{}{
						"user_id": "user",
					})

					out, _ := ioutil.ReadAll(req.Body)
					So(out, ShouldResemble, data)
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"result": map[string]interface{}{"data": "hello"},
					})
				},
			)

			out, err := transport.RunLambda(ctx, "john", data)
			So(out, ShouldEqualJSON, `{"data": "hello"}`)
			So(err, ShouldBeNil)
		})

		Convey("run hook", func() {
			ctx := context.WithValue(context.Background(), "UserID", "user")

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

			httpmock.RegisterResponder("POST", "http://localhost:8000/hook/beforeSave",
				func(req *http.Request) (*http.Response, error) {
					encodedCtx := req.Header.Get("X-Skygear-Plugin-Context")
					decodedCtx := map[string]interface{}{}
					common.DecodeBase64JSON(encodedCtx, &decodedCtx)
					So(decodedCtx, ShouldResemble, map[string]interface{}{
						"user_id": "user",
					})

					in, _ := ioutil.ReadAll(req.Body)
					So(in, ShouldEqualJSON, `{
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
			httpmock.RegisterResponder("POST", "http://localhost:8000/timer/john",
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

			httpmock.RegisterResponder("POST", "http://localhost:8000/provider/com.example/login",
				func(req *http.Request) (*http.Response, error) {
					out, _ := ioutil.ReadAll(req.Body)
					So(out, ShouldEqualJSON, `{"auth_data":{"data": "bye"}}`)

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
