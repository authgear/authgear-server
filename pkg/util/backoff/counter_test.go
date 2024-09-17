package backoff_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/backoff"
)

func TestCounter(t *testing.T) {
	Convey("Counter", t, func() {
		maxInterval := time.Second * 10
		counter := backoff.Counter{Interval: time.Second, MaxInterval: maxInterval}

		So(counter.BackoffDuration(), ShouldEqual, time.Duration(0))

		counter.Increment()
		So(counter.BackoffDuration(), ShouldBeBetweenOrEqual, time.Second*1, time.Second*2)

		counter.Increment()
		So(counter.BackoffDuration(), ShouldBeBetweenOrEqual, time.Second*2, time.Second*4)

		counter.Increment()
		So(counter.BackoffDuration(), ShouldBeBetweenOrEqual, time.Second*4, time.Second*5)

		counter.Increment()
		So(counter.BackoffDuration(), ShouldBeBetweenOrEqual, time.Second*8, time.Second*9)

		counter.Increment()
		So(counter.BackoffDuration(), ShouldEqual, time.Second*10)

		counter.Increment()
		So(counter.BackoffDuration(), ShouldEqual, time.Second*10)

		counter.Reset()
		So(counter.BackoffDuration(), ShouldEqual, time.Duration(0))
	})
}
