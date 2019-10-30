package imageprocessing

import (
	"bytes"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"

	"github.com/davidbyttow/govips/pkg/vips"
	coreIo "github.com/skygeario/skygear-server/pkg/core/io"
)

// Apply applies operations to image.
func Apply(image []byte, operations []Operation) ([]byte, ImageFormat, error) {
	imageRef, err := vips.NewImageFromBuffer(image)
	if err != nil {
		return nil, "", err
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
			return nil, "", err
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

	output, finalImageType, err := ctx.Image.Export(vips.ExportParams{
		Format:         imageType,
		Quality:        quality,
		Interpretation: interpretation,
	})
	if err != nil {
		return nil, "", err
	}

	return output, ImageFormatFromVIPS(finalImageType), nil
}

func ApplyToHTTPResponse(resp *http.Response, ops []Operation) error {
	originalBody := resp.Body
	input, err := ioutil.ReadAll(originalBody)
	if err != nil {
		return err
	}
	defer originalBody.Close()

	output, imageFormat, err := Apply(input, ops)
	if err != nil {
		resp.Body = &coreIo.BytesReaderCloser{Reader: bytes.NewReader(input)}
		return nil
	}

	resp.ContentLength = int64(len(output))
	resp.Header.Set("Content-Length", strconv.Itoa(len(output)))
	resp.Header.Set("Content-Type", imageFormat.MediaType())
	resp.Body = &coreIo.BytesReaderCloser{Reader: bytes.NewReader(output)}
	return nil
}

func IsApplicableToHTTPResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	contentLength := resp.ContentLength
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	// content-length must be known and less than 20MiB.
	if contentLength < 0 || contentLength > 20*1024*1024 {
		return false
	}
	switch mediaType {
	case "image/png":
		return true
	case "image/jpeg":
		return true
	case "image/webp":
		return true
	default:
		return false
	}
}
