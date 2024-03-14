package password

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	gomock "github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func TestValidateCurrentPassword(t *testing.T) {
	now := time.Date(2017, 11, 4, 0, 0, 0, 0, time.UTC)

	ctrl := gomock.NewController(t)
	mockAuthencatorStore := NewMockAuthenticatorStore(ctrl)

	mockAuthencatorStore.EXPECT().List("chima").Return([]*authenticator.Password{
		{
			ID:           "0",
			UserID:       "chima",
			PasswordHash: []byte("$2a$10$EazYxG5cUdf99wGXDU1fguNxvCe7xQLEgr/Ay6VS9fkkVjHZtpJfl"), // random hash
			UpdatedAt:    now.Add(-time.Hour * 24 * 90),
		},
		{
			ID:           "1",
			UserID:       "chima",
			PasswordHash: []byte("$2a$10$EazYxG5cUdf99wGXDU1fguNxvCe7xQLEgr/Ay6VS9fkkVjHZtpJfm"), // "chima"
			UpdatedAt:    now.Add(-time.Hour * 24 * 30),
		},
	}, nil)
	mockAuthencatorStore.EXPECT().List("faseng").Return([]*authenticator.Password{
		{
			ID:           "2",
			UserID:       "faseng",
			PasswordHash: []byte("$2a$10$8Z0zqmCZ3pZUlvLD8lN.B.ecN7MX8uVcZooPUFnCcB8tWR6diVc1a"), // "faseng"
			UpdatedAt:    now.Add(-time.Hour * 24 * 30),
		},
	}, nil)
	mockAuthencatorStore.EXPECT().List("coffee").Return([]*authenticator.Password{
		{
			ID:           "3",
			UserID:       "coffee",
			PasswordHash: []byte("$2a$10$qzmi8TkYosj66xHvc9EfEulKjGoZswJSyNVEmmbLDxNGP/lMm6UXC"), // "coffee"
			UpdatedAt:    now.Add(-time.Hour * 24 * 29),
		},
	}, nil)
	mockAuthencatorStore.EXPECT().List("milktea").Return([]*authenticator.Password{}, nil)

	test := func(pe *Expiry, userID string, password string, expected string) {
		err := pe.Validate(userID, password)
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
			AuthenticatorStore:         mockAuthencatorStore,
			Clock:                      clock.NewMockClockAtTime(now),
		}

		test(pc, "chima", "chima", `
		{
			"name": "Invalid",
			"reason": "PasswordExpiryForceChange",
			"message": "password expired",
			"code": 400
		}
		`)

		test(pc, "faseng", "faseng", `
		{
			"name": "Invalid",
			"reason": "PasswordExpiryForceChange",
			"message": "password expired",
			"code": 400
		}
		`)
		test(pc, "coffee", "coffee", "")
		test(pc, "milktea", "milktea", "")
	})
}
