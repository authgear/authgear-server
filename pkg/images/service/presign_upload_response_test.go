package service

import (
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPresignUploadResponse(t *testing.T) {
	Convey("PresignUploadResponse", t, func() {
		Convey("NewPresignUploadResponse", func() {
			resp := NewPresignUploadResponse(&http.Request{
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
			}, "appid/uuid")
			So(resp, ShouldResemble, PresignUploadResponse{
				Key:    "appid/uuid",
				Method: "PUT",
				URL:    "https://example.com/a/b?a=b&c=d",
				Headers: []HeaderField{
					HeaderField{"Content-Length", "123"},
				},
			})
		})
	})
}
