package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/time"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type providerImpl struct {
	RequestID               string
	Store                   Store
	AuthContext             auth.ContextGetter
	TimeProvider            time.Provider
	Deliverer               Deliverer
	PersistentEventPayloads []event.Payload
}

func NewProvider(
	requestID string,
	store Store,
	authContext auth.ContextGetter,
	timeProvider time.Provider,
	deliverer Deliverer,
) Provider {
	return &providerImpl{
		RequestID:    requestID,
		Store:        store,
		AuthContext:  authContext,
		TimeProvider: timeProvider,
		Deliverer:    deliverer,
	}
}

func (provider *providerImpl) DispatchEvent(payload event.Payload, user *model.User) (err error) {
	var seq int64
	switch typedPayload := payload.(type) {
	case event.OperationPayload:
		seq, err = provider.Store.NextSequenceNumber()
		if err != nil {
			return
		}
		event := event.NewBeforeEvent(seq, typedPayload, provider.makeContext())
		err = provider.Deliverer.DeliverBeforeEvent(event, user)
		if err != nil {
			return err
		}

		// use event.payload since it may have been updated by mutations
		provider.PersistentEventPayloads = append(provider.PersistentEventPayloads, event.Payload)
		return

	case event.NotificationPayload:
		provider.PersistentEventPayloads = append(provider.PersistentEventPayloads, payload)
		err = nil
		return

	default:
		panic(InvalidEventPayload{payload: payload})
	}
}

func (provider *providerImpl) WillCommitTx() error {
	events := []*event.Event{}
	for _, payload := range provider.PersistentEventPayloads {
		seq, err := provider.Store.NextSequenceNumber()
		if err != nil {
			return err
		}

		var ev *event.Event
		switch typedPayload := payload.(type) {
		case event.OperationPayload:
			ev = event.NewAfterEvent(seq, typedPayload, provider.makeContext())
		case event.NotificationPayload:
			ev = event.NewEvent(seq, typedPayload, provider.makeContext())
		default:
			panic(InvalidEventPayload{payload: payload})
		}

		events = append(events, ev)
	}

	err := provider.Store.PersistEvents(events)
	if err != nil {
		return err
	}

	return nil
}

func (provider *providerImpl) DidCommitTx() {
	// TODO(webhook): deliver persisted events
}

func (provider *providerImpl) makeContext() event.Context {
	var requestID, userID, principalID *string

	if provider.RequestID == "" {
		requestID = nil
	} else {
		requestID = &provider.RequestID
	}

	if provider.AuthContext.AuthInfo() == nil {
		userID = nil
		principalID = nil
	} else {
		userID = &provider.AuthContext.AuthInfo().ID
		principalID = &provider.AuthContext.Token().PrincipalID
	}

	return event.Context{
		Timestamp:   provider.TimeProvider.NowUTC().Unix(),
		RequestID:   requestID,
		UserID:      userID,
		PrincipalID: principalID,
	}
}
