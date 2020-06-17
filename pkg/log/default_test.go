package log

import (
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDefaultFormatter(t *testing.T) {
	Convey("DefaultMaskedTextFormatter", t, func() {
		h := NewDefaultLogHook([]string{"SECRET"})

		Convey("should mask sensitive strings", func() {
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
		Convey("should mask JWTs", func() {
			e := &logrus.Entry{
				Message: "logged in",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"authz": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.MiwK31U8C6MNcuYw7EMsAtjioTwG8oOgG0swJeH738k",
				},
			}
			err := h.Fire(e)

			So(err, ShouldBeNil)
			So(e, ShouldResemble, &logrus.Entry{
				Message: "logged in",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"authz": "Bearer ********",
				},
			})
		})
		Convey("should mask session tokens", func() {
			e := &logrus.Entry{
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
			}
			err := h.Fire(e)

			So(err, ShouldBeNil)
			So(e, ShouldResemble, &logrus.Entry{
				Message: "refreshing token",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"tokens": map[string]interface{}{
						"Access":  "********",
						"Refresh": "********",
					},
				},
			})
		})
	})
}
