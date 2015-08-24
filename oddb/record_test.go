package oddb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRecord(t *testing.T) {
	Convey("Set transient field", t, func() {
		note0 := Record{
			ID: NewRecordID("note", "0"),
			Transient: Data{
				"content": "hello world",
			},
		}

		So(note0.Get("content"), ShouldBeNil)
		So(note0.Get("_transient"), ShouldResemble, Data{
			"content": "hello world",
		})
		So(note0.Get("_transient_content"), ShouldEqual, "hello world")
	})

	Convey("Set transient field", t, func() {
		note0 := Record{
			ID: NewRecordID("note", "0"),
		}

		note0.Set("_transient", Data{
			"content": "hello world",
		})

		So(note0.Data["content"], ShouldBeNil)
		So(note0.Transient, ShouldResemble, Data{
			"content": "hello world",
		})
	})

	Convey("Set individual transient field", t, func() {
		note0 := Record{
			ID: NewRecordID("note", "0"),
			Transient: Data{
				"existing": "should be here",
			},
		}

		note0.Set("_transient_content", "hello world")

		So(note0.Data["content"], ShouldBeNil)
		So(note0.Transient, ShouldResemble, Data{
			"content":  "hello world",
			"existing": "should be here",
		})
	})
}
