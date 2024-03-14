package password

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"

	"github.com/authgear/authgear-server/pkg/util/clock"
)

//go:generate mockgen -source=expiry.go -destination=expiry_mock_test.go -package password

type AuthenticatorStore interface {
	List(userID string) ([]*authenticator.Password, error)
}

type Expiry struct {
	ForceChangeEnabled         bool
	ForceChangeSinceLastUpdate config.DurationString
	AuthenticatorStore         AuthenticatorStore
	Clock                      clock.Clock
}

func (pe *Expiry) Validate(password, authID string) error {
	if authID == "" {
		return nil
	}

	authenticators, err := pe.AuthenticatorStore.List(authID)
	if err != nil {
		return err
	}

	for _, pa := range authenticators {
		if IsSamePassword(pa.PasswordHash, password) {
			if pe.ForceChangeEnabled {
				if !pa.UpdatedAt.Add(pe.ForceChangeSinceLastUpdate.Duration()).After(pe.Clock.NowUTC()) {
					return PasswordExpiryForceChange.New("password expired")
				}
			}
		}
	}
	return nil
}
