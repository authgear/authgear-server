package image

import (
	"bytes"
	"image"
	"image/png"
	"io"

	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type Encoder interface {
	MediaType() string
	Encode(w io.Writer, m image.Image) error
}

var CodecPNG Encoder = &pngEncoder{}

type pngEncoder struct{}

func (e *pngEncoder) MediaType() string {
	return "image/png"
}

func (e *pngEncoder) Encode(w io.Writer, m image.Image) error {
	return png.Encode(w, m)
}
func DataURIFromImage(encoder Encoder, m image.Image) (string, error) {
	var buf bytes.Buffer

	out, err := urlutil.DataURIWriter(encoder.MediaType(), &buf)
	if err != nil {
		return "", err
	}

	err = encoder.Encode(out, m)
	if err != nil {
		return "", err
	}

	out.Close()

	return buf.String(), nil
}
