package password

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"

	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Expiry struct {
	ForceChangeEnabled         bool
	ForceChangeSinceLastUpdate config.DurationString
	Clock                      clock.Clock
}

func (pe *Expiry) Validate(authenticator *authenticator.Password) error {
	if authenticator.ExpireAfter != nil && authenticator.ExpireAfter.Before(pe.Clock.NowUTC()) {
		return PasswordExpiryForceChange.New("password expired")
	}

	// Authenticator is already verified with given password prior to this call.
	if pe.ForceChangeEnabled {
		if !authenticator.UpdatedAt.Add(pe.ForceChangeSinceLastUpdate.Duration()).After(pe.Clock.NowUTC()) {
			return PasswordExpiryForceChange.New("password expired")
		}
	}
	return nil
}
