package event

type Type string

type Event struct {
	Version    int32       `json:"version"`
	ID         string      `json:"id"`
	SequenceNo int64       `json:"seq"`
	Type       Type        `json:"type"`
	Payload    interface{} `json:"payload"`
	Context    Context     `json:"context"`
}

type Context struct {
	Timestamp   int64   `json:"timestamp"`
	RequestID   *string `json:"request_id"`
	UserID      *string `json:"user_id"`
	PrincipalID *string `json:"identity_id"`
}
