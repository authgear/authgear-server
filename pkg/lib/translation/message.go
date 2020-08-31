package translation

type MessageSpec struct {
	Name          string
	TXTEmailType  string
	HTMLEmailType string
	SMSType       string
}

func RegisterMessage(msg *MessageSpec) *MessageSpec {
	return msg
}
