package router

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/oursky/ourd/oderr"
)

type MockHander struct {
	outputs Response
}

func (m *MockHander) handle(p *Payload, r *Response) {
	r.Result = m.outputs.Result
	return
}

type ErrHandler struct {
	Err error
}

func (h *ErrHandler) handle(p *Payload, r *Response) {
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
	mockHander := MockHander{
		outputs: mockResp,
	}
	r := NewRouter()
	r.Map("mock:map", mockHander.handle)
	var mockJSON = `{
	"action": "mock:map"
}`

	req, _ := http.NewRequest(
		"POST",
		"http://ourd.dev/api/v1",
		strings.NewReader(mockJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Body.String() != "{\"result\":{\"username\":\"example\"}}\n" {
		t.Fatalf("Simple map failed %v", resp.Body.String())
	}
}

func TestRouterMapMissing(t *testing.T) {
	mockHander := MockHander{
		outputs: Response{},
	}
	r := NewRouter()
	r.Map("mock:map", mockHander.handle)
	var mockJSON = `{
	"action": "missing"
}`

	req, _ := http.NewRequest(
		"POST",
		"http://ourd.dev/api/v1",
		strings.NewReader(mockJSON),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	expectedBody := `{"error":{"type":"RequestInvalid","code":101,"message":"route unmatched"}}
`
	if resp.Body.String() != expectedBody {
		t.Fatalf("want resp.Body.String() = %#v, got %#v", expectedBody, resp.Body.String())
	}
}

type getPreprocessor struct {
	Status int
	Err    error
}

func (p *getPreprocessor) Preprocess(payload *Payload, response *Response) int {
	response.Err = p.Err
	return p.Status
}

func TestErrorHandler(t *testing.T) {
	Convey("Router", t, func() {
		r := NewRouter()

		Convey("returns 400 if handler sets Response.Err", func() {
			errHandler := &ErrHandler{
				Err: errors.New("some error"),
			}
			r.Map("mock:handler", errHandler.handle)

			req, _ := http.NewRequest(
				"POST",
				"http://ourd.dev/api/v1",
				strings.NewReader(`{"action": "mock:handler"}`),
			)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestPreprocess(t *testing.T) {
	r := NewRouter()
	mockHandler := MockHander{outputs: Response{
		Result: "ok",
	}}
	mockPreprocessor := getPreprocessor{}

	r.Map("mock:preprocess", mockHandler.handle, mockPreprocessor.Preprocess)

	Convey("Given a router with a preprocessor", t, func() {
		req, _ := http.NewRequest(
			"POST",
			"http://ourd.dev/api/v1",
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
			mockPreprocessor.Err = errors.New("err")

			r.ServeHTTP(resp, req)

			Convey("it has \"err\" as body", func() {
				So(resp.Body.String(), ShouldEqual, `{"error":{"type":"UnknownError","code":1,"message":"err"}}
`)
			})

			Convey("it has status code = 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When preprocessor gives an oderr preprocessor", func() {
			mockPreprocessor.Status = http.StatusInternalServerError
			mockPreprocessor.Err = oderr.NewUnknownErr(errors.New("oderr"))

			r.ServeHTTP(resp, req)

			Convey("it has \"err\" as body", func() {
				So(resp.Body.String(), ShouldEqual, `{"error":{"type":"UnknownError","code":1,"message":"oderr"}}
`)
			})

			Convey("it has status code = 500", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

// TODO(limouren): fix this test case
// func TestRouterCheckAuth(t *testing.T) {
// 	mockResp := Response{}
// 	type exampleResp struct {
// 		Auth string `json:"auth"`
// 	}
// 	mockResp.Result = exampleResp{
// 		Auth: "ok",
// 	}
// 	mockHander := MockHander{
// 		outputs: mockResp,
// 	}
// 	r := NewRouter()
// 	r.Map("mock:map", mockHander.handle)
// 	r.Preprocess(CheckAuth)

// 	// Positive test
// 	var mockJSON = `{
// 	"action": "mock:map",
// 	"access_token": "validToken"
// }`

// 	req, _ := http.NewRequest(
// 		"POST",
// 		"http://ourd.dev/api/v1",
// 		strings.NewReader(mockJSON),
// 	)
// 	req.Header.Set("Content-Type", "application/json")
// 	resp := httptest.NewRecorder()
// 	r.ServeHTTP(resp, req)

// 	if resp.Body.String() != "{\"result\":{\"auth\":\"ok\"}}" {
// 		t.Fatalf("CheckAuth failed: %v", resp.Body.String())
// 	}

// 	// Negative test
// 	var mockJSON2 = `{
// 	"action": "mock:map",
// 	"access_token": "InValidToken"
// }`

// 	req2, _ := http.NewRequest(
// 		"POST",
// 		"http://ourd.dev/api/v1",
// 		strings.NewReader(mockJSON2),
// 	)
// 	req2.Header.Set("Content-Type", "application/json")
// 	resp2 := httptest.NewRecorder()
// 	r.ServeHTTP(resp2, req2)

// 	if resp2.Body.String() != "Unauthorized request" {
// 		t.Fatalf(
// 			"CheckAuth failed to reject unauthorized request: %v",
// 			resp2.Body.String())
// 	}
// }
