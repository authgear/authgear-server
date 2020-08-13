package log

import (
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLogFormatHook(t *testing.T) {
	Convey("LogFormatHook", t, func() {
		h := FormatHook{Mask: "********"}
		h.MaskPatterns = []MaskPattern{NewPlainMaskPattern("SECRET")}

		Convey("should mask message", func() {
			e := &logrus.Entry{
				Message: "Test SECRET",
				Level:   logrus.ErrorLevel,
			}
			err := h.Fire(e)

			So(err, ShouldBeNil)
			So(e, ShouldResemble, &logrus.Entry{
				Message: "Test ********",
				Level:   logrus.ErrorLevel,
			})
		})
		Convey("should mask string in data", func() {
			e := &logrus.Entry{
				Message: "Test SECRET",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"test": "Get SECRETDATA",
				},
			}
			err := h.Fire(e)

			So(err, ShouldBeNil)
			So(e, ShouldResemble, &logrus.Entry{
				Message: "Test ********",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"test": "Get ********DATA",
				},
			})
		})
		Convey("should mask complex value in data", func() {
			e := &logrus.Entry{
				Message: "Test SECRET",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"app": struct{ Name string }{
						Name: "SECRET app",
					},
				},
			}
			err := h.Fire(e)

			So(err, ShouldBeNil)
			So(e, ShouldResemble, &logrus.Entry{
				Message: "Test ********",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"app": map[string]interface{}{
						"Name": "******** app",
					},
				},
			})
		})
	})
}
