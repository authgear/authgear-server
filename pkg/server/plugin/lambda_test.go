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

package plugin

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/asset/mock_asset"
	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skydbtest"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
)

func TestLambdaCreation(t *testing.T) {
	Convey("create simple lambda", t, func() {
		handler := NewLambdaHandler(map[string]interface{}{
			"name": "hello:world",
		}, nil)

		So(handler.Name, ShouldEqual, "hello:world")
		So(handler.AccessKeyRequired, ShouldBeFalse)
		So(handler.UserRequired, ShouldBeFalse)
	})

	Convey("create user required lambda", t, func() {
		handler := NewLambdaHandler(map[string]interface{}{
			"name":          "hello:world",
			"user_required": true,
		}, nil)

		So(handler.UserRequired, ShouldBeTrue)
	})

	Convey("create key required lambda", t, func() {
		handler := NewLambdaHandler(map[string]interface{}{
			"name":         "hello:world",
			"key_required": true,
		}, nil)

		So(handler.AccessKeyRequired, ShouldBeTrue)
	})
}

func TestLambdaHandler(t *testing.T) {
	Convey("Given a LambdaHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		transport := NewMockTransport(ctrl)
		conn := skydbtest.NewMapConn()

		handler := LambdaHandler{
			Plugin: &Plugin{
				transport: transport,
			},
			Name:       "hello:world",
			AssetStore: mock_asset.NewMockURLSignerStore(ctrl),
		}
		r := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
			p.DBConn = conn
			p.Context = context.Background()
			p.Context = context.WithValue(p.Context, HelloContextKey, "world")
		})
		Convey("should pass input and output", func(c C) {
			transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Do(
				func(ctx context.Context, name string, in []byte) {
					c.So(in, ShouldEqualJSON, `{"args":["bob"]}`)
				},
			).Return([]byte(`{"args":["bob"]}`), nil)

			resp := r.POST(`{
	"args": ["bob"]
}`)
			c.So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":{"args":["bob"]}}`)
		})

		Convey("should pass context", func(c C) {
			transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Do(
				func(ctx context.Context, name string, in []byte) {
					c.So(ctx.Value(HelloContextKey), ShouldEqual, "world")
				},
			).Return([]byte(`{}`), nil)

			r.POST(`{}`)
		})

		Convey("should return error", func(c C) {
			transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Return(
				[]byte{}, fmt.Errorf("an error"),
			)

			resp := r.POST(`{}`)
			c.So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error":{"code":10000,"message":"an error","name":"UnexpectedError"}
}`)
		})

		Convey("should return empty result", func(c C) {
			transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Return(
				[]byte("null"), nil,
			)

			resp := r.POST(`{}`)
			c.So(resp.Body.Bytes(), ShouldEqualJSON, `{}`)
		})
	})

	Convey("Given a LambdaHandler with various types", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := skydbtest.NewMapConn()
		transport := NewMockTransport(ctrl)

		handler := LambdaHandler{
			Plugin: &Plugin{
				transport: transport,
			},
			Name:       "hello:world",
			AssetStore: mock_asset.NewMockURLSignerStore(ctrl),
		}
		r := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
			p.DBConn = conn
			p.Context = context.Background()
			p.Context = context.WithValue(p.Context, HelloContextKey, "world")
		})

		fixtures := map[string]string{
			"integer": "42",
			"date":    `{"$type": "date", "$date": "2016-09-08T06:42:59.871181Z"}`,
			"array":   `["a", "b"]`,
			"map":     `{"a": 1, "b": 2}`,
		}

		for k, v := range fixtures {
			Convey(fmt.Sprintf("should pass %s in array args", k), func(c C) {
				transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Do(
					func(ctx context.Context, name string, in []byte) {
						c.So(in, ShouldEqualJSON, fmt.Sprintf(`{"args":[%s]}`, v))
					},
				).Return(
					[]byte(v), nil,
				)

				resp := r.POST(fmt.Sprintf(`{"args":[%s]}`, v))
				c.So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{"result":%v}`, v))
			})

			Convey(fmt.Sprintf("should pass %s in map args", k), func(c C) {
				transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Do(
					func(ctx context.Context, name string, in []byte) {
						c.So(in, ShouldEqualJSON, fmt.Sprintf(`{"args":{"first_arg":%s}}`, v))
					},
				).Return(
					[]byte(v), nil,
				)

				resp := r.POST(fmt.Sprintf(`{"args":{"first_arg":%s}}`, v))
				c.So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{"result":%v}`, v))
			})
		}
	})

	Convey("Given a LambdaHandler with database", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := skydbtest.NewMapConn()
		transport := NewMockTransport(ctrl)
		store := mock_asset.NewMockURLSignerStore(ctrl)

		handler := LambdaHandler{
			Plugin: &Plugin{
				transport: transport,
			},
			Name:       "hello:world",
			AssetStore: store,
		}
		r := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
			p.DBConn = conn
		})

		Convey("should complete and sign asset", func(c C) {
			assetName := "73ce6795-7304-476b-943e-aa33da076c31"
			minimalAssetPayload := fmt.Sprintf(`{
				"$type": "asset",
				"$name": "%s"
			}`, assetName)

			fullAssetPayload := fmt.Sprintf(`{
						"$content_type":"text/plain",
						"$name":"%s",
						"$type":"asset",
						"$url":"signed-url"
					}`, assetName)

			conn.AssetMap[assetName] = skydb.Asset{
				Name:        assetName,
				ContentType: "text/plain",
				Size:        12,
			}
			store.EXPECT().SignedURL(assetName).Return("signed-url", nil).Times(2)

			transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Do(
				func(ctx context.Context, name string, in []byte) {
					c.So(in, ShouldEqualJSON, fmt.Sprintf(`{"args":[%s]}`, fullAssetPayload))
				},
			).Return(
				[]byte(minimalAssetPayload), nil,
			)
			resp := r.POST(fmt.Sprintf(`{"args":[%s]}`, minimalAssetPayload))
			c.So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{"result":%s}`, fullAssetPayload))
		})

		Convey("should pass asset map verbatim if asset not found", func(c C) {
			assetName := "73ce6795-7304-476b-943e-aa33da076c31"
			minimalAssetPayload := fmt.Sprintf(`{
				"$type": "asset",
				"$name": "%s"
			}`, assetName)

			store.EXPECT().SignedURL(gomock.Any()).Return("", fmt.Errorf("not exist")).AnyTimes()

			transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Do(
				func(ctx context.Context, name string, in []byte) {
					c.So(in, ShouldEqualJSON, fmt.Sprintf(`{"args":[%s]}`, minimalAssetPayload))
				},
			).Return(
				[]byte(minimalAssetPayload), nil,
			)
			resp := r.POST(fmt.Sprintf(`{"args":[%s]}`, minimalAssetPayload))
			c.So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{"result":%s}`, minimalAssetPayload))
		})

		Convey("should pass minimal record content", func(c C) {
			recordID := "note/73ce6795-7304-476b-943e-aa33da076c31"
			minimalRecordPayload := fmt.Sprintf(`{
				"$type": "record",
				"$record": {
					"_id": "%s"
				}
			}`, recordID)

			fullRecordPayload := fmt.Sprintf(`{
				"$type": "record",
				"$record": {
					"_id": "%s",
					"_access": null,
					"_type": "record"
				}
			}`, recordID)

			transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Do(
				func(ctx context.Context, name string, in []byte) {
					c.So(in, ShouldEqualJSON, fmt.Sprintf(`{"args":[%s]}`, fullRecordPayload))
				},
			).Return(
				[]byte(minimalRecordPayload), nil,
			)
			resp := r.POST(fmt.Sprintf(`{"args":[%s]}`, minimalRecordPayload))
			c.So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{"result":%s}`, fullRecordPayload))
		})

		Convey("should pass record content", func(c C) {
			recordID := "note/73ce6795-7304-476b-943e-aa33da076c31"
			assetName := "73ce6795-7304-476b-943e-aa33da076c31"
			inputRecordPayload := fmt.Sprintf(`{
				"$type": "record",
				"$record": {
					"_id": "%s",
					"_ownerID": "john.doe@example.com",
					"_created_at": "2017-07-23T19:30:24Z",
					"_created_by": "john.doe@example.com",
					"_updated_at": "2017-07-23T19:30:24Z",
					"_updated_by": "john.doe@example.com",
					"_access": [{"relation": "friend", "level": "write"}],
					"asset": {"$type": "asset", "$name": "%s"},
					"_transient": {
						"asset": {"$type": "asset", "$name": "%s"}
					}
				}
			}`, recordID, assetName, assetName)
			outputRecordPayload := fmt.Sprintf(`{
				"$type": "record",
				"$record": {
					"_id": "%s",
					"_type": "record",
					"_ownerID": "john.doe@example.com",
					"_created_at": "2017-07-23T19:30:24Z",
					"_created_by": "john.doe@example.com",
					"_updated_at": "2017-07-23T19:30:24Z",
					"_updated_by": "john.doe@example.com",
					"_access": [{"relation": "friend", "level": "write"}],
					"asset": {"$content_type":"text/plain","$name":"%s","$type":"asset","$url":"signed-url"},
					"_transient": {
						"asset": {"$content_type":"text/plain","$name":"%s","$type":"asset","$url":"signed-url"}
					}
				}
			}`, recordID, assetName, assetName)

			conn.AssetMap[assetName] = skydb.Asset{
				Name:        assetName,
				ContentType: "text/plain",
				Size:        12,
			}
			store.EXPECT().SignedURL(assetName).Return("signed-url", nil).Times(4)

			transport.EXPECT().RunLambda(gomock.Any(), "hello:world", gomock.Any()).Do(
				func(ctx context.Context, name string, in []byte) {
					c.So(in, ShouldEqualJSON, fmt.Sprintf(`{"args":[%s]}`, outputRecordPayload))
				},
			).Return(
				[]byte(inputRecordPayload), nil,
			)
			resp := r.POST(fmt.Sprintf(`{"args":[%s]}`, inputRecordPayload))
			c.So(resp.Body.Bytes(), ShouldEqualJSON, fmt.Sprintf(`{"result":%s}`, outputRecordPayload))
		})
	})
}
