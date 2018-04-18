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

	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
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

		handler := LambdaHandler{
			Plugin: &Plugin{
				transport: transport,
			},
			Name: "hello:world",
		}
		r := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
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

		transport := NewMockTransport(ctrl)

		handler := LambdaHandler{
			Plugin: &Plugin{
				transport: transport,
			},
			Name: "hello:world",
		}
		r := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
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
}
