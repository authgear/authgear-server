package ldap

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Provider struct {
	Store *Store
	Clock clock.Clock
}
