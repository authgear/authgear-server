package log

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLogHooks(t *testing.T) {
	Convey("default log hook", t, func() {
		h := NewDefaultMaskLogHook()

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

	Convey("secret log hook", t, func() {
		h := NewSecretMaskLogHook(&config.SecretConfig{
			Secrets: []config.SecretItem{
				{
					Key: config.DatabaseCredentialsKey,
					Data: &config.DatabaseCredentials{
						DatabaseURL:    "postgres://user:password@localhost:5432",
						DatabaseSchema: "public",
					},
				},
			},
		})
		Convey("should mask secret values", func() {
			e := &logrus.Entry{
				Message: "logged in",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"err": "cannot connect to postgres://user:password@localhost:5432",
				},
			}
			err := h.Fire(e)

			So(err, ShouldBeNil)
			So(e, ShouldResemble, &logrus.Entry{
				Message: "logged in",
				Level:   logrus.ErrorLevel,
				Data: logrus.Fields{
					"err": "cannot connect to ********",
				},
			})
		})
	})
}
