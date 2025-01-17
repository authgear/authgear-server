package web

import (
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGlobalEmbeddedResourceManagerWorkdir(t *testing.T) {
	Convey("GlobalEmbeddedResourceManagerWorkdir", t, func() {
		Convey("should throw error if resource directory does not exist", func() {
			m, err := NewGlobalEmbeddedResourceManagerWorkdir(&globalEmbeddedResourceManagerManifest{
				ResourceDir: "testdata/123/generated",
				Name:        "test.json",
			})
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
			So(m, ShouldBeNil)
		})

		Convey("should load manifest content after manager created", func() {
			m, err := NewGlobalEmbeddedResourceManagerWorkdir(&globalEmbeddedResourceManagerManifest{
				ResourceDir: "testdata/resources/authgear/generated",
				Name:        "manifest.json",
			})
			So(err, ShouldBeNil)
			defer m.close()

			manifestContext := m.getManifestContext()

			So(manifestContext, ShouldResemble, &globalEmbeddedResourceManagerManifestContext{
				Content: map[string]string{"test.js": "test.12345678.js"},
			})
			So(err, ShouldBeNil)
		})

		Convey("should reload manifest with any changes", func() {
			m, err := NewGlobalEmbeddedResourceManagerWorkdir(&globalEmbeddedResourceManagerManifest{
				ResourceDir: "testdata/resources/authgear/generated",
				Name:        "manifest.json",
			})
			So(err, ShouldBeNil)
			defer m.close()

			filePath := "testdata/resources/authgear/generated/manifest.json"

			f, err := os.Open(filePath)
			So(err, ShouldBeNil)
			originalContent, err := io.ReadAll(f)
			So(err, ShouldBeNil)
			defer f.Close()

			recoverFileContent := func() {
				// nolint:gosec
				_ = os.WriteFile(
					filePath,
					originalContent,
					0666,
				)
			}

			defer recoverFileContent()

			// nolint:gosec
			err = os.WriteFile(
				filePath,
				[]byte(`{"anotherTest.js": "anotherTest.12345678.js"}`),
				0666,
			)
			So(err, ShouldBeNil)
			runtime.Gosched()
			time.Sleep(500 * time.Millisecond)

			manifestContext := m.getManifestContext()
			So(manifestContext, ShouldResemble, &globalEmbeddedResourceManagerManifestContext{
				Content: map[string]string{"anotherTest.js": "anotherTest.12345678.js"},
			})
		})

		Convey("should return asset file name and prefix by key", func() {
			m, err := NewGlobalEmbeddedResourceManagerWorkdir(&globalEmbeddedResourceManagerManifest{
				ResourceDir: "testdata/resources/authgear/generated",
				Name:        "manifest.json",
			})
			So(err, ShouldBeNil)
			defer m.close()

			// if key exists
			assetFileName, err := m.AssetName("test.js")
			So(err, ShouldBeNil)
			So(assetFileName, ShouldEqual, "test.12345678.js")

			// if key does not exist
			assetFileName, err = m.AssetName("test123.js")
			So(err, ShouldBeError, "specified resource is not configured")
			So(assetFileName, ShouldBeEmpty)
		})
	})
}
