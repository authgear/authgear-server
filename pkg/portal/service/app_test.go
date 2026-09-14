package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	runtimeresource "github.com/authgear/authgear-server"
	"github.com/authgear/authgear-server/pkg/portal/model"
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
		})

		Convey("app ID cannot start with 2 lowercase characters, followed by a hyphen", func() {
			So(service.validateAppID(ctx, "ab-d"), ShouldBeError, ErrAppIDReserved)
			So(service.validateAppID(ctx, "us-east-1"), ShouldBeError, ErrAppIDReserved)
		})

		Convey("siteadmin and site-admin are reserved", func() {
			So(service.validateAppID(ctx, "siteadmin"), ShouldBeError, ErrAppIDReserved)
			So(service.validateAppID(ctx, "site-admin"), ShouldBeError, ErrAppIDReserved)
		})

		Convey("some examples of valid app ID", func() {
			So(service.validateAppID(ctx, "myapp"), ShouldBeNil)
			So(service.validateAppID(ctx, "this-app"), ShouldBeNil)
		})
	})
}

func TestSortAppListItems(t *testing.T) {
	Convey("sortAppListItems", t, func() {
		appIDs := func(items []*model.AppListItem) []string {
			out := make([]string, len(items))
			for i, item := range items {
				out[i] = item.AppID
			}
			return out
		}

		Convey("orders items by app ID in ascending byte order", func() {
			items := []*model.AppListItem{
				{AppID: "zebra-4d3c2b", PublicOrigin: "https://zebra-4d3c2b.example.com"},
				{AppID: "apple-10", PublicOrigin: "https://apple-10.example.com"},
				{AppID: "mango-9f8e7d", PublicOrigin: "https://mango-9f8e7d.example.com"},
				{AppID: "apple-2", PublicOrigin: "https://apple-2.example.com"},
			}

			sortAppListItems(items)

			So(appIDs(items), ShouldResemble, []string{
				"apple-10",
				"apple-2",
				"mango-9f8e7d",
				"zebra-4d3c2b",
			})
			// Each item keeps its own origin; only the order changes.
			So(items[0].PublicOrigin, ShouldEqual, "https://apple-10.example.com")
		})

		Convey("leaves an already sorted list unchanged", func() {
			items := []*model.AppListItem{
				{AppID: "a-project"},
				{AppID: "b-project"},
			}

			sortAppListItems(items)

			So(appIDs(items), ShouldResemble, []string{"a-project", "b-project"})
		})

		Convey("accepts an empty or nil list", func() {
			empty := []*model.AppListItem{}
			sortAppListItems(empty)
			So(empty, ShouldBeEmpty)

			So(func() { sortAppListItems(nil) }, ShouldNotPanic)
		})
	})
}
