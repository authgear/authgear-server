package sso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"

	"github.com/skygeario/skygear-server/pkg/core/skydb"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnlinkHandler(t *testing.T) {
	Convey("Test UnlinkHandler", t, func() {
		providerName := "mock"
		providerUserID := "mock_user_id"

		sh := &UnlinkHandler{}
		sh.ProviderName = providerName
		sh.TxContext = db.NewMockTxContext()
		sh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
		mockOAuthProviderKey := oauth.NewMockProviderKey(providerName, providerUserID)
		mockOAuthProvider := oauth.NewMockProvider(
			map[string]string{
				"faseng.cat.id": mockOAuthProviderKey,
			},
			map[string]oauth.Principal{
				mockOAuthProviderKey: oauth.Principal{
					ProviderName:   providerName,
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
				"result": "OK"
			}`)

			p, e := sh.OAuthAuthProvider.GetPrincipalByProviderUserID(providerName, providerUserID)
			So(e, ShouldEqual, skydb.ErrUserNotFound)
			So(p, ShouldBeNil)
		})
	})
}
