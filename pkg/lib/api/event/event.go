package event

import "fmt"

type Type string

type Payload interface {
	UserID() string
}

// NotificationPayload represents event payload for notifications, with single event type variant
type NotificationPayload interface {
	Payload
	EventType() Type
}

// OperationPayload represents event payload for operations, with BEFORE and AFTER event type variant
type OperationPayload interface {
	Payload
	BeforeEventType() Type
	AfterEventType() Type
}

type Event struct {
	ID      string  `json:"id"`
	Seq     int64   `json:"seq"`
	Type    Type    `json:"type"`
	Payload Payload `json:"payload"`
	Context Context `json:"context"`
}

func newEvent(seqNo int64, payload Payload, context Context) *Event {
	return &Event{
		ID:      fmt.Sprintf("%016x", seqNo),
		Seq:     seqNo,
		Payload: payload,
		Context: context,
	}
}

func NewEvent(seqNo int64, payload NotificationPayload, context Context) *Event {
	event := newEvent(seqNo, payload, context)
	event.Type = payload.EventType()
	return event
}

func NewBeforeEvent(seqNo int64, payload OperationPayload, context Context) *Event {
	event := newEvent(seqNo, payload, context)
	event.Type = payload.BeforeEventType()
	return event
}

func NewAfterEvent(seqNo int64, payload OperationPayload, context Context) *Event {
	event := newEvent(seqNo, payload, context)
	event.Type = payload.AfterEventType()
	return event
}
