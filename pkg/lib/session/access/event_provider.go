package access

type EventProvider struct {
	Store EventStore
}

func (p *EventProvider) InitStream(sessionID string, initialAccess *Event) error {
	if err := p.Store.ResetEventStream(sessionID); err != nil {
		return err
	}
	if err := p.Store.AppendEvent(sessionID, initialAccess); err != nil {
		return err
	}
	return nil
}

func (p *EventProvider) RecordAccess(sessionID string, event *Event) error {
	if err := p.Store.AppendEvent(sessionID, event); err != nil {
		return err
	}
	return nil
}
