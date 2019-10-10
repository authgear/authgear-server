package config

type DefaultConfiguration struct {
	SMTP   SMTPConfiguration   `envconfig:"SMTP"`
	Twilio TwilioConfiguration `envconfig:"TWILIO"`
	Nexmo  NexmoConfiguration  `envconfig:"NEXMO"`
}
