package stdattrs_test

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/stdattrs"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessor(t *testing.T) {
	Convey("Processor", t, func() {
		Convey("PictureAttrProcessor", func() {

			p := stdattrs.PictureAttrProcessor{
				ImagesHost: config.ImagesCDNHost("imagescdn.com"),
				AppID:      "app1",
			}

			test := func(input string, expected string) {
				result, err := p.Process(input)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, expected)
			}

			test("https://example.com/image", "https://example.com/image")
			test("authgearimages:///objectid", "https://imagescdn.com/_images/app1/objectid/profile")
			test("authgearimages://host/objectid", "https://imagescdn.com/_images/app1/objectid/profile")
			test("authgearimages://host/objectid?abc=1", "https://imagescdn.com/_images/app1/objectid/profile")
		})
	})
}
