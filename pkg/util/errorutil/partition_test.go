package errorutil_test

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func TestPartition(t *testing.T) {
	Convey("should return (matched=provided error, unmatched=nil) for non-joined single error with matching predicate", t, func() {
		errA := errors.New("a")
		matched, notMatched := errorutil.Partition(errA, func(err error) bool { return errors.Is(err, errA) })
		So(errors.Is(matched, errA), ShouldBeTrue)
		So(errors.Is(notMatched, errA), ShouldBeFalse)
	})

	Convey("should return (matched=nil, unmatched=provided error) for non-joined single error with non-matching predicate", t, func() {
		errA := errors.New("a")
		errNotA := errors.New("not a")
		matched, notMatched := errorutil.Partition(errA, func(err error) bool { return errors.Is(err, errNotA) })
		So(errors.Is(matched, errA), ShouldBeFalse)
		So(errors.Is(notMatched, errA), ShouldBeTrue)
	})

	Convey("should partition a joined error of 2 errors", t, func() {
		errA := errors.New("a")
		errB := errors.New("b")
		joined := errors.Join(errA, errB)
		matched, notMatched := errorutil.Partition(joined, func(err error) bool {
			return errors.Is(err, errA)
		})
		So(errors.Is(matched, errA), ShouldBeTrue)
		So(errors.Is(matched, errB), ShouldBeFalse)
		So(errors.Is(notMatched, errA), ShouldBeFalse)
		So(errors.Is(notMatched, errB), ShouldBeTrue)
	})

	Convey("should partition a joined error of 6 errors", t, func() {
		errA := errors.New("a")
		errB := errors.New("b")
		errC := errors.New("c")
		errD := errors.New("d")
		errE := errors.New("e")
		errF := errors.New("f")

		joined := errors.Join(errA, errB, errC, errD, errE, errF)

		matched, notMatched := errorutil.Partition(joined, func(err error) bool {
			return errors.Is(err, errA)
		})

		So(errors.Is(matched, errA), ShouldBeTrue)
		So(errors.Is(matched, errB), ShouldBeFalse)
		So(errors.Is(matched, errC), ShouldBeFalse)
		So(errors.Is(matched, errD), ShouldBeFalse)
		So(errors.Is(matched, errE), ShouldBeFalse)
		So(errors.Is(matched, errF), ShouldBeFalse)
		So(errors.Is(notMatched, errA), ShouldBeFalse)
		So(errors.Is(notMatched, errB), ShouldBeTrue)
		So(errors.Is(notMatched, errC), ShouldBeTrue)
		So(errors.Is(notMatched, errD), ShouldBeTrue)
		So(errors.Is(notMatched, errE), ShouldBeTrue)
		So(errors.Is(notMatched, errF), ShouldBeTrue)
	})

	Convey("should partition a joined error of many errors", t, func() {
		errA := errors.New("a")
		errB := errors.New("b")
		errC := errors.New("c")
		errD := errors.New("d")
		errE := errors.New("e")
		errF := errors.New("f")

		joined := errors.Join(errA, errB, errC, errD, errE, errF)

		matched, notMatched := errorutil.Partition(joined, func(err error) bool {
			return errors.Is(err, errA)
		})

		So(errors.Is(matched, errA), ShouldBeTrue)
		So(errors.Is(matched, errB), ShouldBeFalse)
		So(errors.Is(matched, errC), ShouldBeFalse)
		So(errors.Is(matched, errD), ShouldBeFalse)
		So(errors.Is(matched, errE), ShouldBeFalse)
		So(errors.Is(matched, errF), ShouldBeFalse)
		So(errors.Is(notMatched, errA), ShouldBeFalse)
		So(errors.Is(notMatched, errB), ShouldBeTrue)
		So(errors.Is(notMatched, errC), ShouldBeTrue)
		So(errors.Is(notMatched, errD), ShouldBeTrue)
		So(errors.Is(notMatched, errE), ShouldBeTrue)
		So(errors.Is(notMatched, errF), ShouldBeTrue)
	})

	Convey("should NOT mutate the order of joined errors after partition", t, func() {
		errA := errors.New("a")
		errB := errors.New("b")
		errC := errors.New("c")
		errD := errors.New("d")
		errE := errors.New("e")
		errF := errors.New("f")

		joined := errors.Join(errA, errB, errC, errD, errE, errF) // Note order here is A,B,C,D,E,F
		_, notMatched := errorutil.Partition(joined, func(err error) bool {
			return errors.Is(err, errA)
		})

		uw, ok := notMatched.(interface{ Unwrap() []error })
		So(ok, ShouldBeTrue)
		notMatchedErrs := uw.Unwrap()
		So(notMatchedErrs, ShouldResemble, []error{errB, errC, errD, errE, errF}) // Note order here is B,C,D,E,F
	})
}
