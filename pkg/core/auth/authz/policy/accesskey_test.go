package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRequireClient(t *testing.T) {
	Convey("Test RequireClient", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)

		Convey("should return error if auth context has no access key", func() {
			err := RequireClient(req)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should not return error if auth context has api key ", func() {
			req = req.WithContext(auth.WithAccessKey(req.Context(), auth.AccessKey{
				Client: config.OAuthClientConfiguration{},
			}))
			err := RequireClient(req)
			So(err, ShouldBeEmpty)
		})

		Convey("should return error if auth context has master key ", func() {
			req = req.WithContext(auth.WithAccessKey(req.Context(), auth.AccessKey{
				IsMasterKey: true,
			}))
			err := RequireClient(req)
			So(err, ShouldNotBeEmpty)
		})
	})
}

func TestRequireMasterKey(t *testing.T) {
	Convey("Test RequireMasterKey", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)

		Convey("should return error if auth context has no access key", func() {
			err := RequireMasterKey(req)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should return error if auth context has api key ", func() {
			req = req.WithContext(auth.WithAccessKey(req.Context(), auth.AccessKey{
				Client: config.OAuthClientConfiguration{},
			}))
			err := RequireMasterKey(req)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should not return error if auth context has master key ", func() {
			req = req.WithContext(auth.WithAccessKey(req.Context(), auth.AccessKey{
				IsMasterKey: true,
			}))
			err := RequireMasterKey(req)
			So(err, ShouldBeEmpty)
		})
	})
}
