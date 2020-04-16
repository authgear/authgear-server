package mail

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Sender interface {
	Send(opts SendOptions) error
}

type SendOptions struct {
	MessageConfig config.EmailMessageConfiguration
	Recipient     string
	TextBody      string
	HTMLBody      string
}
