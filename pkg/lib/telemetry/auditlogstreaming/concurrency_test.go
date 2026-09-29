package auditlogstreaming

import (
	"context"
	"sync/atomic"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestForEachBounded(t *testing.T) {
	Convey("forEachBounded", t, func() {
		ctx := context.Background()

		Convey("runs fn for every item", func() {
			var count atomic.Int32
			forEachBounded(ctx, []int{1, 2, 3, 4, 5}, 2, func(item int) {
				count.Add(int32(item))
			})
			So(count.Load(), ShouldEqual, 15)
		})

		Convey("recovers a panic in fn and still runs the other items", func() {
			var count atomic.Int32
			So(func() {
				forEachBounded(ctx, []int{1, 2, 3}, 2, func(item int) {
					if item == 2 {
						panic("boom")
					}
					count.Add(1)
				})
			}, ShouldNotPanic)
			So(count.Load(), ShouldEqual, 2)
		})
	})
}
