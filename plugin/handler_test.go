package plugin

import (
	"encoding/json"
	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
	"golang.org/x/net/context"
)

func TestHandlerCreation(t *testing.T) {
	Convey("create simple handler", t, func() {
		handler := NewPluginHandler(pluginHandlerInfo{
			Name: "hello:world",
		}, nil, nil)

		So(handler.Name, ShouldEqual, "hello:world")
		So(handler.AccessKeyRequired, ShouldBeFalse)
		So(handler.UserRequired, ShouldBeFalse)
	})

	Convey("create user required handler", t, func() {
		handler := NewPluginHandler(pluginHandlerInfo{
			Name:         "hello:world",
			UserRequired: true,
		}, nil, nil)

		So(handler.UserRequired, ShouldBeTrue)
	})

	Convey("create key required Handler", t, func() {
		handler := NewPluginHandler(pluginHandlerInfo{
			Name:        "hello:world",
			KeyRequired: true,
		}, nil, nil)

		So(handler.AccessKeyRequired, ShouldBeTrue)
	})
}

func TestHandler(t *testing.T) {
	Convey("test reponse header and body", t, func() {
		transport := &fakeTransport{}
		transport.outBytes, _ = json.Marshal(struct {
			Header map[string][]string `json:"header"`
			Body   []byte              `json:"body"`
		}{
			Header: map[string][]string{
				"X-Skygear": []string{"Chima"},
			},
			Body: []byte(`{"kind": "I can be anything"}`),
		})
		plugin := Plugin{
			transport: transport,
		}
		handler := Handler{
			Plugin: &plugin,
			Name:   "hello:world",
		}
		g := handlertest.NewMockGateway(
			"",
			"",
			[]string{"POST"},
			&handler,
			func(p *router.Payload) {
				p.Context = context.Background()
				p.Context = context.WithValue(p.Context, "hello", "world")
			},
		)

		Convey("handle", func() {
			resp := g.Request("POST", `{
	"args": ["bob"]
}`)

			So(resp.Header().Get("X-Skygear"), ShouldEqual, "Chima")
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"kind": "I can be anything"
}`)
			So(transport.lastContext.Value("hello"), ShouldEqual, "world")
		})

	})
}
