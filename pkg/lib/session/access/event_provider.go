package access

import (
	"time"
)

type EventProvider struct {
	Store EventStore
}

func (p *EventProvider) InitStream(sessionID string, expiry time.Time, initialAccess *Event) error {
	if err := p.Store.ResetEventStream(sessionID); err != nil {
		return err
	}
	if err := p.Store.AppendEvent(sessionID, expiry, initialAccess); err != nil {
		return err
	}
	return nil
}

func (p *EventProvider) RecordAccess(sessionID string, expiry time.Time, event *Event) error {
	if err := p.Store.AppendEvent(sessionID, expiry, event); err != nil {
		return err
	}
	return nil
}
