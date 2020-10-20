package translation

type EmailMessageData struct {
	Sender   string
	ReplyTo  string
	Subject  string
	HTMLBody string
	TextBody string
}

type SMSMessageData struct {
	Sender string
	Body   string
}
