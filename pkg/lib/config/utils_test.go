package config_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestSecretLogHook(t *testing.T) {
	Convey("secret log hook", t, func() {
		h := config.NewSecretMaskLogHook(&config.SecretConfig{
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
