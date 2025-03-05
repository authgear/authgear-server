package log

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestContextCauseHook(t *testing.T) {
	Convey("ContextCauseHook", t, func() {
		test := func(entry *logrus.Entry, expected string) {
			h := &ContextCauseHook{}
			err := h.Fire(entry)
			So(err, ShouldBeNil)

			val, ok := entry.Data["context_cause"].(string)
			if expected != "" {
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, expected)
			} else {
				So(ok, ShouldBeFalse)
			}
		}

		Convey("no context", func() {
			test(&logrus.Entry{Data: make(logrus.Fields)}, "<context-is-nil>")
		})

		Convey("has context, but it is not canceled yet", func() {
			ctx := context.Background()
			test(&logrus.Entry{Context: ctx, Data: make(logrus.Fields)}, "<context-err-is-nil>")
		})

		Convey("has context, canceled without cause", func() {
			ctx := context.Background()

			ctx, cancel := context.WithCancel(ctx)
			cancel()

			// This test observes as a documentation of the stdlib behavior.
			// When it is canceled without cause, the cause is context.Canceled itself.
			test(&logrus.Entry{Context: ctx, Data: make(logrus.Fields)}, "context canceled")
		})

		Convey("has context, canceled with cause", func() {
			ctx := context.Background()

			ctx, cancel := context.WithCancelCause(ctx)
			cancel(fmt.Errorf("the cause"))

			test(&logrus.Entry{Context: ctx, Data: make(logrus.Fields)}, "the cause")
		})

		// Cannot test WithDeadline and and WithTimeout because we have no access to the clock
		// used by the package context.
	})
}
