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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

type MockHandler struct {
	outputs Response
	delay   time.Duration
}

func (m *MockHandler) Setup() {
	return
}

func (m *MockHandler) GetPreprocessors() []Processor {
	return nil
}

func (m *MockHandler) Handle(p *Payload, r *Response) {
	if m.delay.Seconds() > 0 {
		<-time.After(m.delay)
	}
	r.Result = m.outputs.Result
	return
}

type ErrHandler struct {
	Err skyerr.Error
}

func (h *ErrHandler) Setup() {
	return
}

func (h *ErrHandler) GetPreprocessors() []Processor {
	return nil
}

func (h *ErrHandler) Handle(p *Payload, r *Response) {
	r.Err = h.Err
}

func TestRouterMap(t *testing.T) {
	mockResp := Response{}
	type exampleResp struct {
		Username string `json:"username"`
	}
	mockResp.Result = exampleResp{
		Username: "example",
	}
	mockHandler := MockHandler{
		outputs: mockResp,
	}
	r := NewRouter()
	r.Map("mock:map", &mockHandler)
	var mockJSON = `{
	"action": "mock:map"
}`

	req, _ := http.NewRequest(
		"POST",
		"http://skygear.dev/",
		strings.NewReader(mockJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	contentTypeField := resp.Header().Get("Content-Type")
	if contentTypeField != "application/json" {
		t.Fatalf("Unexpected Content-Type in header %v", contentTypeField)
	}

	serverField := resp.Header().Get("Server")
	if !strings.HasPrefix(serverField, "Skygear Server/") {
		t.Fatalf("Unexpected Server in header %v", serverField)
	}

	responseBody := resp.Body.String()
	if responseBody != "{\"result\":{\"username\":\"example\"}}\n" {
		t.Fatalf("Simple map failed %v", responseBody)
	}
}

func TestRouterMapMissing(t *testing.T) {
	mockHandler := MockHandler{
		outputs: Response{},
	}
	r := NewRouter()
	r.Map("mock:map", &mockHandler)
	var mockJSON = `{
	"action": "missing"
}`

	req, _ := http.NewRequest(
		"POST",
		"http://skygear.dev/",
		strings.NewReader(mockJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	expectedBody := `{"error":{"name":"UndefinedOperation","code":117,"message":"route unmatched"}}
`
	serverField := resp.Header().Get("Server")
	if !strings.HasPrefix(serverField, "Skygear Server/") {
		t.Fatalf("Unexpected Server in header %v", serverField)
	}

	responseBody := resp.Body.String()
	if responseBody != expectedBody {
		t.Fatalf("want resp.Body.String() = %#v, got %#v", expectedBody, responseBody)
	}
}

type getPreprocessor struct {
	Status int
	Err    skyerr.Error
}

func (p getPreprocessor) Preprocess(payload *Payload, response *Response) int {
	response.Err = p.Err
	return p.Status
}

func TestURLRouting(t *testing.T) {
	Convey("URL Routing", t, func() {
		Convey("can route to correct handler using URL, instead of payload", func() {
			mockResp := Response{}
			mockResp.Result = struct {
				Message string `json:"message"`
			}{
				Message: "Got it.",
			}

			mockHandler := MockHandler{
				outputs: mockResp,
			}

			r := NewRouter()
			r.Map("mock:map", &mockHandler)

			req, _ := http.NewRequest(
				"POST",
				"http://skygear.dev/mock/map",
				strings.NewReader(""),
			)

			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Body.String(), ShouldEqual, `{"result":{"message":"Got it."}}
`)
		})
	})
}

func TestErrorHandler(t *testing.T) {
	Convey("Router", t, func() {
		r := NewRouter()

		Convey("returns 500 if handler sets Response.Err with unexpected error", func() {
			errHandler := &ErrHandler{
				Err: skyerr.NewError(skyerr.UnexpectedError, "some error"),
			}
			r.Map("mock:handler", errHandler)

			req, _ := http.NewRequest(
				"POST",
				"http://skygear.dev/",
				strings.NewReader(`{"action": "mock:handler"}`),
			)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestPreprocess(t *testing.T) {
	r := NewRouter()
	mockHandler := MockHandler{outputs: Response{
		Result: "ok",
	}}
	mockPreprocessor := getPreprocessor{}

	r.Map("mock:preprocess", &mockHandler, &mockPreprocessor)

	Convey("Given a router with a preprocessor", t, func() {
		req, _ := http.NewRequest(
			"POST",
			"http://skygear.dev/",
			strings.NewReader(`{"action": "mock:preprocess"}`),
		)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		Convey("When preprocessor lets go", func() {
			mockPreprocessor.Status = http.StatusOK
			mockPreprocessor.Err = nil

			r.ServeHTTP(resp, req)

			Convey("it returns ok", func() {
				So(resp.Body.String(), ShouldEqual, "{\"result\":\"ok\"}\n")
			})

			Convey("it has status code = 200", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When preprocessor gives an err", func() {
			mockPreprocessor.Status = http.StatusInternalServerError
			mockPreprocessor.Err = skyerr.NewError(skyerr.UnexpectedError, "err")

			r.ServeHTTP(resp, req)

			Convey("it has \"err\" as body", func() {
				So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"name":"UnexpectedError","code":10000,"message":"err"}}
`)
			})

			Convey("it has status code = 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When preprocessor gives an skyerr preprocessor", func() {
			mockPreprocessor.Status = http.StatusInternalServerError
			mockPreprocessor.Err = skyerr.MakeError(errors.New("skyerr"))

			r.ServeHTTP(resp, req)

			Convey("it has \"err\" as body", func() {
				So(resp.Body.Bytes(), ShouldEqualJSON, `{"error":{"name":"UnexpectedError","code":10000,"message":"skyerr"}}
`)
			})

			Convey("it has status code = 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestPreprocessorRegistry(t *testing.T) {
	mockPreprocessor := &getPreprocessor{}

	Convey("Register preprocessor", t, func() {
		reg := PreprocessorRegistry{}
		reg["mock"] = mockPreprocessor
		So(reg["mock"], ShouldEqual, mockPreprocessor)
	})

	Convey("Get by names", t, func() {
		reg := PreprocessorRegistry{}
		reg["mock"] = mockPreprocessor
		preprocessors := reg.GetByNames("mock")
		So(len(preprocessors), ShouldEqual, 1)
		So(preprocessors[0], ShouldEqual, mockPreprocessor)
	})
}

func TestTimeout(t *testing.T) {
	Convey("Router", t, func() {
		r := NewRouter()
		r.ResponseTimeout = 10 * time.Millisecond

		Convey("Return 503 when request timed out.", func() {
			delayHandler := &MockHandler{
				outputs: Response{
					Result: map[string]string{
						"hello": "wait",
					},
				},
				delay: 20 * time.Millisecond,
			}
			r.Map("delay:handler", delayHandler)

			req, _ := http.NewRequest(
				"POST",
				"http://skygear.dev/",
				strings.NewReader(`{"action": "delay:handler"}`),
			)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, http.StatusServiceUnavailable)
		})
	})
}
