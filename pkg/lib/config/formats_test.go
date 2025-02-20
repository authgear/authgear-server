package config

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func backgroundCtx() context.Context {
	return context.Background()
}

func TestFormatPhone(t *testing.T) {
	f := FormatPhone{}.CheckFormat

	Convey("FormatPhone", t, func() {
		So(f(backgroundCtx(), 1), ShouldBeNil)
		So(f(backgroundCtx(), "+85298765432"), ShouldBeNil)
		So(f(backgroundCtx(), ""), ShouldBeError, "not in E.164 format")
		So(f(backgroundCtx(), "foobar"), ShouldBeError, "not in E.164 format")
	})
}
