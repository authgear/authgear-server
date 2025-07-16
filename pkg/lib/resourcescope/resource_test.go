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
  map[error:invalid scheme:  format:x_resource_uri]`)
		})
		Convey("custom scheme should be error", func() {
			So(ValidateResourceURI(ctx, "custom://host"), ShouldBeError, `invalid value:
<root>: format
  map[error:invalid scheme: custom format:x_resource_uri]`)
		})
		Convey("invalid URI should be error", func() {
			So(ValidateResourceURI(ctx, "invalid"), ShouldBeError, `invalid value:
<root>: format
  map[error:invalid scheme:  format:x_resource_uri]`)
		})
		Convey("https scheme should be valid", func() {
			So(ValidateResourceURI(ctx, "https://host"), ShouldBeNil)
		})
		Convey("https scheme with path should be valid", func() {
			So(ValidateResourceURI(ctx, "https://host/path"), ShouldBeNil)
		})
	})
}
