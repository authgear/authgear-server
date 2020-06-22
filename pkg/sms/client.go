package sms

import (
	"github.com/skygeario/skygear-server/pkg/auth/config"
)

type SendOptions struct {
	MessageConfig config.SMSMessageConfig
	To            string
	Body          string
}

type Client interface {
	Send(opts SendOptions) error
}
