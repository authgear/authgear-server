package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDenyNoAccessKey(t *testing.T) {
	Convey("Test DenyNoAccessKey", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := MemoryContextGetter{}

		Convey("should return error if auth context has no access key", func() {
			err := DenyNoAccessKey(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should not return error if auth context has api key ", func() {
			req = req.WithContext(auth.WithAccessKey(req.Context(), auth.AccessKey{
				Client: config.OAuthClientConfiguration{},
			}))
			err := DenyNoAccessKey(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should return error if auth context has master key ", func() {
			req = req.WithContext(auth.WithAccessKey(req.Context(), auth.AccessKey{
				IsMasterKey: true,
			}))
			err := DenyNoAccessKey(req, ctx)
			So(err, ShouldNotBeEmpty)
		})
	})
}

func TestRequireMasterKey(t *testing.T) {
	Convey("Test RequireMasterKey", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := MemoryContextGetter{}

		Convey("should return error if auth context has no access key", func() {
			err := RequireMasterKey(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should return error if auth context has api key ", func() {
			req = req.WithContext(auth.WithAccessKey(req.Context(), auth.AccessKey{
				Client: config.OAuthClientConfiguration{},
			}))
			err := RequireMasterKey(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should not return error if auth context has master key ", func() {
			req = req.WithContext(auth.WithAccessKey(req.Context(), auth.AccessKey{
				IsMasterKey: true,
			}))
			err := RequireMasterKey(req, ctx)
			So(err, ShouldBeEmpty)
		})
	})
}
