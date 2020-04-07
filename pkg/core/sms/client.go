package sms

type Client interface {
	Send(from string, to string, body string) error
}
