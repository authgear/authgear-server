package mail

import (
	"errors"

	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func NewDialer(c config.SMTPConfiguration) (dialer *gomail.Dialer) {
	if c.Host == "" {
		panic(errors.New("mail server is not configured"))
	}

	dialer = gomail.NewPlainDialer(c.Host, c.Port, c.Login, c.Password)
	switch c.Mode {
	case config.SMTPModeNormal:
		// gomail will infer according to port
	case config.SMTPModeSSL:
		dialer.SSL = true
	}
	return
}
