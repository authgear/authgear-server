package event

import "fmt"

type Type string

type Payload interface {
	UserID() string
	IsAdminAPI() bool
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
}

type Event struct {
	ID            string  `json:"id"`
	Seq           int64   `json:"seq"`
	Type          Type    `json:"type"`
	Payload       Payload `json:"payload"`
	Context       Context `json:"context"`
	IsNonBlocking bool    `json:"-"`
}

func newEvent(seq int64, payload Payload, context Context) *Event {
	e := &Event{
		Payload: payload,
		Context: context,
	}
	e.Seq = seq
	e.ID = fmt.Sprintf("%016x", seq)
	return e
}

func NewBlockingEvent(seqNo int64, payload BlockingPayload, context Context) *Event {
	event := newEvent(seqNo, payload, context)
	event.Type = payload.BlockingEventType()
	return event
}

func NewNonBlockingEvent(seqNo int64, payload NonBlockingPayload, context Context) *Event {
	event := newEvent(seqNo, payload, context)
	event.Type = payload.NonBlockingEventType()
	event.IsNonBlocking = true
	return event
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
