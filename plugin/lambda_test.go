package plugin

import (
	. "github.com/oursky/skygear/ourtest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"fmt"
	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
)

type fakeTransport struct {
	nullTransport
	outBytes []byte
	outErr   error
}

func (t fakeTransport) RunLambda(name string, in []byte) (out []byte, err error) {
	if t.outErr == nil {
		out = t.outBytes
	} else {
		err = t.outErr
	}
	return
}

func TestLambdaHandler(t *testing.T) {
	Convey("test args and stdout", t, func() {
		transport := nullTransport{}
		plugin := Plugin{
			transport: transport,
		}
		r := handlertest.NewSingleRouteRouter(CreateLambdaHandler(&plugin, "hello:world"), func(p *router.Payload) {
		})

		Convey("handle", func() {
			resp := r.POST(`{
	"args": ["bob"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": {"args":["bob"]}
}`)

		})

	})

	Convey("test handle error", t, func() {
		transport := fakeTransport{}
		transport.outErr = fmt.Errorf("an error")
		plugin := Plugin{
			transport: transport,
		}
		r := handlertest.NewSingleRouteRouter(CreateLambdaHandler(&plugin, "hello:world"), func(p *router.Payload) {
		})

		Convey("init", func() {
			resp := r.POST(`{
	"args": ["bob"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error":{"code":1,"message":"an error","type":"UnknownError"}
}`)

		})

	})
}
