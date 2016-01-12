package plugin

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

func TestContextMap(t *testing.T) {
	Convey("blank", t, func() {
		ctx := context.Background()
		So(ContextMap(ctx), ShouldResemble, map[string]interface{}{})
	})

	Convey("nil", t, func() {
		So(ContextMap(nil), ShouldResemble, map[string]interface{}{})
	})

	Convey("UserID", t, func() {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "UserID", "42")
		So(ContextMap(ctx), ShouldResemble, map[string]interface{}{
			"user_id": "42",
		})
	})
}
