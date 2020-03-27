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
