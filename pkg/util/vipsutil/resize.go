package vipsutil

import (
	"image"
	"math"
)

// ImageDimensions is the dimensions of the source image.
type ImageDimensions struct {
	// Width is the width of the source image.
	Width int
	// Height is the height of the source image.
	Height int
}

// ResizeDimensions is the input parameters of the resize operation.
type ResizeDimensions struct {
	// Width is the maximum width of the resulting image.
	Width int
	// Height is the maximum height of the resulting image.
	Height int
}

// ResizeResult is the result of the resize operation.
type ResizeResult struct {
	// Scale is the scale of the resulting image.
	// If it is 1, then scaling is not performed.
	Scale float64
	// Crop is necessary sometimes if the aspect ratios do not match.
	Crop *image.Rectangle
}

var NoopResizeResult = ResizeResult{
	Scale: 1.0,
}

// ResizingMode is the abstraction of different resizing flavors.
type ResizingMode interface {
	Resize(image ImageDimensions, resize ResizeDimensions) ResizeResult
}

type ResizingModeType string

const (
	ResizingModeTypeScaleDown ResizingModeType = "scale-down"
	ResizingModeTypeContain   ResizingModeType = "contain"
	ResizingModeTypeCover     ResizingModeType = "cover"
)

func ResizingModeFromType(t ResizingModeType) ResizingMode {
	switch t {
	case ResizingModeTypeScaleDown:
		return ResizingModeScaleDown{}
	case ResizingModeTypeContain:
		return ResizingModeContain{}
	case ResizingModeTypeCover:
		return ResizingModeCover{}
	default:
		return ResizingModeCover{}
	}
}

func div(x int, y int) float64 {
	return float64(x) / float64(y)
}

func mul(x int, f float64) int {
	return int(math.Round(float64(x) * f))
}
