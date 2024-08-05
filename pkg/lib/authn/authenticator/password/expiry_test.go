package password

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func TestValidateCurrentPassword(t *testing.T) {
	now := time.Date(2017, 11, 4, 0, 0, 0, 0, time.UTC)

	test := func(pe *Expiry, authenticator *authenticator.Password, expected string) {
		err := pe.Validate(authenticator)
		if err != nil {
			e := apierrors.AsAPIError(err)
			b, _ := json.Marshal(e)
			So(string(b), ShouldEqualJSON, expected)
		} else {
			So(expected, ShouldBeEmpty)
		}
	}

	Convey("validate password expiry", t, func(c C) {
		thresholdDays := config.DurationString("720h")

		pc := &Expiry{
			ForceChangeEnabled:         true,
			ForceChangeSinceLastUpdate: thresholdDays,
			Clock:                      clock.NewMockClockAtTime(now),
		}

		test(pc, &authenticator.Password{
			ID:        "1",
			UserID:    "chima",
			UpdatedAt: now.Add(-time.Hour * 24 * 30),
		}, `
		{
			"name": "Invalid",
			"reason": "PasswordExpiryForceChange",
			"message": "password expired",
			"code": 400
		}
		`)

		test(pc, &authenticator.Password{
			ID:        "2",
			UserID:    "faseng",
			UpdatedAt: now.Add(-time.Hour * 24 * 30),
		}, `
		{
			"name": "Invalid",
			"reason": "PasswordExpiryForceChange",
			"message": "password expired",
			"code": 400
		}
		`)

		test(pc, &authenticator.Password{
			ID:        "3",
			UserID:    "coffee",
			UpdatedAt: now.Add(-time.Hour * 24 * 29),
		}, "")
	})

	Convey("expire_after does not require expiry to be enabled", t, func() {
		pc := &Expiry{
			ForceChangeEnabled: false,
			Clock:              clock.NewMockClockAtTime(now),
		}

		expireAfter := now.Add(time.Millisecond * -1)
		test(pc, &authenticator.Password{
			ID:          "3",
			UserID:      "coffee",
			UpdatedAt:   now,
			ExpireAfter: &expireAfter,
		}, `
		{
			"name": "Invalid",
			"reason": "PasswordExpiryForceChange",
			"message": "password expired",
			"code": 400
		}
		`)
	})
}
