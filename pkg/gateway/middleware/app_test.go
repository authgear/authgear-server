package middleware

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

func TestGetDomain(t *testing.T) {
	Convey("getDomain", t, func() {
		store := store.NewMockStore()
		store.Domains = map[string]model.Domain{
			"domain-1": model.Domain{
				ID:         "domain-1",
				Domain:     "example.com",
				Assignment: model.AssignmentTypeMicroservices,
			},
			"domain-2": model.Domain{
				ID:         "domain-2",
				Domain:     "auth.example.com",
				Assignment: model.AssignmentTypeAuth,
			},
			"domain-3": model.Domain{
				ID:         "domain-3",
				Domain:     "app1.cloud.com",
				Assignment: model.AssignmentTypeDefault,
			},
			"domain-4": model.Domain{
				ID:         "domain-4",
				Domain:     "override.app1.cloud.com",
				Assignment: model.AssignmentTypeAsset,
			},
		}
		Convey("should get domain from host", func() {
			cases := []struct {
				host     string
				domainID string
			}{
				{"example.com", "domain-1"},
				{"app.example.com", ""},
				{"auth.example.com", "domain-2"},
				{"app1.cloud.com", "domain-3"},
				{"anything.app1.cloud.com", "domain-3"},
				{"auth.app1.cloud.com", "domain-3"},
				{"override.app1.cloud.com", "domain-4"},
			}

			for _, c := range cases {
				middleware := FindAppMiddleware{
					Store: store,
				}
				domain, _ := middleware.getDomain(c.host)
				if domain != nil {
					So(c.domainID, ShouldEqual, domain.ID)
				} else {
					So(c.domainID, ShouldEqual, "")
				}
			}
		})
	})
}

func TestGetGearName(t *testing.T) {
	Convey("GetGearName", t, func() {
		Convey("should return gear name from path", func() {
			cases := []struct {
				path string
				gear string
			}{
				{"/_auth", "auth"},
				{"/_auth/login/", "auth"},
				{"/auth/", ""},
				{"/_auth////", "auth"},
				{"/_asset/", "asset"},
				{"/index", ""},
				{"", ""},
			}

			for _, c := range cases {
				So(getGearName(c.path), ShouldEqual, c.gear)
			}
		})
	})
}

func TestGetAuthHost(t *testing.T) {
	Convey("getAuthHost", t, func() {
		store := store.NewMockStore()
		store.Domains = map[string]model.Domain{
			"domain-1": model.Domain{
				ID:         "domain-1",
				Domain:     "app1.example.com",
				Assignment: model.AssignmentTypeDefault,
				AppID:      "app1",
			},
			"domain-2": model.Domain{
				ID:         "domain-2",
				Domain:     "app1auth.example.com",
				Assignment: model.AssignmentTypeAuth,
				AppID:      "app1",
			},
			"domain-3": model.Domain{
				ID:         "domain-3",
				Domain:     "app2.example.com",
				Assignment: model.AssignmentTypeDefault,
				AppID:      "app2",
			},
		}
		Convey("should get app auth host", func() {
			cases := []struct {
				appID    string
				domain   *model.Domain
				authHost string
			}{
				// derive from current default domain
				{
					"app1",
					&model.Domain{
						Domain:     "app1.cloud.com",
						Assignment: model.AssignmentTypeDefault,
					},
					"accounts.app1.cloud.com",
				},
				// use current auth domain
				{
					"app1",
					&model.Domain{
						Domain:     "auth.cloud.com",
						Assignment: model.AssignmentTypeAuth,
					},
					"auth.cloud.com",
				},
				// use auth domain if it exists in store
				{
					"app1",
					&model.Domain{
						Assignment: model.AssignmentTypeMicroservices,
					},
					"app1auth.example.com",
				},
				// derive from current default domain if there is no auth domain in store
				{
					"app2",
					&model.Domain{
						Assignment: model.AssignmentTypeAsset,
					},
					"accounts.app2.example.com",
				},
			}

			for _, c := range cases {
				middleware := FindAppMiddleware{
					Store: store,
				}
				authHost, err := middleware.getAuthHost(c.appID, c.domain)

				So(err, ShouldBeNil)
				So(c.authHost, ShouldEqual, authHost)
			}
		})
	})
}
