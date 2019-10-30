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

	switch o.ScalingMode {
	case ResizeScalingModeLfit:
		scale, _, _ := o.ResolveLfit(originalWidth, originalHeight)
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
		scale, _, _ := o.ResolveMfit(originalWidth, originalHeight)
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
		targetWidth, targetHeight := o.ResolveTargetDimension(originalWidth, originalHeight)
		scale, contentWidth, contentHeight := o.ResolveMfit(originalWidth, originalHeight)
		// Do not support upscaling
		if scale >= 1 {
			return nil
		}
		kernal := vips.InputInt("kernel", int(vips.KernelLanczos3))
		err := ctx.Image.Resize(scale, kernal)
		if err != nil {
			return err
		}
		extractX, extractY, extractWidth, extractHeight, needExtract := o.ResolveExtractArea(
			targetWidth,
			targetHeight,
			contentWidth,
			contentHeight,
		)
		if needExtract {
			err := ctx.Image.ExtractArea(extractX, extractY, extractWidth, extractHeight)
			if err != nil {
				return err
			}
		}
	case ResizeScalingModePad:
		targetWidth, targetHeight := o.ResolveTargetDimension(originalWidth, originalHeight)
		scale, contentWidth, contentHeight := o.ResolveLfit(originalWidth, originalHeight)
		// Do not support upscaling
		if scale >= 1 {
			return nil
		}
		kernal := vips.InputInt("kernel", int(vips.KernelLanczos3))
		err := ctx.Image.Resize(scale, kernal)
		if err != nil {
			return err
		}
		embedX, embedY, embedWidth, embedHeight, needEmbed := o.ResolveEmbedArea(
			targetWidth,
			targetHeight,
			contentWidth,
			contentHeight,
		)
		if needEmbed {
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
				embedWidth,
				embedHeight,
				vips.InputInt("extend", int(vips.ExtendBackground)),
				vips.InputBackground("background", float64(o.Color.R), float64(o.Color.G), float64(o.Color.B)),
			)
			if err != nil {
				return err
			}
		}
	case ResizeScalingModeFixed:
		targetWidth, targetHeight := o.ResolveTargetDimension(originalWidth, originalHeight)
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

func (o *Resize) ResolveTargetDimension(originalWidth, originalHeight int) (targetWidth, targetHeight int) {
	aspectRatio := ratio(originalWidth, originalHeight)

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

	return
}

func (o *Resize) ResolveLfit(originalWidth, originalHeight int) (scale float64, contentWidth, contentHeight int) {
	aspectRatio := ratio(originalWidth, originalHeight)
	targetWidth, targetHeight := o.ResolveTargetDimension(originalWidth, originalHeight)

	// w1 and h1 are in the same aspect ratio.
	w1 := targetWidth
	h1 := roundFloat(float64(targetWidth) / aspectRatio)

	// w2 and h2 are in the same aspect ratio.
	h2 := targetHeight
	w2 := roundFloat(float64(targetHeight) * aspectRatio)

	if w1 <= targetWidth && h1 <= targetHeight {
		scale = ratio(w1, originalWidth)
		contentWidth = w1
		contentHeight = h1
	} else {
		scale = ratio(w2, originalWidth)
		contentWidth = w2
		contentHeight = h2
	}

	return
}

func (o *Resize) ResolveMfit(originalWidth, originalHeight int) (scale float64, contentWidth, contentHeight int) {
	aspectRatio := ratio(originalWidth, originalHeight)
	targetWidth, targetHeight := o.ResolveTargetDimension(originalWidth, originalHeight)

	// w1 and h1 are in the same aspect ratio.
	w1 := targetWidth
	h1 := roundFloat(float64(targetWidth) / aspectRatio)

	// w2 and h2 are in the same aspect ratio.
	h2 := targetHeight
	w2 := roundFloat(float64(targetHeight) * aspectRatio)

	if w1 >= targetWidth && h1 >= targetHeight {
		scale = ratio(w1, originalWidth)
		contentWidth = w1
		contentHeight = h1
	} else {
		scale = ratio(w2, originalWidth)
		contentWidth = w2
		contentHeight = h2
	}

	return
}

func (o *Resize) ResolveExtractArea(targetWidth, targetHeight, contentWidth, contentHeight int) (extractX, extractY, extractWidth, extractHeight int, ok bool) {
	if contentWidth > targetWidth || contentHeight > targetHeight {
		ok = true
		if contentWidth > targetWidth {
			extractWidth = targetWidth
			extractHeight = targetHeight
			extractX = (contentWidth - targetWidth) / 2
		}
		if contentHeight > targetHeight {
			extractWidth = targetWidth
			extractHeight = targetHeight
			extractY = (contentHeight - targetHeight) / 2
		}
	}

	return
}

func (o *Resize) ResolveEmbedArea(targetWidth, targetHeight, contentWidth, contentHeight int) (embedX, embedY, embedWidth, embedHeight int, ok bool) {
	if targetWidth > contentWidth || targetHeight > contentHeight {
		ok = true
		embedHeight = contentHeight
		embedWidth = contentWidth
		if targetWidth > contentWidth {
			embedWidth = targetWidth
			embedX = (targetWidth - contentWidth) / 2
		}
		if targetHeight > contentHeight {
			embedHeight = targetHeight
			embedY = (targetHeight - contentHeight) / 2
		}
	}

	return
}

// NewResize returns a Resize with default values.
func NewResize() *Resize {
	return &Resize{
		ScalingMode: ResizeScalingModeDefault,
		Color:       ResizeDefaultColor,
	}
}
