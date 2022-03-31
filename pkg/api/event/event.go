package event

type Type string

type Payload interface {
	UserID() string
	GetTriggeredBy() TriggeredByType
	FillContext(ctx *Context)
}

type BlockingPayload interface {
	Payload
	BlockingEventType() Type
	ApplyMutations(mutations Mutations) (BlockingPayload, bool)
	GenerateFullMutations() Mutations
}

type NonBlockingPayload interface {
	Payload
	NonBlockingEventType() Type
	ForWebHook() bool
	ForAudit() bool
	ReindexUserNeeded() bool
	IsUserDeleted() bool
}

type Event struct {
	ID            string  `json:"id"`
	Seq           int64   `json:"seq"`
	Type          Type    `json:"type"`
	Payload       Payload `json:"payload"`
	Context       Context `json:"context"`
	IsNonBlocking bool    `json:"-"`
}

func (e *Event) ApplyMutations(mutations Mutations) (*Event, bool) {
	if blockingPayload, ok := e.Payload.(BlockingPayload); ok {
		if payload, applied := blockingPayload.ApplyMutations(mutations); applied {
			copied := *e
			copied.Payload = payload
			return &copied, true
		}
	}

	return e, false
}

func (e *Event) GenerateFullMutations() (*Mutations, bool) {
	if blockingPayload, ok := e.Payload.(BlockingPayload); ok {
		mutations := blockingPayload.GenerateFullMutations()
		return &mutations, true
	}

	return nil, false
}
