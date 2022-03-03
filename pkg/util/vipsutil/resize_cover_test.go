package vipsutil

import (
	"testing"
	"testing/quick"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResizingModeCover(t *testing.T) {
	Convey("ResizingModeCover example-based", t, func() {
		test := func(image ImageDimensions, resize ResizeDimensions, scale float64) {
			var m ResizingModeCover
			r := m.Resize(image, resize)
			So(r.Scale, ShouldEqual, scale)
		}

		// Same dimensions.
		test(ImageDimensions{100, 100}, ResizeDimensions{100, 100}, 1.0)

		// Both are smaller.
		test(ImageDimensions{50, 50}, ResizeDimensions{100, 100}, 2.0)
		test(ImageDimensions{50, 75}, ResizeDimensions{100, 100}, 2.0)
		test(ImageDimensions{75, 75}, ResizeDimensions{100, 100}, 4.0/3.0)

		// Width is larger.
		test(ImageDimensions{150, 100}, ResizeDimensions{100, 100}, 1.0)

		// Height is larger.
		test(ImageDimensions{100, 150}, ResizeDimensions{100, 100}, 1.0)
	})

	Convey("ResizingModeCover quick", t, func() {
		f := ResultingImageIsCropped(ResizingModeCover{})
		err := quick.Check(f, nil)
		So(err, ShouldBeNil)
	})
}
