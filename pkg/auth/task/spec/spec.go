package spec

import (
	"errors"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
)

const (
	// PwHousekeeperTaskName provides the name for submiting PwHousekeeperTask
	PwHousekeeperTaskName = "PwHousekeeperTask"
)

type PwHousekeeperTaskParam struct {
	AuthID string
}

func (p PwHousekeeperTaskParam) Validate() error {
	if p.AuthID == "" {
		return errors.New("missing user ID")
	}

	return nil
}

const (
	// TODO(verify): Remove VerifyCodeSendTask and use SendMessagesTask
	// VerifyCodeSendTaskName provides the name for submiting VerifyCodeSendTask
	VerifyCodeSendTaskName = "VerifyCodeSendTask"
)

type VerifyCodeSendTaskParam struct {
	URLPrefix *url.URL
	LoginID   string
	UserID    string
}

const (
	SendMessagesTaskName = "SendMessagesTask"
)

type SendMessagesTaskParam struct {
	EmailMessages []mail.SendOptions
	SMSMessages   []sms.SendOptions
}
