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

	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
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
	Convey("test args and stdout", t, func() {
		transport := &nullTransport{}
		plugin := Plugin{
			transport: transport,
		}
		handler := LambdaHandler{
			Plugin: &plugin,
			Name:   "hello:world",
		}
		r := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
			p.Context = context.Background()
			p.Context = context.WithValue(p.Context, HelloContextKey, "world")
		})

		Convey("handle", func() {
			resp := r.POST(`{
	"args": ["bob"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": {"args":["bob"]}
}`)
			So(transport.lastContext.Value(HelloContextKey), ShouldEqual, "world")
		})

	})

	Convey("test handle error", t, func() {
		transport := &fakeTransport{}
		transport.outErr = fmt.Errorf("an error")
		plugin := Plugin{
			transport: transport,
		}
		handler := LambdaHandler{
			Plugin: &plugin,
			Name:   "hello:world",
		}
		r := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
			p.Context = context.Background()
			p.Context = context.WithValue(p.Context, HelloContextKey, "world")
		})

		Convey("init", func() {
			resp := r.POST(`{
	"args": ["bob"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error":{"code":10000,"message":"an error","name":"UnexpectedError"}
}`)
			So(transport.lastContext.Value(HelloContextKey), ShouldEqual, "world")
		})

	})
}
