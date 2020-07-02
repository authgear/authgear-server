package spec

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/mail"
	"github.com/authgear/authgear-server/pkg/sms"
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
	SendMessagesTaskName = "SendMessagesTask"
)

type SendMessagesTaskParam struct {
	EmailMessages []mail.SendOptions
	SMSMessages   []sms.SendOptions
}
