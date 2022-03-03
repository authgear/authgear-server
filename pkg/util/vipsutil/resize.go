package vipsutil

import (
	"image"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
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

// ApplyTo applies r to imageRef using kernel.
func (r ResizeResult) ApplyTo(imageRef *vips.ImageRef, kernel vips.Kernel) (err error) {
	if r.Scale != 1.0 {
		err = imageRef.Resize(r.Scale, kernel)
		if err != nil {
			return
		}
	}

	if r.Crop != nil {
		dx := r.Crop.Dx()
		dy := r.Crop.Dy()
		x := r.Crop.Min.X
		y := r.Crop.Min.Y
		err = imageRef.ExtractArea(x, y, dx, dy)
		if err != nil {
			return
		}
	}

	return nil
}

// ResizingMode is the abstraction of different resizing flavors.
type ResizingMode interface {
	Resize(image ImageDimensions, resize ResizeDimensions) ResizeResult
}

func div(x int, y int) float64 {
	return float64(x) / float64(y)
}

func mul(x int, f float64) int {
	return int(math.Round(float64(x) * f))
}
