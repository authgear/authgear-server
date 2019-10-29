package mail

import (
	"errors"

	"github.com/go-gomail/gomail"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

var ErrMissingSMTPConfiguration = errors.New("mail: configuration is missing")

type senderImpl struct {
	GomailDialer *gomail.Dialer
}

func NewSender(c config.SMTPConfiguration) Sender {
	var dialer *gomail.Dialer
	if c.IsValid() {
		dialer = gomail.NewPlainDialer(c.Host, c.Port, c.Login, c.Password)
	}
	switch c.Mode {
	case config.SMTPModeNormal:
		// gomail will infer according to port
	case config.SMTPModeSSL:
		dialer.SSL = true
	}
	return &senderImpl{
		GomailDialer: dialer,
	}
}

type updateGomailMessageFunc func(opts *SendOptions, msg *gomail.Message) error

func (s *senderImpl) Send(opts SendOptions) (err error) {
	if s.GomailDialer == nil {
		err = ErrMissingSMTPConfiguration
		return
	}

	message := gomail.NewMessage()

	funcs := []updateGomailMessageFunc{
		applyFrom,
		applyTo,
		applyReplyTo,
		applySubject,
		applyTextBody,
		applyHTMLBody,
	}

	for _, f := range funcs {
		if err = f(&opts, message); err != nil {
			return
		}
	}

	err = s.GomailDialer.DialAndSend(message)
	if err != nil {
		return err
	}

	return nil
}

func applyFrom(opts *SendOptions, message *gomail.Message) error {
	if opts.Sender == "" {
		return errors.New("mail: sender address is missing")
	}

	message.SetHeader("From", opts.Sender)
	return nil
}

func applyTo(opts *SendOptions, message *gomail.Message) error {
	if opts.Recipient == "" {
		return errors.New("mail: recipient address is missing")
	}

	message.SetHeader("To", opts.Recipient)
	return nil
}

func applyReplyTo(opts *SendOptions, message *gomail.Message) error {
	if opts.ReplyTo == "" {
		return nil
	}

	message.SetHeader("Reply-To", opts.ReplyTo)
	return nil
}

func applySubject(opts *SendOptions, message *gomail.Message) error {
	if opts.Subject == "" {
		return errors.New("mail: subject is missing")
	}

	message.SetHeader("Subject", opts.Subject)
	return nil
}

func applyTextBody(opts *SendOptions, message *gomail.Message) error {
	if opts.TextBody == "" {
		return errors.New("mail: text body is missing")
	}

	message.SetBody("text/plain", opts.TextBody)
	return nil
}

func applyHTMLBody(opts *SendOptions, message *gomail.Message) error {
	if opts.HTMLBody == "" {
		return nil
	}

	message.AddAlternative("text/html", opts.HTMLBody)
	return nil
}
