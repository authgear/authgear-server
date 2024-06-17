package sessionlisting_test

import (
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
)

func makeDeviceInfo(deviceName string, deviceModel string) map[string]interface{} {
	return map[string]interface{}{
		"ios": map[string]interface{}{
			"UIDevice": map[string]interface{}{
				"name": deviceName,
			},
			"uname": map[string]interface{}{
				"machine": deviceModel,
			},
		},
	}
}

func makeOfflineGrant(id string, lastAccessAt time.Time, deviceInfo map[string]interface{}, idpSessionID string, clientID string, ssoEnabled bool) *oauth.OfflineGrant {
	return &oauth.OfflineGrant{
		ID:           id,
		ClientID:     clientID,
		CreatedAt:    lastAccessAt,
		IDPSessionID: idpSessionID,
		AccessInfo: access.Info{
			InitialAccess: access.Event{
				Timestamp: lastAccessAt,
			},
			LastAccess: access.Event{
				Timestamp: lastAccessAt,
			},
		},
		DeviceInfo: deviceInfo,
		SSOEnabled: ssoEnabled,
	}
}

func makeIDPSession(id string, lastAccessAt time.Time) *idpsession.IDPSession {
	return &idpsession.IDPSession{
		ID:              id,
		CreatedAt:       lastAccessAt,
		AuthenticatedAt: lastAccessAt,
		AccessInfo: access.Info{
			InitialAccess: access.Event{
				Timestamp: lastAccessAt,
			},
			LastAccess: access.Event{
				Timestamp: lastAccessAt,
			},
		},
	}
}

