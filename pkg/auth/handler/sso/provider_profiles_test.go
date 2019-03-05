package sso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProviderProfilesHandler(t *testing.T) {
	realTime := timeNow
	timeNow = func() time.Time { return zeroTime }
	defer func() {
		timeNow = realTime
	}()

	Convey("Test ProviderProfilesHandler", t, func() {
		ph := &ProviderProfilesHandler{}
		ph.TxContext = db.NewMockTxContext()
		mockOAuthProvider := oauth.NewMockProviderWithPrincipals(
			[]*oauth.Principal{
				&oauth.Principal{
					ID:             "1",
					UserID:         "faseng.cat.id",
					ProviderName:   "provider1",
					ProviderUserID: "provider1.faseng",
					UserProfile: map[string]interface{}{
						"username": "faseng",
						"email":    "faseng@provider1",
						"id":       "provider1.faseng",
					},
				},
				&oauth.Principal{
					ID:             "2",
					UserID:         "faseng.cat.id",
					ProviderName:   "provider2",
					ProviderUserID: "provider2.faseng",
					UserProfile: map[string]interface{}{
						"username": "faseng",
						"email":    "faseng@provider2",
					},
				},
				&oauth.Principal{
					ID:             "3",
					UserID:         "chima.cat.id",
					ProviderName:   "provider1",
					ProviderUserID: "provider1.chima",
					UserProfile: map[string]interface{}{
						"username": "chima",
						"email":    "chima@provider1",
						"id":       "provider1.chima",
					},
				},
			},
		)
		ph.OAuthAuthProvider = mockOAuthProvider
		h := handler.APIHandlerToHandler(ph, ph.TxContext)

		Convey("should return multiple profiles if user connected to multiple providers", func() {
			ph.AuthContext = auth.NewMockContextGetterWithUser(
				"faseng.cat.id", true, map[string]bool{},
			)

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"provider1": {
						"email": "faseng@provider1",
						"id": "provider1.faseng",
						"username": "faseng"
					},
					"provider2": {
						"email": "faseng@provider2",
						"username": "faseng"
					}
				}
			}`)
		})

		Convey("should return empty profiles if user has not connected to provider", func() {
			ph.AuthContext = auth.NewMockContextGetterWithUser(
				"milktea.cat.id", true, map[string]bool{},
			)
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result": {}}`)
		})

	})
}
