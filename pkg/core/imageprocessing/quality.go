package imageprocessing

type Quality struct {
	// AbsoluteQuality is in range [1,100].
	AbsoluteQuality int
}

const DefaultQuality = 85

var _ Operation = &Quality{}

func (o *Quality) Apply(ctx *OperationContext) error {
	// TODO(imageprocessing)
	return nil
}
