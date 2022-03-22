package vipsutil

import (
	"sort"
)

// ResizingModeScaleDown shrink the image so that
// the resulting image is fully fit within ResizeDimensions.
type ResizingModeScaleDown struct{}

var _ ResizingMode = ResizingModeScaleDown{}

func (ResizingModeScaleDown) Resize(image ImageDimensions, resize ResizeDimensions) ResizeResult {
	// If the source image is fit within ResizeDimensions, it is an noop.
	if image.Width <= resize.Width && image.Height <= resize.Height {
		return NoopResizeResult
	}

	var scales []float64

	scaleX := div(resize.Width, image.Width)
	scaledHeight := mul(image.Height, scaleX)
	if scaledHeight <= resize.Height {
		scales = append(scales, scaleX)
	}

	scaleY := div(resize.Height, image.Height)
	scaledWidth := mul(image.Width, scaleY)
	if scaledWidth <= resize.Width {
		scales = append(scales, scaleY)
	}

	sort.Float64s(scales)
	if len(scales) > 0 {
		return ResizeResult{
			Scale: scales[len(scales)-1],
		}
	}

	return NoopResizeResult
}
