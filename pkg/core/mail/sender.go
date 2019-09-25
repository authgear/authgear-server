package mail

type Sender interface {
	Send(opts SendOptions) error
}

type SendOptions struct {
	Sender    string
	Recipient string
	Subject   string
	ReplyTo   string
	TextBody  string
	HTMLBody  string
}
