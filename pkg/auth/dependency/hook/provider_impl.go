package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/time"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type providerImpl struct {
	RequestID    string
	Store        Store
	AuthContext  auth.ContextGetter
	TimeProvider time.Provider
	Deliverer    Deliverer
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

		// TODO(webhook): after events
		return

	case event.NotificationPayload:
		// TODO(webhook): delayed delivery
		err = nil
		return

	default:
		panic(InvalidEventPayload{payload: payload})
	}
}

func (provider *providerImpl) WillCommitTx() error {
	// TODO(webhook): real impl
	return nil
}

func (provider *providerImpl) DidCommitTx() {

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
