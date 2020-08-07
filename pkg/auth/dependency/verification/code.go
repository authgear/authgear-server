package verification

import (
	"time"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/otp"
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
	ExpireAt    time.Time `json:"expire_at"`
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
	}
}
