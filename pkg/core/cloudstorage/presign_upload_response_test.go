package cloudstorage

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPresignUploadResponse(t *testing.T) {
	Convey("PresignUploadResponse", t, func() {
		Convey("NewPresignUploadResponse", func() {
			resp := NewPresignUploadResponse("myimage.png", &http.Request{
				Method: "PUT",
				URL: &url.URL{
					Scheme:   "https",
					Host:     "example.com",
					Path:     "/a/b",
					RawQuery: "a=b&c=d",
				},
				Header: http.Header{
					"Content-Length": []string{"123"},
				},
			})
			So(resp, ShouldResemble, PresignUploadResponse{
				AssetID: "myimage.png",
				Method:  "PUT",
				URL:     "https://example.com/a/b?a=b&c=d",
				Headers: []HeaderField{
					HeaderField{"Content-Length", "123"},
				},
			})
		})
	})
}
