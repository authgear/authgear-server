package sso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"

	"github.com/skygeario/skygear-server/pkg/core/skydb"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnlinkHandler(t *testing.T) {
	Convey("Test UnlinkHandler", t, func() {
		providerID := "mock"
		providerUserID := "mock_user_id"

		sh := &UnlinkHandler{}
		sh.ProviderID = providerID
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		mockOAuthProviderKey := oauth.NewMockProviderKey(providerID, providerUserID)
		mockOAuthProvider := oauth.NewMockProvider(
			map[string]string{
				"faseng.cat.id": mockOAuthProviderKey,
			},
			map[string]oauth.Principal{
				mockOAuthProviderKey: oauth.Principal{
					ProviderName:   providerID,
					UserID:         "faseng.cat.id",
					ProviderUserID: providerUserID,
				},
			},
		)
		sh.OAuthAuthProvider = mockOAuthProvider
		h := handler.APIHandlerToHandler(sh, sh.TxContext)

		Convey("should unlink user id with oauth principal", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"access_token": "token"
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {}
			}`)

			p, e := sh.OAuthAuthProvider.GetPrincipalByProviderUserID(providerID, providerUserID)
			So(e, ShouldEqual, skydb.ErrUserNotFound)
			So(p, ShouldBeNil)
		})
	})
}
