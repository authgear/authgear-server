package otelutil

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.opentelemetry.io/otel/baggage"
)

func TestBaggage(t *testing.T) {
	Convey("GetAuthgearBaggage", t, func() {
		ctx := context.Background()

		Convey("should return nil if no baggage", func() {
			m := GetAuthgearBaggage(ctx)
			So(m, ShouldBeNil)
		})

		Convey("should return keys from baggage", func() {
			m1, _ := baggage.NewMember("authgear_sdk_user_id", "user1")
			m2, _ := baggage.NewMember("authgear_sdk_device_id", "device1")
			b, _ := baggage.New(m1, m2)
			ctx := baggage.ContextWithBaggage(ctx, b)

			m := GetAuthgearBaggage(ctx)
			So(m, ShouldResemble, map[string]string{
				"authgear_sdk_user_id":   "user1",
				"authgear_sdk_device_id": "device1",
			})
		})

		Convey("should return partial keys", func() {
			m1, _ := baggage.NewMember("authgear_sdk_user_id", "user1")
			b, _ := baggage.New(m1)
			ctx := baggage.ContextWithBaggage(ctx, b)

			m := GetAuthgearBaggage(ctx)
			So(m, ShouldResemble, map[string]string{
				"authgear_sdk_user_id": "user1",
			})
		})

		Convey("should ignore keys if value is too long", func() {
			val := ""
			for i := 0; i < 600; i++ {
				val += "a"
			}

			m1, _ := baggage.NewMember("authgear_sdk_user_id", val)
			b, _ := baggage.New(m1)
			ctx := baggage.ContextWithBaggage(ctx, b)

			m := GetAuthgearBaggage(ctx)
			So(m, ShouldBeNil)
		})
	})
}
