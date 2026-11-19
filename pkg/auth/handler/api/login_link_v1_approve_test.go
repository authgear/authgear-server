package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	authapi "github.com/authgear/authgear-server/pkg/auth/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

// stubLoginLinkService records the codes it receives. The generated mock of
// authapi.APIControllerLoginLinkService is test-only in its own package.
type stubLoginLinkService struct {
	err      error
	inspects []string
	approves []string
}

func (s *stubLoginLinkService) Inspect(ctx context.Context, code string) error {
	s.inspects = append(s.inspects, code)
	return s.err
}

func (s *stubLoginLinkService) Approve(ctx context.Context, code string) error {
	s.approves = append(s.approves, code)
	return s.err
}

func serveLoginLinkV1(h http.Handler, contentType string, body string) (int, map[string]any) {
	r, _ := http.NewRequest("POST", "/", strings.NewReader(body))
	r.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	So(err, ShouldBeNil)
	return w.Result().StatusCode, resp
}

func TestLoginLinkV1ApproveHandler(t *testing.T) {
	Convey("LoginLinkV1ApproveHandler", t, func() {
		service := &stubLoginLinkService{}
		h := &LoginLinkV1ApproveHandler{
			Controller: &authapi.APIController{LoginLinks: service},
		}

		Convey("approves the code", func() {
			status, resp := serveLoginLinkV1(h, "application/json", `{"code": "CODE"}`)
			So(status, ShouldEqual, 200)
			So(resp, ShouldResemble, map[string]any{"result": map[string]any{}})
			So(service.approves, ShouldResemble, []string{"CODE"})
			So(service.inspects, ShouldBeEmpty)
		})

		Convey("returns the service error", func() {
			service.err = otp.ErrConsumedCode

			status, resp := serveLoginLinkV1(h, "application/json", `{"code": "CODE"}`)
			So(status, ShouldEqual, 403)
			So(resp["error"].(map[string]any)["reason"], ShouldEqual, "InvalidOTPCode")
		})

		Convey("requires code", func() {
			status, resp := serveLoginLinkV1(h, "application/json", `{}`)
			So(status, ShouldEqual, 400)
			So(resp["error"].(map[string]any)["reason"], ShouldEqual, "ValidationFailed")
			So(service.approves, ShouldBeEmpty)
		})

		Convey("does not pass a malformed body on", func() {
			for _, body := range []string{`{`, `not json`, ``, `[]`} {
				status, _ := serveLoginLinkV1(h, "application/json", body)
				So(status, ShouldNotEqual, 200)
			}
			So(service.approves, ShouldBeEmpty)
		})

		Convey("requires a JSON body", func() {
			status, _ := serveLoginLinkV1(h, "text/plain", `{"code": "CODE"}`)
			So(status, ShouldEqual, 400)
			So(service.approves, ShouldBeEmpty)
		})
	})
}
