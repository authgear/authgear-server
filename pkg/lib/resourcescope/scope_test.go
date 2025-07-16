package resourcescope

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateScope(t *testing.T) {
	ctx := context.Background()
	Convey("ValidateScope", t, func() {
		Convey("valid scope-token", func() {
			So(ValidateScope(ctx, "read"), ShouldBeNil)
			So(ValidateScope(ctx, "foo-bar_123:!#[]~"), ShouldBeNil)
		})
		Convey("invalid: empty string", func() {
			So(ValidateScope(ctx, ""), ShouldBeError, `invalid scope:
<root>: format
  map[error:invalid scope scope:]`)
		})
		Convey("invalid: contains space", func() {
			So(ValidateScope(ctx, "read write"), ShouldBeError, `invalid scope:
<root>: format
  map[error:invalid scope scope:read write]`)
		})
		Convey("invalid: forbidden character (\")", func() {
			So(ValidateScope(ctx, "re\"ad"), ShouldBeError, `invalid scope:
<root>: format
  map[error:invalid scope scope:re"ad]`)
		})
		Convey("invalid: forbidden character (\\)", func() {
			So(ValidateScope(ctx, "re\\ad"), ShouldBeError, `invalid scope:
<root>: format
  map[error:invalid scope scope:re\ad]`)
		})
		Convey("invalid: reserved scope", func() {
			So(ValidateScope(ctx, "openid"), ShouldBeError, `invalid scope:
<root>: blocked
  map[reason:ReservedScope scope:openid]`)
		})
	})
}
