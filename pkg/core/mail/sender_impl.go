package mail

import (
	"context"
	"errors"

	"github.com/go-gomail/gomail"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/intl"
)

var ErrMissingSMTPConfiguration = errors.New("mail: configuration is missing")

type senderImpl struct {
	GomailDialer *gomail.Dialer
	Context      context.Context
}

func NewSender(ctx context.Context, c *config.SMTPConfiguration) Sender {
	var dialer *gomail.Dialer
	if c != nil && c.IsValid() {
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
		Context:      ctx,
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

func (s *senderImpl) applyFrom(opts *SendOptions, message *gomail.Message) error {
	tags := intl.GetPreferredLanguageTags(s.Context)
	sender := intl.LocalizeOIDCStringMap(tags, opts.MessageConfig, "sender")
	if sender == "" {
		return errors.New("mail: sender address is missing")
	}
	message.SetHeader("From", sender)
	return nil
}

func applyTo(opts *SendOptions, message *gomail.Message) error {
	if opts.Recipient == "" {
		return errors.New("mail: recipient address is missing")
	}

	message.SetHeader("To", opts.Recipient)
	return nil
}

func (s *senderImpl) applyReplyTo(opts *SendOptions, message *gomail.Message) error {
	tags := intl.GetPreferredLanguageTags(s.Context)
	replyTo := intl.LocalizeOIDCStringMap(tags, opts.MessageConfig, "reply_to")
	if replyTo == "" {
		return nil
	}

	message.SetHeader("Reply-To", replyTo)
	return nil
}

func (s *senderImpl) applySubject(opts *SendOptions, message *gomail.Message) error {
	tags := intl.GetPreferredLanguageTags(s.Context)
	subject := intl.LocalizeOIDCStringMap(tags, opts.MessageConfig, "subject")
	if subject == "" {
		return errors.New("mail: subject is missing")
	}

	message.SetHeader("Subject", subject)
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
