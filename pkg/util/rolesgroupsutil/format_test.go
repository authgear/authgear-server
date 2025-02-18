package rolesgroupsutil

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFormatKey(t *testing.T) {
	Convey("FormatKey", t, func() {
		ctx := context.Background()
		f := FormatKey{}.CheckFormat
		So(f(ctx, nil), ShouldBeNil)
		So(f(ctx, 1), ShouldBeNil)
		So(f(ctx, ""), ShouldBeNil)
		So(f(ctx, "authgear:"), ShouldBeError, "key cannot start with the preserved prefix: `authgear:`")
	})
}
