package event

import "github.com/skygeario/skygear-server/pkg/core/uuid"

type Type string

type Payload interface {
}

// UserAwarePayload represents event payload that can apply mutations on its own user object
type UserAwarePayload interface {
	Payload
	ApplyingMutations(Mutations) UserAwarePayload
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
		ID:      uuid.New(),
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
