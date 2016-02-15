package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/oursky/skygear/skytest"
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

		mockHander := MockHander{
			outputs: mockResp,
		}

		r := NewRouter()
		r.Map("mock:map", &mockHander)
		mockJSON := `{"action": "mock:map"}`

		routeWithMiddleware := CORSMiddleware(r, testingCORSHost)

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

		Convey("Skip Handling if No CORS Host", func() {
			corsRequestMethod := "POST"
			corsRequestHeaders := "origin, content-type"

			routeWithEmptyCORSHost := CORSMiddleware(r, "")
			req, _ := http.NewRequest(
				"OPTIONS",
				"http://skygear.dev/",
				strings.NewReader(mockJSON),
			)
			req.Header.Set("Access-Control-Request-Method", corsRequestMethod)
			req.Header.Set("Access-Control-Request-Headers", corsRequestHeaders)

			resp := httptest.NewRecorder()
			routeWithEmptyCORSHost.ServeHTTP(resp, req)

			Convey("Should not set Access-Control-Allow-Origin header", func() {
				allowOrigin := resp.Header().Get("Access-Control-Allow-Origin")
				So(allowOrigin, ShouldEqual, "")
			})

			Convey("Should not set Access-Control-Allow-Methods header", func() {
				allowMethods := resp.Header().Get("Access-Control-Allow-Methods")
				So(allowMethods, ShouldEqual, "")
			})

			Convey("Should not set Access-Control-Allow-Headers header", func() {
				allowHeaders := resp.Header().Get("Access-Control-Allow-Headers")
				So(allowHeaders, ShouldEqual, "")
			})

			Convey("Should directly call next handler", func() {
				respContent := resp.Body.String()
				mockRespJSON, _ := json.Marshal(mockResp)
				So(respContent, ShouldEqualJSON, mockRespJSON)
			})
		})
	})
}
