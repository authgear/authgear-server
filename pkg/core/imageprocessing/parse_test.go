package imageprocessing

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Parse", t, func() {
		Convey("valid cases", func() {
			cases := []struct {
				Input    string
				Expected []Operation
			}{
				{"image/format,jpg", []Operation{
					&Format{
						ImageFormat: ImageFormatJPEG,
					},
				}},
				{"image/quality,Q_85", []Operation{
					&Quality{
						AbsoluteQuality: 85,
					},
				}},
				{"image/resize,m_fixed,w_1,h_2,l_3,s_4,color_FFEEDD", []Operation{
					&Resize{
						ScalingMode: ResizeScalingModeFixed,
						Width:       1,
						Height:      2,
						LongerSide:  3,
						ShorterSide: 4,
						Color: Color{
							R: 255,
							G: 238,
							B: 221,
						},
					},
				}},
				{
					"image/resize,m_fixed,w_1,h_2,l_3,s_4,color_FFEEDD/format,jpg/quality,Q_85",
					[]Operation{
						&Resize{
							ScalingMode: ResizeScalingModeFixed,
							Width:       1,
							Height:      2,
							LongerSide:  3,
							ShorterSide: 4,
							Color: Color{
								R: 255,
								G: 238,
								B: 221,
							},
						},
						&Format{
							ImageFormat: ImageFormatJPEG,
						},
						&Quality{
							AbsoluteQuality: 85,
						},
					}},
			}

			for _, c := range cases {
				actual, err := Parse(c.Input)
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, c.Expected)
			}
		})

		Convey("error cases", func() {
			cases := []struct {
				Input  string
				ErrMsg string
			}{
				{"", "invalid asset type: "},

				{"image/unknown", "invalid operation: unknown"},

				{"image/format", "invalid format: "},

				{"image/quality", "invalid quality: "},
				{"image/quality,Q_a", "value 'a' is not an integer"},
				{"image/quality,Q_101", "value '101' is not in range [1,100]"},

				{"image/resize,m", "invalid scaling mode: "},
				{"image/resize,m_unknown", "invalid scaling mode: unknown"},

				{"image/resize,w", "value '' is not an integer"},
				{"image/resize,w_0", "value '0' is not in range [1,4096]"},

				{"image/resize,color", "invalid color: "},
				{"image/resize,color_G", "invalid color: G"},

				{"image/resize,w_1/resize,w_1/resize,w_1/resize,w_1/resize,w_1/resize,w_1/resize,w_1/resize,w_1/resize,w_1/resize,w_1/resize,w_1", "query too long"},
			}
			for _, c := range cases {
				_, err := Parse(c.Input)
				So(err, ShouldBeError, c.ErrMsg)
			}
		})
	})
}
