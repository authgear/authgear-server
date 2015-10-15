package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/oursky/skygear/ourtest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRouterMatchByURL(t *testing.T) {
	Convey("Gateway", t, func() {

		Convey("matches simple url action", func() {
			g := NewGateway("endpoint")
			g.POST(func(payload *Payload, resp *Response) {
				resp.WriteEntity(struct {
					Status string `json:"status"`
				}{"ok"})
			})

			req, _ := http.NewRequest("POST", "http://ourd.test/endpoint", nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)
			So(w.Body.Bytes(), ShouldEqualJSON, `{"status": "ok"}`)
		})

		Convey("matches url with parameters", func() {
			g := NewGateway(`(?:entity|model)/([0-9a-zA-Z\-_]+)`)
			g.POST(func(p *Payload, resp *Response) {
				resp.WriteEntity(p.Params)
			})

			req, _ := http.NewRequest("POST", `http://ourd.test/model/Zz0-_`, nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			So(w.Body.Bytes(), ShouldEqualJSON, `["Zz0-_"]`)
		})

		Convey("help url handler fill in payload from Header", func() {
			g := NewGateway("endpoint")
			g.POST(func(p *Payload, resp *Response) {
				resp.WriteEntity(struct {
					APIKey      string `json:"api-key"`
					AccessToken string `json:"access-token"`
				}{p.APIKey(), p.AccessToken()})
			})

			req, _ := http.NewRequest("POST", `http://ourd.test/endpoint`, nil)
			req.Header.Add("X-Ourd-API-Key", "someapikey")
			req.Header.Add("X-Ourd-Access-Token", "someaccesstoken")

			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
                "api-key": "someapikey",
                "access-token": "someaccesstoken"
            }`)
		})
	})
}
