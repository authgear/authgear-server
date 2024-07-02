package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/clock"
	gomock "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTokenHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	appID := config.AppID("app-id")
	clock := clock.NewMockClockAt("2020-02-01T00:00:00Z")
	tokenService := NewMockTokenHandlerTokenService(ctrl)
	offlineGrantService := NewMockTokenHandlerOfflineGrantService(ctrl)
	offlineGrants := NewMockTokenHandlerOfflineGrantStore(ctrl)
	idTokenIssuer := NewMockIDTokenIssuer(ctrl)
	clientResolver := &mockClientResolver{ClientConfig: &config.OAuthClientConfig{
		ClientID:   "app-id",
		GrantTypes: []string{"authorization_code", "refresh_token"},
	}}
	h := &handler.TokenHandler{
		Context:             context.Background(),
		AppID:               appID,
		Config:              &config.OAuthConfig{},
		HTTPOrigin:          "http://accounts.example.com",
		TokenService:        tokenService,
		Clock:               clock,
		RemoteIP:            "1.2.3.4",
		UserAgentString:     "UA",
		OfflineGrantService: offlineGrantService,
		OfflineGrants:       offlineGrants,
		IDTokenIssuer:       idTokenIssuer,
		ClientResolver:      clientResolver,
	}
	handle := func(req *http.Request, r protocol.TokenRequest) *httptest.ResponseRecorder {
		rw := httptest.NewRecorder()
		result := h.Handle(rw, req, r)
		result.WriteResponse(rw, req)
		return rw
	}

	Convey("handle refresh token", t, func() {
		Convey("success", func() {
			req, _ := http.NewRequest("POST", "/token", nil)
			r := protocol.TokenRequest{}
			r["grant_type"] = "refresh_token"
			r["client_id"] = "app-id"
			r["refresh_token"] = "asdf"
			offlineGrant := &oauth.OfflineGrant{
				ID:       "offline-grant-id",
				ClientID: "app-id",
				Scopes:   []string{"openid"},
			}
			tokenService.EXPECT().ParseRefreshToken("asdf").Return(&oauth.Authorization{}, offlineGrant, nil)
			idTokenIssuer.EXPECT().IssueIDToken(gomock.Any()).Return("id-token", nil)
			tokenService.EXPECT().IssueAccessGrant(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			expireAt := time.Date(2020, 02, 01, 1, 0, 0, 0, time.UTC)
			offlineGrantService.EXPECT().ComputeOfflineGrantExpiry(offlineGrant).Return(expireAt, nil)
			offlineGrants.EXPECT().AccessOfflineGrantAndUpdateDeviceInfo("offline-grant-id", access.NewEvent(clock.NowUTC(), "1.2.3.4", "UA"), gomock.Any(), expireAt).Return(offlineGrant, nil)
			res := handle(req, r)
			So(res.Result().StatusCode, ShouldEqual, 200)
		})
	})
}
