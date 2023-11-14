package web

import (
	"image"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

func CreateQRCodeImage(content string, width int, height int, level qr.ErrorCorrectionLevel) (image.Image, error) {
	b, err := qr.Encode(content, level, qr.Auto)

	if err != nil {
		return nil, err
	}

	b, err = barcode.Scale(b, width, height)

	if err != nil {
		return nil, err
	}

	return b, nil
}
