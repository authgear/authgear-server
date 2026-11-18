package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestLoginLinkV1ApproveHandler(t *testing.T) {
	Convey("LoginLinkV1ApproveHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		otpCodes := NewMockLoginLinkV1ApproveOTPCodeService(ctrl)
		events := NewMockLoginLinkV1ApproveEventStore(ctrl)
		h := &LoginLinkV1ApproveHandler{
			Config:   &config.AppConfig{},
			OTPCodes: otpCodes,
			Events:   events,
		}

		serve := func(body string) (int, map[string]any) {
			r, _ := http.NewRequest("POST", "/api/v1/login_link/approve", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			var resp map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			So(err, ShouldBeNil)
			return w.Result().StatusCode, resp
		}

		Convey("approves the code and refreshes the flow", func() {
			otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "CODE").Return("user@example.com", nil)
			otpCodes.EXPECT().VerifyOTP(gomock.Any(), gomock.Any(), "user@example.com", "CODE", &otp.VerifyOptions{SkipConsume: true}).Return(nil)
			otpCodes.EXPECT().SetSubmittedCode(gomock.Any(), gomock.Any(), "user@example.com", "CODE").Return(&otp.State{
				AuthenticationFlowWebsocketChannelName: "channel",
			}, nil)
			events.EXPECT().Publish(gomock.Any(), "channel", authflow.NewEventRefresh()).Return(nil)

			status, resp := serve(`{"code": "CODE"}`)
			So(status, ShouldEqual, 200)
			So(resp, ShouldResemble, map[string]any{"result": map[string]any{}})
		})

		Convey("does not publish without a websocket channel", func() {
			otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "CODE").Return("user@example.com", nil)
			otpCodes.EXPECT().VerifyOTP(gomock.Any(), gomock.Any(), "user@example.com", "CODE", gomock.Any()).Return(nil)
			otpCodes.EXPECT().SetSubmittedCode(gomock.Any(), gomock.Any(), "user@example.com", "CODE").Return(&otp.State{}, nil)

			status, _ := serve(`{"code": "CODE"}`)
			So(status, ShouldEqual, 200)
		})

		Convey("rejects an unknown code", func() {
			otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "UNKNOWN").Return("", otp.ErrCodeNotFound)

			status, resp := serve(`{"code": "UNKNOWN"}`)
			So(status, ShouldEqual, 403)
			So(resp["error"].(map[string]any)["reason"], ShouldEqual, "InvalidOTPCode")
		})

		Convey("rejects a consumed code", func() {
			otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "CODE").Return("user@example.com", nil)
			otpCodes.EXPECT().VerifyOTP(gomock.Any(), gomock.Any(), "user@example.com", "CODE", gomock.Any()).Return(otp.ErrConsumedCode)

			status, resp := serve(`{"code": "CODE"}`)
			So(status, ShouldEqual, 403)
			So(resp["error"].(map[string]any)["reason"], ShouldEqual, "InvalidOTPCode")
		})

		Convey("returns rate limit errors", func() {
			otpCodes.EXPECT().LookupCode(gomock.Any(), otp.PurposeOOBOTP, "CODE").Return("user@example.com", nil)
			otpCodes.EXPECT().VerifyOTP(gomock.Any(), gomock.Any(), "user@example.com", "CODE", gomock.Any()).Return(otp.ErrTooManyAttempts)

			status, resp := serve(`{"code": "CODE"}`)
			So(status, ShouldEqual, 429)
			So(resp["error"].(map[string]any)["reason"], ShouldEqual, "RateLimited")
		})

		Convey("requires code", func() {
			status, resp := serve(`{}`)
			So(status, ShouldEqual, 400)
			So(resp["error"].(map[string]any)["reason"], ShouldEqual, "ValidationFailed")
		})
	})
}
