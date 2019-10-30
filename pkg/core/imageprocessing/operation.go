package imageprocessing

import (
	"github.com/davidbyttow/govips/pkg/vips"
)

type OperationContext struct {
	Image   *vips.ImageRef
	Quality int
	Format  ImageFormat
}

// Operation applies processing to image.
type Operation interface {
	Apply(ctx *OperationContext) error
}
