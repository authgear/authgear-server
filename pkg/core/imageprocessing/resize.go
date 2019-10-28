package imageprocessing

import (
	"github.com/davidbyttow/govips/pkg/vips"
)

type ResizeScalingMode string

func (m ResizeScalingMode) ShouldInferAnotherSideWithAspectRatio() bool {
	switch m {
	case ResizeScalingModeLfit:
		return true
	case ResizeScalingModeMfit:
		return true
	case ResizeScalingModeFill:
		return false
	case ResizeScalingModePad:
		return false
	case ResizeScalingModeFixed:
		return true
	default:
		panic("unreachable")
	}
}

const (
	ResizeScalingModeLfit  ResizeScalingMode = "lfit"
	ResizeScalingModeMfit  ResizeScalingMode = "mfit"
	ResizeScalingModeFill  ResizeScalingMode = "fill"
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

// nolint: gocyclo
func (o *Resize) Apply(ctx *OperationContext) error {
	originalWidth := ctx.Image.Width()
	originalHeight := ctx.Image.Height()
	aspectRatio := ratio(originalWidth, originalHeight)

	targetWidth := 0
	targetHeight := 0
	var longerTargetSide *int
	var shorterTargetSide *int

	if originalWidth > originalHeight {
		longerTargetSide = &targetWidth
		shorterTargetSide = &targetHeight
	} else {
		longerTargetSide = &targetHeight
		shorterTargetSide = &targetWidth
	}

	// Set targetWidth and targetHeight if they are specified directly.
	if o.LongerSide != 0 {
		*longerTargetSide = o.LongerSide
	}
	if o.ShorterSide != 0 {
		*shorterTargetSide = o.ShorterSide
	}
	if o.Width != 0 {
		targetWidth = o.Width
	}
	if o.Height != 0 {
		targetHeight = o.Height
	}

	// w h l s are not specified, w and h resolve to original size.
	if targetWidth == 0 && targetHeight == 0 {
		targetWidth = originalWidth
		targetHeight = originalHeight
	}

	// w is not specified
	if targetWidth == 0 {
		if o.ScalingMode.ShouldInferAnotherSideWithAspectRatio() {
			targetWidth = roundFloat(float64(targetHeight) * aspectRatio)
		} else {
			targetWidth = targetHeight
		}
	}

	// h is not specified
	if targetHeight == 0 {
		if o.ScalingMode.ShouldInferAnotherSideWithAspectRatio() {
			targetHeight = roundFloat(float64(targetWidth) / aspectRatio)
		} else {
			targetHeight = targetWidth
		}
	}

	// targetWidth and targetHeight are now non-zero.
	// But the aspect ratio may not be the same as the original one.

	// w1 and h1 are in the same aspect ratio.
	w1 := targetWidth
	h1 := roundFloat(float64(targetWidth) / aspectRatio)

	// w2 and h2 are in the same aspect ratio.
	h2 := targetHeight
	w2 := roundFloat(float64(targetHeight) * aspectRatio)

	switch o.ScalingMode {
	case ResizeScalingModeLfit:
		// The final image is at most w x h.
		var scale float64
		if w1 <= targetWidth && h1 <= targetHeight {
			scale = ratio(w1, originalWidth)
		} else {
			scale = ratio(w2, originalWidth)
		}
		// Do not support upscaling
		if scale >= 1 {
			return nil
		}
		kernal := vips.InputInt("kernel", int(vips.KernelLanczos3))
		err := ctx.Image.Resize(scale, kernal)
		if err != nil {
			return err
		}
	case ResizeScalingModeMfit:
		// The final image is at least w x h.
		var scale float64
		if w1 >= targetWidth && h1 >= targetHeight {
			scale = ratio(w1, originalWidth)
		} else {
			scale = ratio(w2, originalWidth)
		}
		// Do not support upscaling
		if scale >= 1 {
			return nil
		}
		kernal := vips.InputInt("kernel", int(vips.KernelLanczos3))
		err := ctx.Image.Resize(scale, kernal)
		if err != nil {
			return err
		}
	case ResizeScalingModeFill:
		// The content image is at least w x h.
		var scale float64
		var contentWidth int
		var contentHeight int
		if w1 >= targetWidth && h1 >= targetHeight {
			scale = ratio(w1, originalWidth)
			contentWidth = w1
			contentHeight = h1
		} else {
			scale = ratio(w2, originalWidth)
			contentWidth = w2
			contentHeight = h2
		}
		// Do not support upscaling
		if scale >= 1 {
			return nil
		}
		kernal := vips.InputInt("kernel", int(vips.KernelLanczos3))
		err := ctx.Image.Resize(scale, kernal)
		if err != nil {
			return err
		}
		// We need to crop the content image if needed.
		if contentWidth > targetWidth || contentHeight > targetHeight {
			var extractX, extractY, extractWidth, extractHeight int
			if contentWidth > targetWidth {
				extractWidth = targetWidth
				extractHeight = targetHeight
				extractX = (contentWidth - targetWidth) >> 1
			}
			if contentHeight > targetHeight {
				extractWidth = targetWidth
				extractHeight = targetHeight
				extractY = (contentHeight - targetHeight) >> 1
			}
			err := ctx.Image.ExtractArea(extractX, extractY, extractWidth, extractHeight)
			if err != nil {
				return err
			}
		}
	case ResizeScalingModePad:
		// The content image is at most w x h.
		var scale float64
		var contentWidth int
		var contentHeight int
		if w1 <= targetWidth && h1 <= targetHeight {
			scale = ratio(w1, originalWidth)
			contentWidth = w1
			contentHeight = h1
		} else {
			scale = ratio(w2, originalWidth)
			contentWidth = w2
			contentHeight = h2
		}
		// Do not support upscaling
		if scale >= 1 {
			return nil
		}
		kernal := vips.InputInt("kernel", int(vips.KernelLanczos3))
		err := ctx.Image.Resize(scale, kernal)
		if err != nil {
			return err
		}
		// We need to embed the content image if needed.
		if targetWidth > contentWidth || targetHeight > contentHeight {
			var embedX, embedY, embedW, embedH int
			if targetWidth > contentWidth {
				embedW = targetWidth
				embedH = contentHeight
				embedX = (targetWidth - contentWidth) >> 1
			}
			if targetHeight > contentHeight {
				embedH = targetHeight
				embedW = contentWidth
				embedY = (targetHeight - contentHeight) >> 1
			}
			// embed expects bands <= 3.
			// So we must flatten the image first.
			if ctx.Image.Bands() > 3 {
				err := ctx.Image.Flatten(
					vips.InputBackground("background", float64(o.Color.R), float64(o.Color.G), float64(o.Color.B)),
				)
				if err != nil {
					return err
				}
			}
			err := ctx.Image.Embed(
				embedX,
				embedY,
				embedW,
				embedH,
				vips.InputInt("extend", int(vips.ExtendBackground)),
				vips.InputBackground("background", float64(o.Color.R), float64(o.Color.G), float64(o.Color.B)),
			)
			if err != nil {
				return err
			}
		}
	case ResizeScalingModeFixed:
		scale := ratio(targetWidth, originalWidth)
		vscale := ratio(targetHeight, originalHeight)
		// No change at all
		if scale == 1 && vscale == 1 {
			return nil
		}
		// Do not support upscaling
		if scale > 1 || vscale > 1 {
			return nil
		}
		kernal := vips.InputInt("kernel", int(vips.KernelLanczos3))
		vscaleOpt := vips.InputDouble("vscale", vscale)
		err := ctx.Image.Resize(scale, vscaleOpt, kernal)
		if err != nil {
			return err
		}
	default:
		panic("unreachable")
	}

	return nil
}

// NewResize returns a Resize with default values.
func NewResize() *Resize {
	return &Resize{
		ScalingMode: ResizeScalingModeDefault,
		Color:       ResizeDefaultColor,
	}
}