func TestSessionListingService(t *testing.T) {
	Convey("FilterForDisplay", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		any := gomock.Any()

		idpSessions := NewMockIDPSessionProvider(ctrl)
		offlineGrants := NewMockOfflineGrantService(ctrl)

		svc := sessionlisting.SessionListingService{
			IDPSessions:   idpSessions,
			OfflineGrants: offlineGrants,
			OAuthConfig: &config.OAuthConfig{
				Clients: []config.OAuthClientConfig{
					{
						ClientID:        "spa-client-id",
						ApplicationType: config.OAuthClientApplicationTypeSPA,
					}, {
						ClientID:        "third-party-app-client-id",
						ApplicationType: config.OAuthClientApplicationTypeThirdPartyApp,
					},
				},
			},
		}

		Convey("should sort sessions", func() {
			deviceInfo := makeDeviceInfo("myiphone", "iPhone15,2")
			idpSession := makeIDPSession("1", time.Date(2006, 1, 1, 1, 1, 1, 0, time.UTC))
			idpSession2 := makeIDPSession("2", time.Date(2006, 2, 1, 1, 1, 1, 0, time.UTC))
			offlineGrant := makeOfflineGrant("3", time.Date(2006, 3, 1, 1, 1, 1, 0, time.UTC), deviceInfo, idpSession.ID, "spa-client-id", false)
			offlineGrant2 := makeOfflineGrant("4", time.Date(2006, 4, 1, 1, 1, 1, 0, time.UTC), deviceInfo, idpSession.ID, "third-party-app-client-id", false)

			idpSessions.EXPECT().CheckSessionExpired(any).Times(2).Return(false)
			offlineGrants.EXPECT().CheckSessionExpired(any).Times(1).Return(false, time.Time{}, nil)

			session, err := svc.FilterForDisplay([]session.ListableSession{
				offlineGrant2,
				idpSession,
				offlineGrant,
				idpSession2,
			}, nil)
			So(err, ShouldBeNil)
			So(session, ShouldResemble, []*sessionlisting.Session{
				{Session: offlineGrant.ToAPIModel(), IsDevice: true},
				{Session: idpSession2.ToAPIModel()},
				{Session: idpSession.ToAPIModel()},
			})
		})

		Convey("should remove third party client refresh token", func() {
			deviceInfo := makeDeviceInfo("myiphone", "iPhone15,2")
			idpSession := makeIDPSession("1", time.Date(2006, 1, 1, 1, 1, 1, 0, time.UTC))
			idpSession2 := makeIDPSession("2", time.Date(2006, 2, 1, 1, 1, 1, 0, time.UTC))
			offlineGrant := makeOfflineGrant("3", time.Date(2006, 3, 1, 1, 1, 1, 0, time.UTC), deviceInfo, idpSession.ID, "spa-client-id", false)
			offlineGrant2 := makeOfflineGrant("4", time.Date(2006, 4, 1, 1, 1, 1, 0, time.UTC), deviceInfo, idpSession.ID, "spa-client-id", false)

			idpSessions.EXPECT().CheckSessionExpired(any).Times(2).Return(false)
			offlineGrants.EXPECT().CheckSessionExpired(any).Times(2).Return(false, time.Time{}, nil)

			session, err := svc.FilterForDisplay([]session.ListableSession{
				offlineGrant2,
				idpSession,
				offlineGrant,
				idpSession2,
			}, nil)
			So(err, ShouldBeNil)
			So(session, ShouldResemble, []*sessionlisting.Session{
				{Session: offlineGrant2.ToAPIModel(), IsDevice: true},
				{Session: offlineGrant.ToAPIModel(), IsDevice: true},
				{Session: idpSession2.ToAPIModel()},
				{Session: idpSession.ToAPIModel()},
			})
		})

		Convey("should removed expired sessions", func() {
			deviceInfo := makeDeviceInfo("myiphone", "iPhone15,2")
			idpSession := makeIDPSession("1", time.Date(2006, 1, 1, 1, 1, 1, 0, time.UTC))
			idpSession2 := makeIDPSession("2", time.Date(2006, 2, 1, 1, 1, 1, 0, time.UTC))
			offlineGrant := makeOfflineGrant("3", time.Date(2006, 3, 1, 1, 1, 1, 0, time.UTC), deviceInfo, idpSession.ID, "spa-client-id", false)
			offlineGrant2 := makeOfflineGrant("4", time.Date(2006, 4, 1, 1, 1, 1, 0, time.UTC), deviceInfo, idpSession.ID, "spa-client-id", false)

			idpSessions.EXPECT().CheckSessionExpired(idpSession).Return(true)
			idpSessions.EXPECT().CheckSessionExpired(idpSession2).Return(false)
			offlineGrants.EXPECT().CheckSessionExpired(offlineGrant).Return(true, time.Time{}, nil)
			offlineGrants.EXPECT().CheckSessionExpired(offlineGrant2).Return(false, time.Time{}, nil)

			session, err := svc.FilterForDisplay([]session.ListableSession{
				offlineGrant2,
				idpSession,
				offlineGrant,
				idpSession2,
			}, nil)
			So(err, ShouldBeNil)
			So(session, ShouldResemble, []*sessionlisting.Session{
				{Session: offlineGrant2.ToAPIModel(), IsDevice: true},
				{Session: idpSession2.ToAPIModel()},
			})
		})

		Convey("test sso group", func() {
			deviceInfo1 := makeDeviceInfo("myiphone", "iPhone14,2")
			deviceInfo2 := makeDeviceInfo("myiphone", "iPhone15,2")
			idpSession := makeIDPSession("1", time.Date(2006, 5, 1, 1, 1, 1, 0, time.UTC))
			idpSession2 := makeIDPSession("2", time.Date(2006, 2, 1, 1, 1, 1, 0, time.UTC))
			offlineGrant := makeOfflineGrant("3", time.Date(2006, 3, 1, 1, 1, 1, 0, time.UTC), deviceInfo1, idpSession.ID, "spa-client-id", false)
			offlineGrant2 := makeOfflineGrant("4", time.Date(2006, 4, 1, 1, 1, 1, 0, time.UTC), deviceInfo1, idpSession.ID, "spa-client-id", true)
			offlineGrant3 := makeOfflineGrant("5", time.Date(2006, 5, 1, 1, 1, 1, 0, time.UTC), deviceInfo2, idpSession.ID, "spa-client-id", true)

			idpSessions.EXPECT().CheckSessionExpired(any).AnyTimes().Return(false)
			offlineGrants.EXPECT().CheckSessionExpired(any).AnyTimes().Return(false, time.Time{}, nil)

			updatedIDPSessionModel := idpSession.ToAPIModel()
			offlineGrant3SessionModel := offlineGrant3.ToAPIModel()
			// For the same SSO group, idp should use the last accessed offline grant display name
			updatedIDPSessionModel.DisplayName = offlineGrant3SessionModel.DisplayName

			Convey("should show idp sessions only in the same SSO group", func() {
				session, err := svc.FilterForDisplay([]session.ListableSession{
					offlineGrant2,
					idpSession,
					offlineGrant,
					idpSession2,
					offlineGrant3,
				}, nil)
				So(err, ShouldBeNil)
				So(session, ShouldResemble, []*sessionlisting.Session{
					{Session: updatedIDPSessionModel, IsDevice: true},
					{Session: offlineGrant.ToAPIModel(), IsDevice: true},
					{Session: idpSession2.ToAPIModel()},
				})
			})

			Convey("should show session IsCurrent for idp session", func() {
				session, err := svc.FilterForDisplay([]session.ListableSession{
					offlineGrant2,
					idpSession,
					offlineGrant,
					idpSession2,
					offlineGrant3,
				}, idpSession)
				So(err, ShouldBeNil)
				So(session, ShouldResemble, []*sessionlisting.Session{
					{Session: updatedIDPSessionModel, IsDevice: true, IsCurrent: true},
					{Session: offlineGrant.ToAPIModel(), IsDevice: true},
					{Session: idpSession2.ToAPIModel()},
				})
			})

			Convey("should show session IsCurrent for offline grant in same sso group", func() {
				session, err := svc.FilterForDisplay([]session.ListableSession{
					offlineGrant2,
					idpSession,
					offlineGrant,
					idpSession2,
					offlineGrant3,
				}, offlineGrant2)
				So(err, ShouldBeNil)
				So(session, ShouldResemble, []*sessionlisting.Session{
					{Session: updatedIDPSessionModel, IsDevice: true, IsCurrent: true},
					{Session: offlineGrant.ToAPIModel(), IsDevice: true},
					{Session: idpSession2.ToAPIModel()},
				})
			})

			Convey("should show session IsCurrent for sso disabled offline grant", func() {
				session, err := svc.FilterForDisplay([]session.ListableSession{
					offlineGrant2,
					idpSession,
					offlineGrant,
					idpSession2,
					offlineGrant3,
				}, offlineGrant)
				So(err, ShouldBeNil)
				So(session, ShouldResemble, []*sessionlisting.Session{
					{Session: updatedIDPSessionModel, IsDevice: true},
					{Session: offlineGrant.ToAPIModel(), IsDevice: true, IsCurrent: true},
					{Session: idpSession2.ToAPIModel()},
				})
			})

		})

	})
}
