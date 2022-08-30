package filepathutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseHashedPath(t *testing.T) {
	Convey("parsePathWithHash", t, func() {
		filePath, hash, _ := ParseHashedPath("logo.hash1.png")
		So(filePath, ShouldEqual, "logo.png")
		So(hash, ShouldEqual, "hash1")

		filePath, hash, _ = ParseHashedPath("/img/logo.hash2.png")
		So(filePath, ShouldEqual, "/img/logo.png")
		So(hash, ShouldEqual, "hash2")

		filePath, hash, _ = ParseHashedPath("/img/logo.hash")
		So(filePath, ShouldEqual, "/img/logo")
		So(hash, ShouldEqual, "hash")

		filePath, hash, _ = ParseHashedPath("logo.hash")
		So(filePath, ShouldEqual, "logo")
		So(hash, ShouldEqual, "hash")

		filePath, hash, _ = ParseHashedPath("logo")
		So(filePath, ShouldEqual, "")
		So(hash, ShouldEqual, "")

		filePath, hash, _ = ParseHashedPath("script.hash.js.map")
		So(filePath, ShouldEqual, "script.js.map")
		So(hash, ShouldEqual, "hash")
	})
}

func TestIsSourceMapPath(t *testing.T) {
	Convey("IsSourceMapPath", t, func() {
		result := IsSourceMapPath("script.hash.js.map")
		So(result, ShouldEqual, true)

		result = IsSourceMapPath("style.css")
		So(result, ShouldEqual, false)
	})
}
