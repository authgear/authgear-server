package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	runtimeresource "github.com/authgear/authgear-server"
	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestValidateAppID(t *testing.T) {
	resourceManager := resource.NewManagerWithDir(resource.NewManagerWithDirOptions{
		Registry:              portalresource.PortalRegistry,
		BuiltinResourceFS:     runtimeresource.EmbedFS_resources_portal,
		BuiltinResourceFSRoot: runtimeresource.RelativePath_resources_portal,
	})

	service := AppService{
		Resources: resourceManager,
	}

	ctx := context.Background()
	Convey("validateAppID", t, func() {
		Convey("empty app ID is invalid", func() {
			So(service.validateAppID(ctx, ""), ShouldBeError, ErrAppIDInvalid)
		})

		Convey("app ID whose len < 4 is invalid", func() {
			So(service.validateAppID(ctx, "a"), ShouldBeError, ErrAppIDInvalid)
			So(service.validateAppID(ctx, "ab"), ShouldBeError, ErrAppIDInvalid)
			So(service.validateAppID(ctx, "abc"), ShouldBeError, ErrAppIDInvalid)
		})

		Convey("app ID whose len > 32 is invalid", func() {
			So(service.validateAppID(ctx, "01234567890123456789012345678901"), ShouldBeNil)
			So(service.validateAppID(ctx, "012345678901234567890123456789012"), ShouldBeError, ErrAppIDInvalid)
		})

		Convey("app ID can only start with a-z 0-9", func() {
			So(service.validateAppID(ctx, "Abcd"), ShouldBeError, ErrAppIDInvalid)
			So(service.validateAppID(ctx, "-bcd"), ShouldBeError, ErrAppIDInvalid)

			So(service.validateAppID(ctx, "abcd"), ShouldBeNil)
			So(service.validateAppID(ctx, "0bcd"), ShouldBeNil)
		})

		Convey("app ID can only end with a-z 0-9", func() {
			So(service.validateAppID(ctx, "abcD"), ShouldBeError, ErrAppIDInvalid)
			So(service.validateAppID(ctx, "abc-"), ShouldBeError, ErrAppIDInvalid)

			So(service.validateAppID(ctx, "abcd"), ShouldBeNil)
			So(service.validateAppID(ctx, "abc0"), ShouldBeNil)
		})

		Convey("app ID can contain hyphen", func() {
			So(service.validateAppID(ctx, "a-cd"), ShouldBeNil)
			So(service.validateAppID(ctx, "ab-d"), ShouldBeNil)
		})
	})
}
