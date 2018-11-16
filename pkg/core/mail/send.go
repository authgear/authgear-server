package mail

import (
	"errors"

	"github.com/go-gomail/gomail"
)

type SendRequest struct {
	Dialer      *gomail.Dialer
	Sender      string
	SenderName  string
	Recipient   string
	Subject     string
	ReplyTo     string
	ReplyToName string
	TextBody    string
	HTMLBody    string
}

type updateMessageFunc func(*gomail.Message) error

func (r *SendRequest) Execute() (err error) {
	message := gomail.NewMessage()

	funcs := []updateMessageFunc{
		r.applyFrom,
		r.applyTo,
		r.applyReplyTo,
		r.applySubject,
		r.applyTextBody,
		r.applyHTMLBody,
	}

	for _, f := range funcs {
		if err = f(message); err != nil {
			return
		}
	}

	err = r.Dialer.DialAndSend(message)
	return
}

func (r *SendRequest) applyFrom(message *gomail.Message) error {
	if r.Sender == "" {
		return errors.New("sender address is missing")
	}

	if r.SenderName == "" {
		message.SetHeader("From", r.Sender)
	} else {
		message.SetAddressHeader("From", r.Sender, r.SenderName)
	}

	return nil
}

func (r *SendRequest) applyTo(message *gomail.Message) error {
	if r.Recipient == "" {
		return errors.New("recipient address is missing")
	}

	message.SetHeader("To", r.Recipient)
	return nil
}

func (r *SendRequest) applyReplyTo(message *gomail.Message) error {
	if r.ReplyTo == "" {
		return nil
	}

	if r.SenderName == "" {
		message.SetHeader("Reply-To", r.ReplyTo)
	} else {
		message.SetAddressHeader("Reply-To", r.ReplyTo, r.ReplyToName)
	}

	return nil
}

func (r *SendRequest) applySubject(message *gomail.Message) error {
	if r.Subject == "" {
		return errors.New("subject is missing")
	}

	message.SetHeader("Subject", r.Subject)
	return nil
}

func (r *SendRequest) applyTextBody(message *gomail.Message) error {
	if r.TextBody == "" {
		return errors.New("text body is missing")
	}

	message.SetBody("text/plain", r.TextBody)
	return nil
}

func (r *SendRequest) applyHTMLBody(message *gomail.Message) error {
	if r.HTMLBody == "" {
		return nil
	}

	message.AddAlternative("text/html", r.HTMLBody)
	return nil
}
