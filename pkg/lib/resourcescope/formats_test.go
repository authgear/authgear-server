package resourcescope

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatResourceURI_CheckFormat(t *testing.T) {
	ctx := context.Background()
	format := FormatResourceURI{}
	Convey("FormatResourceURI.CheckFormat", t, func() {
		Convey("empty", func() {
			So(format.CheckFormat(ctx, ""), ShouldBeError, "resource URI must have non-empty host")
		})
		Convey("custom scheme should be error", func() {
			So(format.CheckFormat(ctx, "custom://host"), ShouldBeError, "invalid scheme: custom")
		})
		Convey("invalid URI should be error", func() {
			So(format.CheckFormat(ctx, "invalid"), ShouldBeError, "resource URI must have non-empty host")
		})
		Convey("https scheme should be valid", func() {
			So(format.CheckFormat(ctx, "https://host"), ShouldBeNil)
		})
		Convey("https scheme with path should be valid", func() {
			So(format.CheckFormat(ctx, "https://host/path"), ShouldBeNil)
		})
		Convey("query should be rejected", func() {
			So(format.CheckFormat(ctx, "https://host/path?query=1"), ShouldBeError, "resource URI must not have query")
		})
		Convey("fragment should be rejected", func() {
			So(format.CheckFormat(ctx, "https://host/path#fragment"), ShouldBeError, "resource URI must not have fragment")
		})
		Convey("opaque URI should be rejected", func() {
			So(format.CheckFormat(ctx, "https:opaque"), ShouldBeError, "resource URI must start with https://")
		})
		Convey("userinfo should be rejected", func() {
			So(format.CheckFormat(ctx, "https://user@host"), ShouldBeError, "resource URI must not have user info")
		})
		Convey("empty host should be rejected", func() {
			So(format.CheckFormat(ctx, "https:///path"), ShouldBeError, "resource URI must have non-empty host")
		})
	})
}

func TestFormatScopeToken_CheckFormat(t *testing.T) {
	ctx := context.Background()
	format := FormatScopeToken{}
	Convey("FormatScopeToken.CheckFormat", t, func() {
		Convey("valid scope-token", func() {
			So(format.CheckFormat(ctx, "read"), ShouldBeNil)
			So(format.CheckFormat(ctx, "foo-bar_123:!#[]~"), ShouldBeNil)
		})
		Convey("invalid: empty string", func() {
			So(format.CheckFormat(ctx, ""), ShouldBeError, "invalid scope-token: forbidden character")
		})
		Convey("invalid: contains space", func() {
			So(format.CheckFormat(ctx, "read write"), ShouldBeError, "invalid scope-token: forbidden character")
		})
		Convey("invalid: forbidden character (\")", func() {
			So(format.CheckFormat(ctx, "re\"ad"), ShouldBeError, "invalid scope-token: forbidden character")
		})
		Convey("invalid: forbidden character (\\)", func() {
			So(format.CheckFormat(ctx, "re\\ad"), ShouldBeError, "invalid scope-token: forbidden character")
		})
	})
}
