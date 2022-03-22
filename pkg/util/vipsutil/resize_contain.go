package vipsutil

import (
	"sort"
)

type ResizingModeContain struct{}

var _ ResizingMode = ResizingModeContain{}

func (ResizingModeContain) Resize(image ImageDimensions, resize ResizeDimensions) ResizeResult {
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
