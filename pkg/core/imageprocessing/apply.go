package imageprocessing

import (
	"github.com/davidbyttow/govips/pkg/vips"
)

// Apply applies operations to image.
func Apply(image []byte, operations []Operation) ([]byte, error) {
	imageRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		return nil, err
	}
	defer imageRef.Close()

	// Remember the original format.
	imageType := imageRef.Format()
	// Remember the original interpretation
	interpretation := imageRef.Interpretation()
	ctx := &OperationContext{
		Image: imageRef,
	}

	// Apply the operations in order.
	for _, op := range operations {
		err := op.Apply(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Change the format if needed.
	if ctx.Format != "" {
		imageType = ctx.Format.VIPSImageType()
	}

	// Change the quality if needed.
	quality := DefaultQuality
	if ctx.Quality != 0 {
		quality = ctx.Quality
	}

	output, _, err := ctx.Image.Export(vips.ExportParams{
		Format:         imageType,
		Quality:        quality,
		Interpretation: interpretation,
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}
