package errorutil_test

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type testUnwrap struct {
	Label string
	Inner error
}

func (e *testUnwrap) Error() string { return fmt.Sprintf("%v: %v", e.Label, e.Inner.Error()) }
func (e *testUnwrap) Unwrap() error { return e.Inner }

func TestUnwrap(t *testing.T) {
	wrap := func(err error, label string) error {
		return &testUnwrap{Label: label, Inner: err}
	}

	a := errors.New("a")
	wrapA := wrap(a, "label a")
	b := errors.New("b")
	wrapB := wrap(b, "label b")
	c := errors.New("c")
	wrapC := wrap(c, "label c")
	d := errors.New("d")
	wrapD := wrap(d, "label d")

	wrapAwrapB := errors.Join(wrapA, wrapB)
	wrapCwrapD := errors.Join(wrapC, wrapD)
	wrapwrapAwrapB := wrap(wrapAwrapB, "label ab")
	err := errors.Join(wrapwrapAwrapB, wrapCwrapD)

	elements := []error{
		a,
		wrapA,
		b,
		wrapB,
		c,
		wrapC,
		d,
		wrapD,
	}

	collect := func(err error, arr []error) []error {
		for _, e := range elements {
			// We intentionally do not use errors.Is here.
			if err == e {
				return append(arr, err)
			}
		}
		return arr
	}

	test := func(err error, expected []error) {
		var order []error
		errorutil.Unwrap(err, func(err error) {
			order = collect(err, order)
		})
		So(order, ShouldResemble, expected)
	}

	Convey("Unwrap", t, func() {
		test(a, []error{a})
		test(wrapA, []error{wrapA, a})
		test(wrapAwrapB, []error{wrapA, a, wrapB, b})
		test(wrapwrapAwrapB, []error{wrapA, a, wrapB, b})
		test(err, []error{wrapA, a, wrapB, b, wrapC, c, wrapD, d})
	})
}
