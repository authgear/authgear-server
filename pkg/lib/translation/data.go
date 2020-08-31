package translation

type AppMetadata struct {
	AppName string
	LogoURI string
}

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
