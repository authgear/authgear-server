package cmdsetup

import (
	"gopkg.in/gomail.v2"
)

type SendTestEmailOptions struct {
	Host          string
	Port          int
	Username      string
	Password      string
	SenderAddress string
	ToAddress     string
}

func SendTestEmail(opts SendTestEmailOptions) error {
	dialer := gomail.NewDialer(opts.Host, opts.Port, opts.Username, opts.Password)
	message := gomail.NewMessage()
	message.SetHeader("From", opts.SenderAddress)
	message.SetHeader("To", opts.ToAddress)
	message.SetHeader("Subject", "Test email from Authgear")
	message.SetBody("text/plain", "Hi there! There is a test email from Authgear")
	return dialer.DialAndSend(message)
}
