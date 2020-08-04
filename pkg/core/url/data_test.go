package url

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

		w.Write([]byte("Hello, World"))
		w.Close()

		So(string(buf.Bytes()), ShouldResemble, "data:text/plain;base64,SGVsbG8sIFdvcmxk")
	})
}
