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
					GrantID:   "grant-id",
					Token:     "new-refresh-token",
					TokenHash: "new-refresh-token-hash",
				}, &oauth.OfflineGrant{
					ID: "grant-id",
				}, nil)
				accessGrants.EXPECT().PrepareUserAccessGrant(gomock.Any(), gomock.Any()).Return(nil, nil)

				result1, err := s.IssueAccessGrantByRefreshToken(context.Background(), handler.IssueAccessGrantByRefreshTokenOptions{
					ShouldRotateRefreshToken: true,
					PrepareUserAccessGrantOptions: oauth.PrepareUserAccessGrantOptions{
						SessionLike: &oauth.OfflineGrant{
							ID: "grant-id",
						},
						InitialRefreshTokenHash: "initial-refresh-token-hash",
					},
				})

				resp := protocol.TokenResponse{}
				So(err, ShouldBeNil)
				So(result1.RotateRefreshTokenResult, ShouldNotBeNil)
				So(result1.RotateRefreshTokenResult.TokenHash, ShouldEqual, "new-refresh-token-hash")
				So(result1.RotateRefreshTokenResult.GrantID, ShouldEqual, "grant-id")

				result1.RotateRefreshTokenResult.WriteTo(resp)
				So(resp["refresh_token"], ShouldEqual, "grant-id.new-refresh-token")
			})

			Convey("should not rotate refresh token", func() {
				accessGrants.EXPECT().PrepareUserAccessGrant(gomock.Any(), gomock.Any()).Return(nil, nil)

				resp := protocol.TokenResponse{}
				result1, err := s.IssueAccessGrantByRefreshToken(context.Background(), handler.IssueAccessGrantByRefreshTokenOptions{
					ShouldRotateRefreshToken: false,
					PrepareUserAccessGrantOptions: oauth.PrepareUserAccessGrantOptions{
						SessionLike: &oauth.OfflineGrant{
							ID: "grant-id",
						},
					},
				})

				So(err, ShouldBeNil)
				So(result1.RotateRefreshTokenResult, ShouldBeNil)

				result1.RotateRefreshTokenResult.WriteTo(resp)
				So(resp, ShouldNotContainKey, "refresh_token")
			})
		})
	})
}
