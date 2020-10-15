package config

type MailConfig struct {
	Sender  string `envconfig:"SENDER"`
	ReplyTo string `envconfig:"REPLY_TO"`
}
