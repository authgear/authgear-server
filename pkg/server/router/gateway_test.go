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

package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRouterMatchByURL(t *testing.T) {
	Convey("Gateway", t, func() {

		Convey("matches simple url action", func() {
			g := NewGateway("endpoint", "/endpoint", "tag", nil)
			g.POST(NewFuncHandler(func(payload *Payload, resp *Response) {
				writeEntity(resp.Writer(), struct {
					Status string `json:"status"`
				}{"ok"})
			}))

			req, _ := http.NewRequest("POST", "http://skygear.test/endpoint", nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)
			So(w.Body.Bytes(), ShouldEqualJSON, `{"status": "ok"}`)
		})

		Convey("don't matches simple url action with wrong method", func() {
			g := NewGateway("endpoint", "/endpoint", "tag", nil)
			g.POST(NewFuncHandler(func(payload *Payload, resp *Response) {
				writeEntity(resp.Writer(), struct {
					Status string `json:"status"`
				}{"ok"})
			}))

			req, _ := http.NewRequest("GET", "http://skygear.test/endpoint", nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 404)
		})

		Convey("matches simple url action with passin ServeMux", func() {
			mux := http.NewServeMux()
			g := NewGateway("endpoint", "/endpoint", "tag", mux)
			g.POST(NewFuncHandler(func(payload *Payload, resp *Response) {
				writeEntity(resp.Writer(), struct {
					Status string `json:"status"`
				}{"ok"})
			}))

			req, _ := http.NewRequest("POST", "http://skygear.test/endpoint", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			So(w.Body.Bytes(), ShouldEqualJSON, `{"status": "ok"}`)
		})

		Convey("don't matches simple url action with passin ServeMux", func() {
			mux := http.NewServeMux()
			g := NewGateway("endpoint", "/endpoint", "tag", mux)
			g.POST(NewFuncHandler(func(payload *Payload, resp *Response) {
				writeEntity(resp.Writer(), struct {
					Status string `json:"status"`
				}{"ok"})
			}))

			req, _ := http.NewRequest("POST", "http://skygear.test/404", nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, 404)
		})

		Convey("matches url with parameters", func() {
			g := NewGateway(`(?:entity|model)/([0-9a-zA-Z\-_]+)`, "/model", "tag", nil)
			g.POST(NewFuncHandler(func(p *Payload, resp *Response) {
				writeEntity(resp.Writer(), p.Params)
			}))

			req, _ := http.NewRequest("POST", `http://skygear.test/model/Zz0-_`, nil)
			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			So(w.Body.Bytes(), ShouldEqualJSON, `["Zz0-_"]`)
		})

		Convey("help url handler fill in payload from Header", func() {
			g := NewGateway("endpoint", "/endpoint", "tag", nil)
			g.POST(NewFuncHandler(func(p *Payload, resp *Response) {
				writeEntity(resp.Writer(), struct {
					APIKey      string `json:"api-key"`
					AccessToken string `json:"access-token"`
				}{p.APIKey(), p.AccessTokenString()})
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
			g := NewGateway("endpoint", "/endpoint", "tag", nil)
			g.POST(NewFuncHandler(func(p *Payload, resp *Response) {
				writeEntity(resp.Writer(), struct {
					APIKey      string `json:"api-key"`
					AccessToken string `json:"access-token"`
				}{p.APIKey(), p.AccessTokenString()})
			}))

			req, _ := http.NewRequest("POST", `http://skygear.test/endpoint?api_key=someapikey&access_token=someaccesstoken`, nil)

			w := httptest.NewRecorder()
			g.ServeHTTP(w, req)

			So(w.Body.Bytes(), ShouldEqualJSON, `{
                "api-key": "someapikey",
                "access-token": "someaccesstoken"
            }`)
		})

	})
}
