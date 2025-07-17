package resourcescope

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateResourceURI(t *testing.T) {
	ctx := context.Background()
	Convey("ValidateResourceURI", t, func() {

		Convey("empty", func() {
			So(ValidateResourceURI(ctx, ""), ShouldBeError, `invalid value:
<root>: minLength
  map[actual:0 expected:1]
<root>: format
  map[error:resource URI must have non-empty host format:x_resource_uri]`)
		})
		Convey("custom scheme should be error", func() {
			So(ValidateResourceURI(ctx, "custom://host"), ShouldBeError, `invalid value:
<root>: format
  map[error:invalid scheme: custom format:x_resource_uri]`)
		})
		Convey("invalid URI should be error", func() {
			So(ValidateResourceURI(ctx, "invalid"), ShouldBeError, `invalid value:
<root>: format
  map[error:resource URI must have non-empty host format:x_resource_uri]`)
		})
		Convey("https scheme should be valid", func() {
			So(ValidateResourceURI(ctx, "https://host"), ShouldBeNil)
		})
		Convey("https scheme with path should be valid", func() {
			So(ValidateResourceURI(ctx, "https://host/path"), ShouldBeNil)
		})
		Convey("query should be rejected", func() {
			So(ValidateResourceURI(ctx, "https://host/path?query=1"), ShouldBeError, `invalid value:
<root>: format
  map[error:resource URI must not have query format:x_resource_uri]`)
		})
		Convey("fragment should be rejected", func() {
			So(ValidateResourceURI(ctx, "https://host/path#fragment"), ShouldBeError, `invalid value:
<root>: format
  map[error:resource URI must not have fragment format:x_resource_uri]`)
		})
		Convey("opaque URI should be rejected", func() {
			So(ValidateResourceURI(ctx, "https:opaque"), ShouldBeError, `invalid value:
<root>: format
  map[error:resource URI must start with https:// format:x_resource_uri]`)
		})
		Convey("userinfo should be rejected", func() {
			So(ValidateResourceURI(ctx, "https://user@host"), ShouldBeError, `invalid value:
<root>: format
  map[error:resource URI must not have user info format:x_resource_uri]`)
		})
		Convey("empty host should be rejected", func() {
			So(ValidateResourceURI(ctx, "https:///path"), ShouldBeError, `invalid value:
<root>: format
  map[error:resource URI must have non-empty host format:x_resource_uri]`)
		})
	})
}
