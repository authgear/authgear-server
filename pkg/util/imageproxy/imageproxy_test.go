package imageproxy

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func extractKey(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	return fmt.Sprintf("%s/%s", parts[2], parts[3])
}

func TestGCPGCSDirector(t *testing.T) {
	Convey("GCPGCSDirector", t, func() {
		d := GCPGCSDirector{
			ExtractKey: extractKey,
			BucketName: "example",
		}

		r, _ := http.NewRequest("GET", "/_images/app/objectid/profile", nil)
		director := d.Director
		director(r)

		So(r.URL.String(), ShouldEqual, "https://storage.googleapis.com/example/app/objectid")
		So(r.Host, ShouldEqual, "storage.googleapis.com")
	})
}

func TestAWSS3Director(t *testing.T) {
	Convey("AWSS3Director", t, func() {
		d := AWSS3Director{
			ExtractKey: extractKey,
			BucketName: "example",
			Region:     "us-east-2",
		}

		r, _ := http.NewRequest("GET", "/_images/app/objectid/profile", nil)
		director := d.Director
		director(r)

		So(r.URL.String(), ShouldEqual, "https://example.s3.us-east-2.amazonaws.com/app/objectid")
		So(r.Host, ShouldEqual, "example.s3.us-east-2.amazonaws.com")
	})
}

func TestAzureBlobStorageDirector(t *testing.T) {
	Convey("AzureBlobStorageDirector", t, func() {
		d := AzureBlobStorageDirector{
			ExtractKey:     extractKey,
			StorageAccount: "myaccount",
			Container:      "mycontainer",
		}

		r, _ := http.NewRequest("GET", "/_images/app/objectid/profile", nil)
		director := d.Director
		director(r)

		So(r.URL.String(), ShouldEqual, "https://myaccount.blob.core.windows.net/mycontainer/app/objectid")
		So(r.Host, ShouldEqual, "myaccount.blob.core.windows.net")
	})
}
