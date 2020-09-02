package mail

import (
	"errors"

	"github.com/go-gomail/gomail"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var ErrMissingSMTPConfiguration = errors.New("mail: configuration is missing")

type SendOptions struct {
	Sender    string
	ReplyTo   string
	Subject   string
	Recipient string
	TextBody  string
	HTMLBody  string
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("mail-sender")} }

type Sender struct {
	Logger       Logger
	DevMode      config.DevMode
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

func (s *Sender) Send(opts SendOptions) (err error) {
	if s.DevMode {
		s.Logger.
			WithField("recipient", opts.Recipient).
			WithField("body", opts.TextBody).
			WithField("sender", opts.Sender).
			WithField("subject", opts.Subject).
			WithField("reply_to", opts.ReplyTo).
			Warn("skip sending email in development mode")
		return nil
	}

	if s.GomailDialer == nil {
		err = ErrMissingSMTPConfiguration
		return
	}

	message := gomail.NewMessage()

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
	message.SetHeader("Reply-To", opts.ReplyTo)
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
