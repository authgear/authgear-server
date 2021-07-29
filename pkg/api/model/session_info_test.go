package model_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
)

func TestSessionInfo(t *testing.T) {
	Convey("SessionInfo", t, func() {
		Convey("should write to HTTP headers correctly", func() {
			rw := httptest.NewRecorder()

			Convey("invalid auth", func() {
				var i = &model.SessionInfo{
					IsValid: false,
				}

				i.PopulateHeaders(rw)
				So(rw.Header(), ShouldResemble, http.Header{
					"X-Authgear-Session-Valid": []string{"false"},
				})
			})

			Convey("valid auth", func() {
				var i = &model.SessionInfo{
					IsValid:         true,
					UserID:          "user-id",
					UserAnonymous:   true,
					UserVerified:    true,
					SessionAMR:      []string{"pwd", "mfa", "otp"},
					AuthenticatedAt: time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC),
				}

				i.PopulateHeaders(rw)
				So(rw.Header(), ShouldResemble, http.Header{
					"X-Authgear-Session-Valid":            []string{"true"},
					"X-Authgear-User-Id":                  []string{"user-id"},
					"X-Authgear-User-Anonymous":           []string{"true"},
					"X-Authgear-User-Verified":            []string{"true"},
					"X-Authgear-Session-Amr":              []string{"pwd mfa otp"},
					"X-Authgear-Session-Authenticated-At": []string{"1136171045"},
				})
			})
		})

		Convey("PopulateHeaders and NewSessionInfoFromHeaders are inverse", func() {
			test := func(info *model.SessionInfo) {
				rw := httptest.NewRecorder()
				info.PopulateHeaders(rw)
				expected, err := model.NewSessionInfoFromHeaders(rw.Header())
				So(err, ShouldBeNil)
				So(expected, ShouldResemble, info)
			}

			test(nil)

			test(&model.SessionInfo{})

			test(&model.SessionInfo{IsValid: false})

			test(&model.SessionInfo{
				IsValid:         true,
				UserID:          "user-id",
				UserAnonymous:   true,
				UserVerified:    true,
				SessionAMR:      []string{"pwd", "mfa", "otp"},
				AuthenticatedAt: time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC),
			})
		})
	})
}
