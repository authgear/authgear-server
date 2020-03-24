package auth

type AccessEventProvider struct {
	Store AccessEventStore
}

func (p *AccessEventProvider) InitStream(s AuthSession) error {
	if err := p.Store.ResetEventStream(s); err != nil {
		return err
	}
	if err := p.Store.AppendAccessEvent(s, &s.GetAccessInfo().InitialAccess); err != nil {
		return err
	}
	return nil
}

func (p *AccessEventProvider) RecordAccess(s AuthSession, event AccessEvent) error {
	if err := p.Store.AppendAccessEvent(s, &event); err != nil {
		return err
	}
	return nil
}
