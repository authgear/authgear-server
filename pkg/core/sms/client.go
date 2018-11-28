package sms

type Client interface {
	Send(to string, body string) error
}
