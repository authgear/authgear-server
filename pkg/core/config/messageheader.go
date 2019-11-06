package config

type MessageHeader struct {
	Subject string `json:"subject"`
	Sender  string `json:"sender"`
	ReplyTo string `json:"reply_to"`
}
