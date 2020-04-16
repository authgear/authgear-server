package sms

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type SendOptions struct {
	MessageConfig config.SMSMessageConfiguration
	To            string
	Body          string
}

type Client interface {
	Send(opts SendOptions) error
}
