package image

import (
	"image"
	"image/color"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDataURIFromImage(t *testing.T) {
	Convey("DataURIFromImage", t, func() {
		width := 40
		height := 30
		img := image.NewRGBA(image.Rectangle{
			image.Point{0, 0},
			image.Point{width, height},
		})

		makeColor := func(n int, d int) color.Color {
			ratio := float64(n) / float64(d)
			v := uint8(ratio*0xff) & 0xff
			return color.RGBA{v, v, v, 0xff}
		}

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				color := makeColor(x+y, width+height)
				img.Set(x, y, color)
			}
		}

		dataURI, err := DataURIFromImage(CodecPNG, img)
		So(err, ShouldBeNil)
		So(dataURI, ShouldEqual, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACgAAAAeCAIAAADRv8uKAAAAfElEQVR4nOyVsQrAIAxEQy7//82ldCgttjqYlyWLOAgPnxcvzExSRFzrymbLSS+hSvIS6n1jmHqCS6irqjOszFUnvcVEdV4C/lSn5u5TdXbax6qBGRuoZib7rRr7T7yE+lBNUrudMGq3E0btdsKo3U4YtdsJo0o6AgAA//+E4DNsK371OQAAAABJRU5ErkJggg==")
	})
}
