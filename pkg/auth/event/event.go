package event

import "github.com/skygeario/skygear-server/pkg/core/uuid"

type Type string

type Payload interface {
	Version() int32
}

type Event struct {
	Version    int32   `json:"version"`
	ID         string  `json:"id"`
	SequenceNo int64   `json:"seq"`
	Type       Type    `json:"type"`
	Payload    Payload `json:"payload"`
	Context    Context `json:"context"`
}

func NewEvent(eventType Type, seqNo int64, payload Payload, context Context) *Event {
	return &Event{
		Version:    payload.Version() + ContextVersion,
		ID:         uuid.New(),
		SequenceNo: seqNo,
		Type:       eventType,
		Payload:    payload,
		Context:    context,
	}
}
