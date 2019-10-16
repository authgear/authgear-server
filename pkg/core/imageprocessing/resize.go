package imageprocessing

type ResizeScalingMode string

const (
	ResizeScalingModeLfit  ResizeScalingMode = "lfit"
	ResizeScalingModeMfit  ResizeScalingMode = "mfit"
	ResizeScalingModePad   ResizeScalingMode = "pad"
	ResizeScalingModeFixed ResizeScalingMode = "fixed"
)

const ResizeScalingModeDefault = ResizeScalingModeLfit

var ResizeDefaultColor = ColorWhite

type Resize struct {
	ScalingMode ResizeScalingMode
	Width       int
	Height      int
	LongerSide  int
	ShorterSide int
	Color       Color
}

var _ Operation = &Resize{}

func (o *Resize) Apply(ctx *OperationContext) error {
	// TODO(imageprocessing)
	return nil
}

// NewResize returns a Resize with default values.
func NewResize() *Resize {
	return &Resize{
		ScalingMode: ResizeScalingModeDefault,
		Color:       ResizeDefaultColor,
	}
}
