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
	Convey("PathWithHash", t, func() {
		So(
			PathWithHash("logo.png", "hash1"),
			ShouldEqual,
			"logo.hash1.png",
		)

		So(
			PathWithHash("/img/logo.png", "hash2"),
			ShouldEqual,
			"/img/logo.hash2.png",
		)

		So(
			PathWithHash("/img/logo", "hash"),
			ShouldEqual,
			"/img/logo.hash",
		)
	})

	Convey("parsePathWithHash", t, func() {
		filePath, hash := ParsePathWithHash("logo.hash1.png")
		So(filePath, ShouldEqual, "logo.png")
		So(hash, ShouldEqual, "hash1")

		filePath, hash = ParsePathWithHash("/img/logo.hash2.png")
		So(filePath, ShouldEqual, "/img/logo.png")
		So(hash, ShouldEqual, "hash2")

		filePath, hash = ParsePathWithHash("/img/logo.hash")
		So(filePath, ShouldEqual, "/img/logo")
		So(hash, ShouldEqual, "hash")

		filePath, hash = ParsePathWithHash("logo.hash")
		So(filePath, ShouldEqual, "logo")
		So(hash, ShouldEqual, "hash")

		filePath, hash = ParsePathWithHash("logo")
		So(filePath, ShouldEqual, "")
		So(hash, ShouldEqual, "")
	})
}
