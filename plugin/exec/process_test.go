package exec

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRun(t *testing.T) {
	Convey("test args and stdout", t, func() {
		transport := execTransport{
			Path: "/bin/echo",
			Args: []string{},
		}

		Convey("init", func() {
			out, err := transport.RunInit()
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "init")
		})

		Convey("op", func() {
			out, err := transport.RunLambda("hello:world", []byte{})
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "op hello:world")
		})

		Convey("handler", func() {
			out, err := transport.RunHandler("hello:world", []byte{})
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "handler hello:world")
		})

		Convey("hook", func() {
			out, err := transport.RunHook("note", "beforeSave", []byte{})
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "hook note:beforeSave")
		})
	})

	Convey("test stdin", t, func() {
		transport := execTransport{
			Path: "/bin/sh",
			Args: []string{"-c", `"cat"`},
		}

		Convey("init", func() {
			out, err := transport.RunInit()
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "")
		})

		Convey("op", func() {
			out, err := transport.RunLambda("hello:world", []byte("hello world"))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "hello world")
		})

		Convey("handler", func() {
			out, err := transport.RunHandler("hello:world", []byte("hello world"))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "hello world")
		})

		Convey("hook", func() {
			out, err := transport.RunHook("note", "beforeSave", []byte("hello world"))
			So(err, ShouldBeNil)
			So(string(out), ShouldEqual, "hello world")
		})
	})

	Convey("test exec error", t, func() {
		Convey("file not found", func() {
			transport := execTransport{
				Path: "/tmp/nonexistent",
				Args: []string{},
			}

			_, err := transport.RunInit()
			So(err, ShouldNotBeNil)
		})

		Convey("not executable", func() {
			transport := execTransport{
				Path: "/dev/null",
				Args: []string{},
			}

			_, err := transport.RunInit()
			So(err, ShouldNotBeNil)
		})

		Convey("return false", func() {
			transport := execTransport{
				Path: "/bin/false",
				Args: []string{},
			}

			_, err := transport.RunInit()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestFactory(t *testing.T) {
	Convey("test factory", t, func() {
		factory := execTransportFactory{}
		transport := factory.Open("/bin/echo", []string{"plugin"})

		So(transport, ShouldHaveSameTypeAs, execTransport{})
		So(transport.(execTransport).Path, ShouldResemble, "/bin/echo")
		So(transport.(execTransport).Args, ShouldResemble, []string{"plugin"})
	})
}
