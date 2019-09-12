package mfa

type Sender interface {
	Send(code string, phone string, email string) error
}
