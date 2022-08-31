package filepathutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExt(t *testing.T) {
	Convey("Ext", t, func() {
		So(Ext("logo.hash1.png"), ShouldEqual, ".png")
		So(Ext("logo.hash1.png.map"), ShouldEqual, ".png.map")
		So(Ext("logo.hash1"), ShouldEqual, ".hash1")
	})
}

func TestParseHashedPath(t *testing.T) {
	Convey("ParsePathWithHash", t, func() {
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

func TestMakeHashedPath(t *testing.T) {
	Convey("MakeHashedPath", t, func() {
		So(MakeHashedPath("logo.png", "hash1"), ShouldEqual, "logo.hash1.png")
		So(MakeHashedPath("/img/logo.png", "hash2"), ShouldEqual, "/img/logo.hash2.png")
		So(MakeHashedPath("/img/logo", "hash"), ShouldEqual, "/img/logo.hash")
		So(MakeHashedPath("logo", "hash"), ShouldEqual, "logo.hash")
		So(MakeHashedPath("logo", ""), ShouldEqual, "logo")
		So(MakeHashedPath("script.js.map", "hash"), ShouldEqual, "script.hash.js.map")
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
