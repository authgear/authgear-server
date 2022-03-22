package vipsutil

import (
	"github.com/davidbyttow/govips/v2/vips"
)

func Export(imageRef *vips.ImageRef) ([]byte, *vips.ImageMetadata, error) {
	imageType := imageRef.Format()
	switch imageType {
	case vips.ImageTypeJPEG:
		return imageRef.ExportJpeg(&vips.JpegExportParams{
			StripMetadata: true,
			Quality:       80,
			Interlace:     true,
			SubsampleMode: vips.VipsForeignSubsampleOn,
		})
	case vips.ImageTypePNG:
		return imageRef.ExportPng(&vips.PngExportParams{
			StripMetadata: true,
			Compression:   6,
		})
	case vips.ImageTypeGIF:
		return imageRef.ExportGIF(&vips.GifExportParams{
			StripMetadata: true,
			Quality:       75,
		})
	case vips.ImageTypeWEBP:
		return imageRef.ExportWebp(&vips.WebpExportParams{
			StripMetadata:   true,
			Quality:         75,
			ReductionEffort: 4,
		})
	default:
		return imageRef.ExportNative()
	}
}
