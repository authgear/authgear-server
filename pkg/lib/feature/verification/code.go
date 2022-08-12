package verification

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
)

type CodeKey struct {
	WebSessionID string
	LoginIDType  string
	LoginID      string
}

type Code struct {
	UserID       string `json:"user_id"`
	IdentityID   string `json:"identity_id"`
	IdentityType string `json:"identity_type"`

	LoginIDType string    `json:"login_id_type"`
	LoginID     string    `json:"login_id"`
	Code        string    `json:"code"`
	ExpireAt    time.Time `json:"expire_at"`

	WebSessionID string `json:"web_session_id"`

	RequestedByUser bool `json:"requested_by_user"`
}

func (c *Code) CodeKey() *CodeKey {
	return &CodeKey{
		WebSessionID: c.WebSessionID,
		LoginIDType:  c.LoginIDType,
		LoginID:      c.LoginID,
	}
}

func (c *Code) SendResult() *otp.CodeSendResult {
	var channel string
	switch model.LoginIDKeyType(c.LoginIDType) {
	case model.LoginIDKeyTypeEmail:
		channel = string(model.AuthenticatorOOBChannelEmail)
	case model.LoginIDKeyTypePhone:
		channel = string(model.AuthenticatorOOBChannelSMS)
	default:
		panic("verification: unsupported login ID type: " + c.LoginIDType)
	}

	return &otp.CodeSendResult{
		Target:     c.LoginID,
		Channel:    channel,
		CodeLength: len(c.Code),
	}
}
