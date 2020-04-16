package oob

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type Authenticator struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	Channel   authn.AuthenticatorOOBChannel
	Phone     string
	Email     string
}
