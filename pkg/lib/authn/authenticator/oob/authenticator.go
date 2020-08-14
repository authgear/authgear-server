package oob

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn"
)

const (
	// OOBOTPValidDuration is 20 minutes according to the suggestion in
	// https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html#step-3-send-a-token-over-a-side-channel
	OOBOTPValidDuration time.Duration = 20 * time.Minute
	// OOBOTPSendCooldownSeconds is 60 seconds.
	OOBOTPSendCooldownSeconds = 60
)

type Authenticator struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	Channel   authn.AuthenticatorOOBChannel
	Phone     string
	Email     string
	Tag       []string
}
