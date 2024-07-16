package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	sessiontest "github.com/authgear/authgear-server/pkg/lib/session/test"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestAccountManagementV1IdentificationOAuthHandlerRequestValidation(t *testing.T) {
	Convey("AccountManagementV1IdentificationOAuthHandler request validation", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		jsonResponseWriter := httputil.JSONResponseWriter{}
		svc := NewMockAccountManagementV1IdentificationOAuthHandlerService(ctrl)
		h := AccountManagementV1IdentificationOAuthHandler{
			JSON:    &jsonResponseWriter,
			Service: svc,
		}

		Convey("empty object", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader("{}"))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 400)
			So(w.Body.String(), ShouldEqualJSON, `
{
    "error": {
        "name": "Invalid",
        "reason": "ValidationFailed",
        "message": "invalid request body",
        "code": 400,
        "info": {
            "causes": [
                {
                    "location": "",
                    "kind": "required",
                    "details": {
                        "actual": null,
                        "expected": [
                            "query",
                            "token"
                        ],
                        "missing": [
                            "query",
                            "token"
                        ]
                    }
                }
            ]
        }
    }
}
		`)
		})

		Convey("valid", func() {
			r, _ := http.NewRequest("POST", "", strings.NewReader(`
{
	"token": "token",
	"query": "?code=code"
}
			`))
			r.Header.Set("Content-Type", "application/json")
			mockSession := sessiontest.NewMockSession()
			r = mockSession.ToRequest(r)
			w := httptest.NewRecorder()

			svc.EXPECT().FinishAdding(gomock.Any()).Times(1).Return(&accountmanagement.FinishAddingOutput{}, nil)
			h.ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})
}
