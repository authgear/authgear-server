package vipsutil

import (
	"testing"
	"testing/quick"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResizingModeScaleDown(t *testing.T) {
	Convey("ResizingModeScaleDown example-based", t, func() {
		test := func(image ImageDimensions, resize ResizeDimensions, scale float64) {
			var m ResizingModeScaleDown
			r := m.Resize(image, resize)
			So(r.Scale, ShouldEqual, scale)
		}

		// Same dimensions.
		test(ImageDimensions{100, 100}, ResizeDimensions{100, 100}, 1.0)

		// Both are smaller.
		test(ImageDimensions{50, 50}, ResizeDimensions{100, 100}, 1.0)
		test(ImageDimensions{50, 50}, ResizeDimensions{100, 100}, 1.0)
		test(ImageDimensions{50, 75}, ResizeDimensions{100, 100}, 1.0)
		test(ImageDimensions{75, 75}, ResizeDimensions{100, 100}, 1.0)

		// Width is larger.
		test(ImageDimensions{150, 100}, ResizeDimensions{100, 100}, 2.0/3.0)

		// Height is larger.
		test(ImageDimensions{100, 150}, ResizeDimensions{100, 100}, 2.0/3.0)
	})

	Convey("ResizingModeScaleDown quick", t, func() {
		f := ResultingImageNeverLargerThanResizeDimensions(ResizingModeScaleDown{})
		err := quick.Check(f, nil)
		So(err, ShouldBeNil)
	})
}
