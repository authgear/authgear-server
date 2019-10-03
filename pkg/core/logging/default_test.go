package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDefaultFormatter(t *testing.T) {
	Convey("DefaultMaskedTextFormatter", t, func() {
		fmt := NewDefaultMaskedTextFormatter([]string{"SECRET"})

		Convey("should mask sensitive strings", func() {
			buf, err := fmt.Format(&logrus.Entry{
				Message: "Test SECRET",
				Level:   logrus.ErrorLevel,
			})

			So(err, ShouldBeNil)
			So(string(buf), ShouldEqual, `time="0001-01-01T00:00:00Z" level=error msg="Test ********"`+"\n")
		})
		Convey("should mask JWTs", func() {
			buf, err := fmt.Format(&logrus.Entry{
				Message: "logged in",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"authz": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.MiwK31U8C6MNcuYw7EMsAtjioTwG8oOgG0swJeH738k",
				},
			})

			So(err, ShouldBeNil)
			So(string(buf), ShouldEqual, `time="0001-01-01T00:00:00Z" level=error msg="logged in" authz="Bearer ********"`+"\n")
		})
		Convey("should mask session tokens", func() {
			buf, err := fmt.Format(&logrus.Entry{
				Message: "refreshing token",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"tokens": struct {
						Access  string
						Refresh string
					}{
						Access:  "54448008-84f9-4413-8d61-036f0a6d7878.dyHMxL8P1N7l3amK2sKBKCSPLzhiwTEA",
						Refresh: "54448008-84f9-4413-8d61-036f0a6d7878.5EFoSwEoc0mRE7fNGvPNqUjWc1VlY5vG",
					},
				},
			})

			So(err, ShouldBeNil)
			So(string(buf), ShouldEqual, `time="0001-01-01T00:00:00Z" level=error msg="refreshing token" tokens="{******** ********}"`+"\n")
		})
	})
}
