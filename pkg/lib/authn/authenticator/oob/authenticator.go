package oob

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
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
	OOBAuthenticatorType model.AuthenticatorType
	Phone                string
	Email                string
}
