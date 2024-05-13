package stdattrs

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestPictureTransformer(t *testing.T) {
	Convey("PictureTransformer", t, func() {
		t := &PictureTransformer{
			HTTPProto:     httputil.HTTPProto("https"),
			HTTPHost:      httputil.HTTPHost("app.authgearapps.com"),
			ImagesCDNHost: config.ImagesCDNHost("authgearappscdn.com"),
		}

		Convey("has no effect on non-picture attributes", func() {
			test := func(key string, anyValue interface{}) {
				actual, err := t.StorageFormToRepresentationForm(key, anyValue)
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, anyValue)

				actual, err = t.RepresentationFormToStorageForm(key, anyValue)
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, anyValue)
			}

			test("email", "")
			test("name", "")
		})

		Convey("has no effect if the value is not a string", func() {
			test := func(anyValue interface{}) {
				actual, err := t.StorageFormToRepresentationForm(stdattrs.Picture, anyValue)
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, anyValue)

				actual, err = t.RepresentationFormToStorageForm(stdattrs.Picture, anyValue)
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, anyValue)
			}

			test(nil)
			test(false)
			test(1)
		})

		Convey("transform from storage form to representation form", func() {
			test := func(storage string, repr string) {
				actual, err := t.StorageFormToRepresentationForm(stdattrs.Picture, storage)
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, repr)
			}

			test("authgearimages:///app/objectid", "https://authgearappscdn.com/_images/app/objectid/profile")
			t.ImagesCDNHost = ""
			test("authgearimages:///app/objectid", "https://app.authgearapps.com/_images/app/objectid/profile")
		})

		Convey("transform from representation form to storage form", func() {
			test := func(repr string, storage string) {
				actual, err := t.RepresentationFormToStorageForm(stdattrs.Picture, repr)
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, storage)
			}

			test("https://authgearappscdn.com/_images/app/objectid/profile", "authgearimages:///app/objectid")
			test("https://authgearappscdn.com/_images/app/objectid/", "authgearimages:///app/objectid")
			test("https://authgearappscdn.com/_images/app/objectid", "authgearimages:///app/objectid")
			test("https://authgearappscdn.com/app/objectid", "https://authgearappscdn.com/app/objectid")
		})
	})
}
