package event

const ContextVersion int32 = 1

type Context struct {
	Timestamp   int64   `json:"timestamp"`
	RequestID   *string `json:"request_id"`
	UserID      *string `json:"user_id"`
	PrincipalID *string `json:"identity_id"`
}
