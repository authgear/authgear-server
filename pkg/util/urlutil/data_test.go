package urlutil

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDataURIWriter(t *testing.T) {
	Convey("DataURIWriter", t, func() {
		var buf bytes.Buffer
		w, err := DataURIWriter("text/plain", &buf)
		So(err, ShouldBeNil)

		_, _ = w.Write([]byte("Hello, World"))
		_ = w.Close()

		So(buf.String(), ShouldResemble, "data:text/plain;base64,SGVsbG8sIFdvcmxk")
	})
}
