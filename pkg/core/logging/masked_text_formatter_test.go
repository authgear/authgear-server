package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMaskedTextFormatter(t *testing.T) {
	Convey("MaskedTextFormatter", t, func() {
		fmt := MaskedTextFormatter{Mask: "********"}
		fmt.Patterns = []MaskPattern{NewPlainMaskPattern("SECRET")}

		Convey("should mask message", func() {
			buf, err := fmt.Format(&logrus.Entry{
				Message: "Test SECRET",
				Level:   logrus.ErrorLevel,
			})

			So(err, ShouldBeNil)
			So(string(buf), ShouldEqual, `time="0001-01-01T00:00:00Z" level=error msg="Test ********"`+"\n")
		})
		Convey("should mask string in data", func() {
			buf, err := fmt.Format(&logrus.Entry{
				Message: "Test SECRET",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"test": "Get SECRETDATA",
				},
			})

			So(err, ShouldBeNil)
			So(string(buf), ShouldEqual, `time="0001-01-01T00:00:00Z" level=error msg="Test ********" test="Get ********DATA"`+"\n")
		})
		Convey("should mask complex value in data", func() {
			buf, err := fmt.Format(&logrus.Entry{
				Message: "Test SECRET",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"app": struct{ Name string }{
						Name: "SECRET app",
					},
				},
			})

			So(err, ShouldBeNil)
			So(string(buf), ShouldEqual, `time="0001-01-01T00:00:00Z" level=error msg="Test ********" app="{******** app}"`+"\n")
		})
	})
}
