package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRouterMatchByURL(t *testing.T) {
	Convey("Gateway", t, func() {

		Convey("matches simple url action", func() {
			g := NewGateway("endpoint")
			g.POST(NewFuncHandler(func(payload *Payload, resp *Response) {
				resp.WriteEntity(struct {
					Status string `json:"status"`
				}{"ok"})
			}))

			req, _ := http.NewRequest("POST", "http://skygear.test/endpoint", nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)
			So(w.Body.Bytes(), ShouldEqualJSON, `{"status": "ok"}`)
		})

		Convey("matches url with parameters", func() {
			g := NewGateway(`(?:entity|model)/([0-9a-zA-Z\-_]+)`)
			g.POST(NewFuncHandler(func(p *Payload, resp *Response) {
				resp.WriteEntity(p.Params)
			}))

			req, _ := http.NewRequest("POST", `http://skygear.test/model/Zz0-_`, nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			So(w.Body.Bytes(), ShouldEqualJSON, `["Zz0-_"]`)
		})

		Convey("help url handler fill in payload from Header", func() {
			g := NewGateway("endpoint")
			g.POST(NewFuncHandler(func(p *Payload, resp *Response) {
				resp.WriteEntity(struct {
					APIKey      string `json:"api-key"`
					AccessToken string `json:"access-token"`
				}{p.APIKey(), p.AccessToken()})
			}))

			req, _ := http.NewRequest("POST", `http://skygear.test/endpoint`, nil)
			req.Header.Add("X-Skygear-API-Key", "someapikey")
			req.Header.Add("X-Skygear-Access-Token", "someaccesstoken")

			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
                "api-key": "someapikey",
                "access-token": "someaccesstoken"
            }`)
		})

		Convey("fill in api key from query string", func() {
			g := NewGateway("endpoint")
			g.POST(NewFuncHandler(func(p *Payload, resp *Response) {
				resp.WriteEntity(struct {
					APIKey      string `json:"api-key"`
					AccessToken string `json:"access-token"`
				}{p.APIKey(), p.AccessToken()})
			}))

			req, _ := http.NewRequest("POST", `http://skygear.test/endpoint?api_key=someapikey&access_token=someaccesstoken`, nil)

			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
                "api-key": "someapikey",
                "access-token": "someaccesstoken"
            }`)
		})

		Convey("fill in api key from url encoded form", func() {
			g := NewGateway("endpoint")
			g.POST(NewFuncHandler(func(p *Payload, resp *Response) {
				resp.WriteEntity(struct {
					APIKey      string `json:"api-key"`
					AccessToken string `json:"access-token"`
				}{p.APIKey(), p.AccessToken()})
			}))

			req, _ := http.NewRequest("POST", `http://skygear.test/endpoint`, strings.NewReader("api_key=someapikey&access_token=someaccesstoken"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
                "api-key": "someapikey",
                "access-token": "someaccesstoken"
            }`)
		})
	})
}
