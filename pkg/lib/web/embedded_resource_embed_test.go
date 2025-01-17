package web

import (
	"embed"
	"errors"
	"io"
	"io/fs"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//go:embed testdata/embed
var testEmbedFS embed.FS

func TestGlobalEmbeddedResourceManagerEmbed(t *testing.T) {
	Convey("GlobalEmbeddedResourceManagerEmbed", t, func() {
		m, err := NewGlobalEmbeddedResourceManagerEmbed(NewGlobalEmbeddedResourceManagerEmbedOptions{
			EmbedFS:                               testEmbedFS,
			EmbedFSRoot:                           "testdata/embed/a",
			ManifestFilenameRelativeToEmbedFSRoot: "manifest.json",
		})
		So(err, ShouldBeNil)

		Convey("basic AssetName() and Open()", func() {
			name, err := m.AssetName("index.js")
			So(err, ShouldBeNil)
			So(name, ShouldEqual, "index.deadbeef.js")

			f, err := m.Open(name)
			So(err, ShouldBeNil)
			defer f.Close()

			contents, err := io.ReadAll(f)
			So(err, ShouldBeNil)
			So(string(contents), ShouldEqual, "console.log(\"hello\");\n")
		})

		Convey("unknown asset name", func() {
			_, err := m.AssetName("nonsense")
			So(err, ShouldBeError, "specified resource is not configured")
		})

		Convey("reading parent directory", func() {
			_, err := m.Open("../evil")
			So(errors.Is(err, fs.ErrInvalid), ShouldBeTrue)
		})

		Convey("reading non-existent file", func() {
			_, err := m.Open("notfound")
			So(errors.Is(err, fs.ErrNotExist), ShouldBeTrue)
		})
	})
}
