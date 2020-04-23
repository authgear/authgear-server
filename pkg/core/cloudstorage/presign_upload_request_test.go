package cloudstorage

import (
	"net/http"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPresignUploadRequest(t *testing.T) {
	Convey("PresignUploadRequest", t, func() {
		Convey("DeriveAssetName", func() {
			Convey("Use random name", func() {
				r := PresignUploadRequest{
					Prefix: "prefix-",
					Headers: map[string]interface{}{
						"content-type":  "image/png",
						"cache-control": "no-store",
					},
				}
				assetName, err := r.DeriveAssetName()
				So(err, ShouldBeNil)
				So(strings.HasPrefix(assetName, "prefix-"), ShouldBeTrue)
				So(strings.HasSuffix(assetName, ".png"), ShouldBeTrue)
			})
		})

		Convey("Sanitize", func() {
			r := PresignUploadRequest{}
			r.Sanitize()
			So(r.Headers, ShouldResemble, map[string]interface{}{
				"cache-control": "max-age=3600",
				"content-type":  "application/octet-stream",
			})

			r = PresignUploadRequest{
				Headers: map[string]interface{}{
					"cache-control": "",
					"content-type":  "",
				},
			}
			r.Sanitize()
			So(r.Headers, ShouldResemble, map[string]interface{}{
				"cache-control": "max-age=3600",
				"content-type":  "application/octet-stream",
			})

			r = PresignUploadRequest{
				Headers: map[string]interface{}{
					"cache-control": "no-store",
					"content-type":  "application/json",
				},
			}
			r.Sanitize()
			So(r.Headers, ShouldResemble, map[string]interface{}{
				"cache-control": "no-store",
				"content-type":  "application/json",
			})
		})

		Convey("HTTPHeader", func() {
			r := PresignUploadRequest{
				Headers: map[string]interface{}{
					"content-length": "123",
					"cache-control":  "max-age=3600",
				},
			}
			httpHeader := r.HTTPHeader()
			So(httpHeader, ShouldResemble, http.Header{
				"Content-Length": []string{"123"},
				"Cache-Control":  []string{"max-age=3600"},
			})
		})
	})
}
