package verification

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	// SendCooldownSeconds is 60 seconds.
	SendCooldownSeconds = 60
)

type Code struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	IdentityID   string `json:"identity_id"`
	IdentityType string `json:"identity_type"`

	LoginIDType string    `json:"login_id_type"`
	LoginID     string    `json:"login_id"`
	Code        string    `json:"code"`
	SentAt      time.Time `json:"sent_at"`
	ExpireAt    time.Time `json:"expire_at"`

	WebSessionID string `json:"web_session_id"`

	RequestedByUser bool `json:"requested_by_user"`
}

func (c *Code) SendResult() *otp.CodeSendResult {
	var channel string
	switch config.LoginIDKeyType(c.LoginIDType) {
	case config.LoginIDKeyTypeEmail:
		channel = string(authn.AuthenticatorOOBChannelEmail)
	case config.LoginIDKeyTypePhone:
		channel = string(authn.AuthenticatorOOBChannelSMS)
	default:
		panic("verification: unsupported login ID type: " + c.LoginIDType)
	}

	return &otp.CodeSendResult{
		Target:       c.LoginID,
		Channel:      channel,
		CodeLength:   len(c.Code),
		SendCooldown: SendCooldownSeconds,
		SentAt:       c.SentAt,
	}
}

func NewCodeID() string {
	code := rand.StringWithAlphabet(16, base32.Alphabet, rand.SecureRand)
	return code
}
