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

func ImageFormatFromVIPS(f vips.ImageType) ImageFormat {
	switch f {
	case vips.ImageTypeJPEG:
		return ImageFormatJPEG
	case vips.ImageTypePNG:
		return ImageFormatPNG
	case vips.ImageTypeWEBP:
		return ImageFormatWebP
	default:
		panic("unreachable")
	}
}

func (f ImageFormat) MediaType() string {
	switch f {
	case ImageFormatJPEG:
		return "image/jpeg"
	case ImageFormatPNG:
		return "image/png"
	case ImageFormatWebP:
		return "image/webp"
	default:
		panic("unreachable")
	}
}

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
