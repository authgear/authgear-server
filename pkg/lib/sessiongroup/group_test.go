package sessiongroup

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
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

func makeOfflineGrant(id string, lastAccessAt time.Time, deviceInfo map[string]interface{}, idpSessionID string) *oauth.OfflineGrant {
	return &oauth.OfflineGrant{
		ID:           id,
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

func TestGroup(t *testing.T) {
	timeA := time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC)
	timeB := time.Date(2006, 1, 2, 3, 4, 5, 1, time.UTC)

	Convey("Group", t, func() {
		Convey("empty list", func() {
			actual := Group(nil)
			So(actual, ShouldHaveLength, 0)
		})

		Convey("single IDP session", func() {
			idpSession := makeIDPSession("1", timeA)
			actual := Group([]session.Session{idpSession})
			So(actual, ShouldResemble, []*model.SessionGroup{
				&model.SessionGroup{
					Type:           model.SessionGroupTypeUngrouped,
					LastAccessedAt: timeA,
					Sessions: []*model.Session{
						&model.Session{
							Meta: model.Meta{
								ID:        "1",
								CreatedAt: timeA,
								UpdatedAt: timeA,
							},
							Type:           model.SessionTypeIDP,
							LastAccessedAt: timeA,
						},
					},
				},
			})
		})

		Convey("single offline grant", func() {
			deviceInfo := makeDeviceInfo("myiphone", "iPhone13,1")
			offlineGrant := makeOfflineGrant("1", timeA, deviceInfo, "")
			actual := Group([]session.Session{offlineGrant})
			So(actual, ShouldResemble, []*model.SessionGroup{
				&model.SessionGroup{
					Type:            model.SessionGroupTypeGrouped,
					LastAccessedAt:  timeA,
					DisplayName:     "iPhone 12 mini",
					OfflineGrantIDs: []string{"1"},
					Sessions: []*model.Session{
						&model.Session{
							Meta: model.Meta{
								ID:        "1",
								CreatedAt: timeA,
								UpdatedAt: timeA,
							},
							Type:           model.SessionTypeOfflineGrant,
							LastAccessedAt: timeA,
							DisplayName:    "iPhone 12 mini",
						},
					},
				},
			})
		})

		Convey("two offline grants", func() {
			deviceInfo := makeDeviceInfo("myiphone", "iPhone13,1")
			offlineGrant1 := makeOfflineGrant("1", timeA, deviceInfo, "")
			offlineGrant2 := makeOfflineGrant("2", timeA, deviceInfo, "")
			actual := Group([]session.Session{offlineGrant1, offlineGrant2})
			So(actual, ShouldResemble, []*model.SessionGroup{
				&model.SessionGroup{
					Type:            model.SessionGroupTypeGrouped,
					LastAccessedAt:  timeA,
					DisplayName:     "iPhone 12 mini",
					OfflineGrantIDs: []string{"1", "2"},
					Sessions: []*model.Session{
						&model.Session{
							Meta: model.Meta{
								ID:        "1",
								CreatedAt: timeA,
								UpdatedAt: timeA,
							},
							Type:           model.SessionTypeOfflineGrant,
							LastAccessedAt: timeA,
							DisplayName:    "iPhone 12 mini",
						},
						&model.Session{
							Meta: model.Meta{
								ID:        "2",
								CreatedAt: timeA,
								UpdatedAt: timeA,
							},
							Type:           model.SessionTypeOfflineGrant,
							LastAccessedAt: timeA,
							DisplayName:    "iPhone 12 mini",
						},
					},
				},
			})
		})

		Convey("two offline grants on different devices", func() {
			deviceInfo1 := makeDeviceInfo("myiphone", "iPhone13,1")
			deviceInfo2 := makeDeviceInfo("theiriphone", "iPhone13,2")
			offlineGrant1 := makeOfflineGrant("1", timeA, deviceInfo1, "")
			offlineGrant2 := makeOfflineGrant("2", timeB, deviceInfo2, "")
			actual := Group([]session.Session{offlineGrant1, offlineGrant2})
			So(actual, ShouldResemble, []*model.SessionGroup{
				&model.SessionGroup{
					Type:            model.SessionGroupTypeGrouped,
					LastAccessedAt:  timeB,
					DisplayName:     "iPhone 12",
					OfflineGrantIDs: []string{"2"},
					Sessions: []*model.Session{
						&model.Session{
							Meta: model.Meta{
								ID:        "2",
								CreatedAt: timeB,
								UpdatedAt: timeB,
							},
							Type:           model.SessionTypeOfflineGrant,
							LastAccessedAt: timeB,
							DisplayName:    "iPhone 12",
						},
					},
				},
				&model.SessionGroup{
					Type:            model.SessionGroupTypeGrouped,
					LastAccessedAt:  timeA,
					DisplayName:     "iPhone 12 mini",
					OfflineGrantIDs: []string{"1"},
					Sessions: []*model.Session{
						&model.Session{
							Meta: model.Meta{
								ID:        "1",
								CreatedAt: timeA,
								UpdatedAt: timeA,
							},
							Type:           model.SessionTypeOfflineGrant,
							LastAccessedAt: timeA,
							DisplayName:    "iPhone 12 mini",
						},
					},
				},
			})
		})

		Convey("offline grant and IDP session", func() {
			idpSession := makeIDPSession("1", timeA)
			deviceInfo := makeDeviceInfo("myiphone", "iPhone13,1")
			offlineGrant := makeOfflineGrant("2", timeB, deviceInfo, "1")
			actual := Group([]session.Session{offlineGrant, idpSession})
			So(actual, ShouldResemble, []*model.SessionGroup{
				&model.SessionGroup{
					Type:            model.SessionGroupTypeGrouped,
					LastAccessedAt:  timeB,
					DisplayName:     "iPhone 12 mini",
					OfflineGrantIDs: []string{"2"},
					Sessions: []*model.Session{
						&model.Session{
							Meta: model.Meta{
								ID:        "2",
								CreatedAt: timeB,
								UpdatedAt: timeB,
							},
							Type:           model.SessionTypeOfflineGrant,
							LastAccessedAt: timeB,
							DisplayName:    "iPhone 12 mini",
						},
						&model.Session{
							Meta: model.Meta{
								ID:        "1",
								CreatedAt: timeA,
								UpdatedAt: timeA,
							},
							Type:           model.SessionTypeIDP,
							LastAccessedAt: timeA,
						},
					},
				},
			})
		})

		Convey("ungrouped IDP session", func() {
			idpSession1 := makeIDPSession("1", timeA)
			idpSession2 := makeIDPSession("2", timeA)
			deviceInfo := makeDeviceInfo("myiphone", "iPhone13,1")
			offlineGrant := makeOfflineGrant("3", timeB, deviceInfo, "1")
			actual := Group([]session.Session{offlineGrant, idpSession1, idpSession2})
			So(actual, ShouldResemble, []*model.SessionGroup{
				&model.SessionGroup{
					Type:            model.SessionGroupTypeGrouped,
					LastAccessedAt:  timeB,
					DisplayName:     "iPhone 12 mini",
					OfflineGrantIDs: []string{"3"},
					Sessions: []*model.Session{
						&model.Session{
							Meta: model.Meta{
								ID:        "3",
								CreatedAt: timeB,
								UpdatedAt: timeB,
							},
							Type:           model.SessionTypeOfflineGrant,
							LastAccessedAt: timeB,
							DisplayName:    "iPhone 12 mini",
						},
						&model.Session{
							Meta: model.Meta{
								ID:        "1",
								CreatedAt: timeA,
								UpdatedAt: timeA,
							},
							Type:           model.SessionTypeIDP,
							LastAccessedAt: timeA,
						},
					},
				},
				&model.SessionGroup{
					Type:           model.SessionGroupTypeUngrouped,
					LastAccessedAt: timeA,
					Sessions: []*model.Session{
						&model.Session{
							Meta: model.Meta{
								ID:        "2",
								CreatedAt: timeA,
								UpdatedAt: timeA,
							},
							Type:           model.SessionTypeIDP,
							LastAccessedAt: timeA,
						},
					},
				},
			})
		})
	})
}
