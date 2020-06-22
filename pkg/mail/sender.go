package mail

import (
	"github.com/skygeario/skygear-server/pkg/auth/config"
)

type Sender interface {
	Send(opts SendOptions) error
}

type SendOptions struct {
	MessageConfig config.EmailMessageConfig
	Recipient     string
	TextBody      string
	HTMLBody      string
}
