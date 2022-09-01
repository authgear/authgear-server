package web

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStaticAssetURL(t *testing.T) {
	Convey("StaticAssetURL", t, func() {
		url, err := staticAssetURL("http://example.com", "/static", "asset.png")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://example.com/static/asset.png")

		url, err = staticAssetURL("http://example.com", "", "asset.png")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://example.com/asset.png")

		url, err = staticAssetURL("http://example.com", "https://cdn.example.com", "img/logo.png")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "https://cdn.example.com/img/logo.png")

		url, err = staticAssetURL("http://example.com", "//cdn.example.com", "main.css")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "http://cdn.example.com/main.css")

		url, err = staticAssetURL("https://example.com", "//cdn.example.com", "main.css")
		So(err, ShouldBeNil)
		So(url, ShouldEqual, "https://cdn.example.com/main.css")
	})
}

func TestHasedPath(t *testing.T) {
	Convey("LookLikeAHash", t, func() {
		result := LookLikeAHash("c90cf340")
		So(result, ShouldEqual, true)

		result = LookLikeAHash("123")
		So(result, ShouldEqual, false)
	})
}
