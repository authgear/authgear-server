package mail

import (
	"errors"

	"gopkg.in/gomail.v2"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

var ErrNoAvailableSMTPConfiguration = apierrors.InternalError.WithReason("NoAvailableSMTPConfiguration").New("no available SMTP configuration")

type SendOptions struct {
	Sender    string
	ReplyTo   string
	Subject   string
	Recipient string
	TextBody  string
	HTMLBody  string
}

type Sender struct {
	GomailDialer *gomail.Dialer
}

func NewGomailDialer(smtp *config.SMTPServerCredentials) *gomail.Dialer {
	if smtp != nil {
		dialer := gomail.NewDialer(smtp.Host, smtp.Port, smtp.Username, smtp.Password)
		switch smtp.Mode {
		case config.SMTPModeNormal:
			// gomail will infer according to port
		case config.SMTPModeSSL:
			dialer.SSL = true
		}
		return dialer
	}
	return nil
}

type updateGomailMessageFunc func(opts *SendOptions, msg *gomail.Message) error

func (s *Sender) PrepareMessage(opts SendOptions) (message *gomail.Message, err error) {
	if s.GomailDialer == nil {
		err = ErrNoAvailableSMTPConfiguration
		return
	}

	message = gomail.NewMessage()

	funcs := []updateGomailMessageFunc{
		s.applyFrom,
		applyTo,
		s.applyReplyTo,
		s.applySubject,
		applyTextBody,
		applyHTMLBody,
	}

	for _, f := range funcs {
		if err = f(&opts, message); err != nil {
			return
		}
	}

	return
}

func (s *Sender) Send(message *gomail.Message) (err error) {
	if s.GomailDialer == nil {
		err = ErrNoAvailableSMTPConfiguration
		return
	}

	err = s.GomailDialer.DialAndSend(message)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sender) applyFrom(opts *SendOptions, message *gomail.Message) error {
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

func (s *Sender) applyReplyTo(opts *SendOptions, message *gomail.Message) error {
	if opts.ReplyTo != "" {
		message.SetHeader("Reply-To", opts.ReplyTo)
	}
	return nil
}

func (s *Sender) applySubject(opts *SendOptions, message *gomail.Message) error {
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
