package vipsutil

import (
	"math"
	"math/rand"
	"reflect"
)

type Length int

func (Length) Generate(rand *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(Length(rand.Intn(4096) + 1))
}

func ResultingImageNeverLargerThanResizeDimensions(m ResizingMode) func(imageWidth Length, imageHeight Length, resizeWidth Length, resizeHeight Length) bool {
	return func(imageWidth Length, imageHeight Length, resizeWidth Length, resizeHeight Length) bool {
		image := ImageDimensions{Width: int(imageWidth), Height: int(imageHeight)}
		resize := ResizeDimensions{Width: int(resizeWidth), Height: int(resizeHeight)}
		r := m.Resize(image, resize)

		resultWidth := int(math.Round(r.Scale * float64(imageWidth)))
		resultHeight := int(math.Round(r.Scale * float64(imageHeight)))

		if resultWidth > int(resizeWidth) || resultHeight > int(resizeHeight) {
			return false
		}
		return true
	}
}
