package db

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWithTx(t *testing.T) {
	Convey("WithTx", t, func() {
		c := &MockContext{}

		Convey("commit", func() {
			f := func() error { return nil }
			err := WithTx(c, f)
			So(err, ShouldBeNil)
			So(c.DidBegin, ShouldBeTrue)
			So(c.DidCommit, ShouldBeTrue)
			So(c.DidRollback, ShouldBeFalse)
		})

		Convey("rollback on error", func() {
			fErr := errors.New("error")
			f := func() error { return fErr }
			err := WithTx(c, f)
			So(err, ShouldEqual, fErr)
			So(c.DidBegin, ShouldBeTrue)
			So(c.DidCommit, ShouldBeFalse)
			So(c.DidRollback, ShouldBeTrue)
		})

		Convey("rollback on panic", func() {
			f := func() error { panic(errors.New("panic")) }
			defer func() {
				_ = recover()
				So(c.DidBegin, ShouldBeTrue)
				So(c.DidCommit, ShouldBeFalse)
				So(c.DidRollback, ShouldBeTrue)
			}()
			_ = WithTx(c, f)
		})
	})
}

func TestReadOnly(t *testing.T) {
	Convey("ReadOnly", t, func() {
		c := &MockContext{}

		Convey("rollback when no error", func() {
			f := func() error { return nil }
			err := ReadOnly(c, f)
			So(err, ShouldBeNil)
			So(c.DidBegin, ShouldBeTrue)
			So(c.DidCommit, ShouldBeFalse)
			So(c.DidRollback, ShouldBeTrue)
		})

		Convey("rollback on error", func() {
			fErr := errors.New("error")
			f := func() error { return fErr }
			err := ReadOnly(c, f)
			So(err, ShouldEqual, fErr)
			So(c.DidBegin, ShouldBeTrue)
			So(c.DidCommit, ShouldBeFalse)
			So(c.DidRollback, ShouldBeTrue)
		})

		Convey("rollback on panic", func() {
			f := func() error { panic(errors.New("panic")) }
			defer func() {
				_ = recover()
				So(c.DidBegin, ShouldBeTrue)
				So(c.DidCommit, ShouldBeFalse)
				So(c.DidRollback, ShouldBeTrue)
			}()
			_ = ReadOnly(c, f)
		})
	})
}
