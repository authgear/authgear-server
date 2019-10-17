package imageprocessing

import (
	"github.com/davidbyttow/govips/pkg/vips"
)

type ImageFormat string

const (
	ImageFormatJPEG ImageFormat = "jpg"
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatWebP ImageFormat = "webp"
)

func (f ImageFormat) VIPSImageType() vips.ImageType {
	switch f {
	case ImageFormatJPEG:
		return vips.ImageTypeJPEG
	case ImageFormatPNG:
		return vips.ImageTypePNG
	case ImageFormatWebP:
		return vips.ImageTypeWEBP
	default:
		panic("unreachable")
	}
}

type Format struct {
	ImageFormat ImageFormat
}

var _ Operation = &Format{}

func (o *Format) Apply(ctx *OperationContext) error {
	ctx.Format = o.ImageFormat
	return nil
}
