package oob

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const (
	// OOBOTPValidDuration is duration.UserInteraction.
	OOBOTPValidDuration = duration.UserInteraction
	// OOBOTPSendCooldownSeconds is 60 seconds.
	OOBOTPSendCooldownSeconds = 60
)

type Authenticator struct {
	ID                   string
	IsDefault            bool
	Kind                 string
	UserID               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	OOBAuthenticatorType authn.AuthenticatorType
	Phone                string
	Email                string
}
