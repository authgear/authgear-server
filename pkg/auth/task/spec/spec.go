package spec

import (
	"errors"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/model"
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
	// VerifyCodeSendTaskName provides the name for submiting VerifyCodeSendTask
	VerifyCodeSendTaskName = "VerifyCodeSendTask"
)

type VerifyCodeSendTaskParam struct {
	URLPrefix *url.URL
	LoginID   string
	UserID    string
}

const (
	// WelcomeEmailSendTaskName provides the name for submiting WelcomeEmailSendTask
	WelcomeEmailSendTaskName = "WelcomeEmailSendTask"
)

type WelcomeEmailSendTaskParam struct {
	URLPrefix *url.URL
	Email     string
	User      model.User
}
