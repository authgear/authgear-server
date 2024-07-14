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

func TestAccountManagementV1IdentificationHandlerRequestValidation(t *testing.T) {
	Convey("AccountManagementV1IdentificationHandler request validation", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		jsonResponseWriter := httputil.JSONResponseWriter{}
		svc := NewMockAccountManagementV1IdentificationHandlerService(ctrl)
		h := AccountManagementV1IdentificationHandler{
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
                            "alias",
                            "identification",
                            "redirect_uri"
                        ],
                        "missing": [
                            "alias",
                            "identification",
                            "redirect_uri"
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
	"identification": "oauth",
	"alias": "google",
	"redirect_uri": "myapp.com://host/path"
}
			`))
			r.Header.Set("Content-Type", "application/json")
			mockSession := sessiontest.NewMockSession()
			r = mockSession.ToRequest(r)
			w := httptest.NewRecorder()

			svc.EXPECT().StartAdding(gomock.Any()).Times(1).Return(&accountmanagement.StartAddingOutput{
				Token:            "token",
				AuthorizationURL: "https://google.com",
			}, nil)
			h.ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
			So(w.Body.String(), ShouldEqualJSON, `{
				"result": {
					"token": "token",
					"authorization_url": "https://google.com"
				}
			}`)
		})
	})
}
