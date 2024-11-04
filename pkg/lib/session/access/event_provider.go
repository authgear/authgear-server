package access

import (
	"context"
	"time"
)

type EventProvider struct {
	Store EventStore
}

func (p *EventProvider) InitStream(ctx context.Context, sessionID string, expiry time.Time, initialAccess *Event) error {
	if err := p.Store.ResetEventStream(ctx, sessionID); err != nil {
		return err
	}
	if err := p.Store.AppendEvent(ctx, sessionID, expiry, initialAccess); err != nil {
		return err
	}
	return nil
}

func (p *EventProvider) RecordAccess(ctx context.Context, sessionID string, expiry time.Time, event *Event) error {
	if err := p.Store.AppendEvent(ctx, sessionID, expiry, event); err != nil {
		return err
	}
	return nil
}
