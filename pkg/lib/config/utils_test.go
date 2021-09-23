package config_test

import (
	"encoding/json"
	"testing"
	"time"

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

	Convey("test Date type", t, func() {
		Convey("test MarshalJSON", func() {
			t := config.Date(time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC))
			dateB, _ := json.Marshal(&t)
			So(string(dateB), ShouldResemble, `"2006-01-02"`)

			var tPtr *time.Time
			dateB, _ = json.Marshal(tPtr)
			So(string(dateB), ShouldResemble, `null`)
		})
	})
}
