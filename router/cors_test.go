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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/skygeario/skygear-server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCORSMiddleware(t *testing.T) {
	Convey("CORS Middleware", t, func() {
		type testingResp struct {
			Message string `json:"message"`
		}

		testingRespMsg := "Testing Response For CORS Test"
		testingCORSHost := "http://skygear.dev2"

		mockResp := Response{}
		mockResp.Result = testingResp{
			Message: testingRespMsg,
		}

		mockHandler := MockHandler{
			outputs: mockResp,
		}

		r := NewRouter()
		r.Map("mock:map", &mockHandler)
		mockJSON := `{"action": "mock:map"}`

		routeWithMiddleware := &CORSMiddleware{
			Next:   r,
			Origin: testingCORSHost,
		}

		Convey("Handle Preflight Request", func() {
			corsRequestMethod := "POST"
			corsRequestHeaders := "origin, content-type"

			req, _ := http.NewRequest(
				"OPTIONS",
				"http://skygear.dev/",
				strings.NewReader(mockJSON),
			)
			req.Header.Set("Access-Control-Request-Method", corsRequestMethod)
			req.Header.Set("Access-Control-Request-Headers", corsRequestHeaders)

			resp := httptest.NewRecorder()
			routeWithMiddleware.ServeHTTP(resp, req)

			Convey("Should set Access-Control-Allow-Origin header", func() {
				allowOrigin := resp.Header().Get("Access-Control-Allow-Origin")
				So(allowOrigin, ShouldEqual, testingCORSHost)
			})

			Convey("Should set Access-Control-Allow-Methods header", func() {
				allowMethods := resp.Header().Get("Access-Control-Allow-Methods")
				So(allowMethods, ShouldEqual, corsRequestMethod)
			})

			Convey("Should set Access-Control-Allow-Headers header", func() {
				allowHeaders := resp.Header().Get("Access-Control-Allow-Headers")
				So(allowHeaders, ShouldEqual, corsRequestHeaders)
			})

			Convey("Should not call next handler", func() {
				respContent := resp.Body.String()
				So(respContent, ShouldNotContainSubstring, testingRespMsg)
			})
		})

		Convey("Handle Normal Request", func() {
			req, _ := http.NewRequest(
				"POST",
				"http://skygear.dev/",
				strings.NewReader(mockJSON),
			)

			resp := httptest.NewRecorder()
			routeWithMiddleware.ServeHTTP(resp, req)

			Convey("Should call next handler", func() {
				mockRespJSON, _ := json.Marshal(mockResp)
				So(resp.Body.String(), ShouldEqualJSON, mockRespJSON)
			})
		})

	})
}
