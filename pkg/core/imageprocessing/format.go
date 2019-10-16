package imageprocessing

type ImageFormat string

const (
	ImageFormatJPEG ImageFormat = "jpg"
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatWebP ImageFormat = "webp"
)

type Format struct {
	ImageFormat ImageFormat
}

var _ Operation = &Format{}

func (o *Format) Apply(ctx *OperationContext) error {
	// TODO(imageprocessing)
	return nil
}
