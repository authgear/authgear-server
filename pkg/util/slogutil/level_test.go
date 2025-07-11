package slogutil

import (
	"log/slog"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseLevel(t *testing.T) {
	Convey("ParseLevel", t, func() {
		Convey("should parse debug level", func() {
			So(ParseLevel("debug"), ShouldEqual, slog.LevelDebug)
			So(ParseLevel("DEBUG"), ShouldEqual, slog.LevelDebug)
			So(ParseLevel("Debug"), ShouldEqual, slog.LevelDebug)
		})

		Convey("should parse info level", func() {
			So(ParseLevel("info"), ShouldEqual, slog.LevelInfo)
			So(ParseLevel("INFO"), ShouldEqual, slog.LevelInfo)
			So(ParseLevel("Info"), ShouldEqual, slog.LevelInfo)
		})

		Convey("should parse warn level", func() {
			So(ParseLevel("warn"), ShouldEqual, slog.LevelWarn)
			So(ParseLevel("WARN"), ShouldEqual, slog.LevelWarn)
			So(ParseLevel("Warn"), ShouldEqual, slog.LevelWarn)
		})

		Convey("should parse warning level", func() {
			So(ParseLevel("warning"), ShouldEqual, slog.LevelWarn)
			So(ParseLevel("WARNING"), ShouldEqual, slog.LevelWarn)
			So(ParseLevel("Warning"), ShouldEqual, slog.LevelWarn)
		})

		Convey("should parse error level", func() {
			So(ParseLevel("error"), ShouldEqual, slog.LevelError)
			So(ParseLevel("ERROR"), ShouldEqual, slog.LevelError)
			So(ParseLevel("Error"), ShouldEqual, slog.LevelError)
		})

		Convey("should default to warn for unknown levels", func() {
			So(ParseLevel("unknown"), ShouldEqual, slog.LevelWarn)
			So(ParseLevel(""), ShouldEqual, slog.LevelWarn)
			So(ParseLevel("trace"), ShouldEqual, slog.LevelWarn)
			So(ParseLevel("fatal"), ShouldEqual, slog.LevelWarn)
		})
	})
}
