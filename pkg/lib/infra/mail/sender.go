package mail

import (
	"errors"
	"fmt"
	netmail "net/mail"

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

// SetFromHeader sets the RFC 5322 From header so that only the display
// name is RFC 2047 encoded, not the angle-addr literal.
//
// gomail's SetHeader("From", sender) encodes the whole value as one encoded
// word when it contains non-ASCII, turning
//
//	"範例 <noreply@example.com>"
//
// into
//
//	=?UTF-8?q?=E7=AF=84=E4=BE=8B_<noreply@example.com>?=
//
// which is invalid — RFC 5322 §3.4 requires the addr-spec to appear outside
// any encoded word. SetAddressHeader encodes only the display name, producing:
//
//	=?UTF-8?q?=E7=AF=84=E4=BE=8B?= <noreply@example.com>
//
// See RFC 5322 §3.4 (https://datatracker.ietf.org/doc/html/rfc5322#section-3.4) and
// RFC 2047 (https://datatracker.ietf.org/doc/html/rfc2047) for the encoding rules.
// RFC 6532 (https://datatracker.ietf.org/doc/html/rfc6532) would allow raw UTF-8
// in headers, but requires the SMTPUTF8 extension (RFC 6531) on the relay; RFC 2047
// encoded words are used here for compatibility with all SMTP servers.
func SetFromHeader(message *gomail.Message, sender string) error {
	addr, err := netmail.ParseAddress(sender)
	if err != nil {
		return fmt.Errorf("invalid sender address %q: %w", sender, err)
	}
	message.SetAddressHeader("From", addr.Address, addr.Name)
	return nil
}

func (s *Sender) applyFrom(opts *SendOptions, message *gomail.Message) error {
	return SetFromHeader(message, opts.Sender)
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
