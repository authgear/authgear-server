package plugin

import (
	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"fmt"
	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
	"golang.org/x/net/context"
)

type fakeTransport struct {
	nullTransport
	outBytes []byte
	outErr   error
}

func (t *fakeTransport) RunLambda(ctx context.Context, name string, in []byte) (out []byte, err error) {
	t.lastContext = ctx
	if t.outErr == nil {
		out = t.outBytes
	} else {
		err = t.outErr
	}
	return
}

func TestLambdaCreation(t *testing.T) {
	Convey("create simple lambda", t, func() {
		handler := NewLambdaHandler(map[string]interface{}{
			"name": "hello:world",
		}, nil, nil)

		So(handler.Name, ShouldEqual, "hello:world")
		So(handler.AccessKeyRequired, ShouldBeFalse)
		So(handler.UserRequired, ShouldBeFalse)
	})

	Convey("create user required lambda", t, func() {
		handler := NewLambdaHandler(map[string]interface{}{
			"name":          "hello:world",
			"user_required": true,
		}, nil, nil)

		So(handler.UserRequired, ShouldBeTrue)
	})

	Convey("create key required lambda", t, func() {
		handler := NewLambdaHandler(map[string]interface{}{
			"name":         "hello:world",
			"key_required": true,
		}, nil, nil)

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
			p.Context = context.WithValue(p.Context, "hello", "world")
		})

		Convey("handle", func() {
			resp := r.POST(`{
	"args": ["bob"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": {"args":["bob"]}
}`)
			So(transport.lastContext.Value("hello"), ShouldEqual, "world")
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
			p.Context = context.WithValue(p.Context, "hello", "world")
		})

		Convey("init", func() {
			resp := r.POST(`{
	"args": ["bob"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error":{"code":10000,"message":"an error","name":"UnexpectedError"}
}`)
			So(transport.lastContext.Value("hello"), ShouldEqual, "world")
		})

	})
}
