package config_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestParseRateLimitsEnv(t *testing.T) {
	Convey("parse rate limits env string", t, func() {
		parse := func(s string) (config.RateLimitsEnvironmentConfigEntry, error) {
			var e config.RateLimitsEnvironmentConfigEntry
			err := e.Set(s)
			return e, err
		}

		e, err := parse("")
		So(err, ShouldBeNil)
		So(e, ShouldResemble, config.RateLimitsEnvironmentConfigEntry{Enabled: false})

		_, err = parse("1h")
		So(err, ShouldBeError, "invalid rate limit: 1h")

		_, err = parse("1.1/h")
		So(err, ShouldBeError, `invalid burst value: strconv.Atoi: parsing "1.1": invalid syntax`)

		_, err = parse("-1/h")
		So(err, ShouldBeError, "invalid burst value: -1")

		_, err = parse("1/h")
		So(err, ShouldBeError, `invalid period value: time: invalid duration "h"`)

		_, err = parse("1/-1h")
		So(err, ShouldBeError, `invalid period value: -1h0m0s`)

		e, err = parse("50/24h")
		So(err, ShouldBeNil)
		So(e, ShouldResemble, config.RateLimitsEnvironmentConfigEntry{
			Enabled: true,
			Period:  24 * time.Hour,
			Burst:   50,
		})
	})
}
