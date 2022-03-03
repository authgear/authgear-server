package vipsutil

import (
	"image"
	"sort"
)

type ResizingModeCover struct{}

var _ ResizingMode = ResizingModeCover{}

func (ResizingModeCover) Resize(img ImageDimensions, resize ResizeDimensions) ResizeResult {
	var scales []float64

	scaleX := div(resize.Width, img.Width)
	scaledHeight := mul(img.Height, scaleX)
	if scaledHeight >= resize.Height {
		scales = append(scales, scaleX)
	}

	scaleY := div(resize.Height, img.Height)
	scaledWidth := mul(img.Width, scaleY)
	if scaledWidth >= resize.Width {
		scales = append(scales, scaleY)
	}

	sort.Float64s(scales)
	if len(scales) <= 0 {
		return NoopResizeResult
	}

	// Use the smaller scale.
	scale := scales[0]
	scaledWidth = mul(img.Width, scale)
	scaledHeight = mul(img.Height, scale)

	// No crop is needed.
	if scaledWidth <= resize.Width && scaledHeight <= resize.Height {
		return ResizeResult{
			Scale: scale,
		}
	}

	cropX := (scaledWidth - resize.Width) / 2
	cropY := (scaledHeight - resize.Height) / 2
	cropWidth := resize.Width
	cropHeight := resize.Height
	cropRect := image.Rect(cropX, cropY, cropX+cropWidth, cropY+cropHeight)
	return ResizeResult{
		Scale: scale,
		Crop:  &cropRect,
	}
}
