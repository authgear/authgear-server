package event

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
)

func newEvent(seq int64, payload event.Payload, context event.Context) *event.Event {
	e := &event.Event{
		Payload: payload,
		Context: context,
	}
	e.Seq = seq
	e.ID = fmt.Sprintf("%016x", seq)
	return e
}

func newBlockingEvent(seqNo int64, payload event.BlockingPayload, context event.Context) *event.Event {
	event := newEvent(seqNo, payload, context)
	event.Type = payload.BlockingEventType()
	return event
}

func newNonBlockingEvent(seqNo int64, payload event.NonBlockingPayload, context event.Context) *event.Event {
	event := newEvent(seqNo, payload, context)
	event.Type = payload.NonBlockingEventType()
	event.IsNonBlocking = true
	return event
}
