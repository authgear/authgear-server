package handler_test

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestTokenService(t *testing.T) {
	Convey("TokenService", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMockClockAt("2020-02-01T00:00:00Z")
		offlineGrants := NewMockTokenServiceOfflineGrantService(ctrl)
		accessGrants := NewMockTokenServiceAccessGrantService(ctrl)

		s := &handler.TokenService{
			Clock:               clock,
			OfflineGrantService: offlineGrants,
			AccessGrantService:  accessGrants,
		}

		Convey("IssueAccessGrantByRefreshToken", func() {
			Convey("should rotate refresh token", func() {
				offlineGrants.EXPECT().GetOfflineGrant(gomock.Any(), "grant-id").Return(&oauth.OfflineGrant{
					ID: "grant-id",
				}, nil)
				offlineGrants.EXPECT().RotateRefreshToken(gomock.Any(), gomock.Any()).Return(&oauth.RotateRefreshTokenResult{
					Token:     "new-refresh-token",
					TokenHash: "new-refresh-token-hash",
				}, &oauth.OfflineGrant{
					ID: "grant-id",
				}, nil)
				accessGrants.EXPECT().IssueAccessGrant(gomock.Any(), gomock.Any()).Return(&oauth.IssueAccessGrantResult{}, nil)

				resp := protocol.TokenResponse{}
				err := s.IssueAccessGrantByRefreshToken(context.Background(), handler.IssueAccessGrantByRefreshTokenOptions{
					ShouldRotateRefreshToken: true,
					IssueAccessGrantOptions: oauth.IssueAccessGrantOptions{
						SessionLike: &oauth.OfflineGrant{
							ID: "grant-id",
						},
						InitialRefreshTokenHash: "refresh-token-hash",
					},
				}, resp)

				So(err, ShouldBeNil)
				So(resp, ShouldContainKey, "refresh_token")
				So(resp["refresh_token"], ShouldEqual, "grant-id.new-refresh-token")
			})

			Convey("should not rotate refresh token", func() {
				accessGrants.EXPECT().IssueAccessGrant(gomock.Any(), gomock.Any()).Return(&oauth.IssueAccessGrantResult{}, nil)

				resp := protocol.TokenResponse{}
				err := s.IssueAccessGrantByRefreshToken(context.Background(), handler.IssueAccessGrantByRefreshTokenOptions{
					ShouldRotateRefreshToken: false,
					IssueAccessGrantOptions: oauth.IssueAccessGrantOptions{
						SessionLike: &oauth.OfflineGrant{
							ID: "grant-id",
						},
					},
				}, resp)

				So(err, ShouldBeNil)
				So(resp, ShouldNotContainKey, "refresh_token")
			})
		})
	})
}
