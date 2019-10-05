package cloudstorage

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRewriteHeaderName(t *testing.T) {
	Convey("RewriteHeaderName", t, func() {
		header := http.Header{
			"Content-Type":        {"image/png"},
			"Content-Disposition": {`attachment; filename="file.png"`},
		}
		mapping := map[string]string{
			"content-disposition": "x-app-content-disposition",
		}
		actual := RewriteHeaderName(header, mapping)
		expected := http.Header{
			"Content-Type":              {"image/png"},
			"X-App-Content-Disposition": {`attachment; filename="file.png"`},
		}
		So(actual, ShouldResemble, expected)
	})
}
