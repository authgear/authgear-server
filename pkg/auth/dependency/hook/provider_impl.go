package hook

import (
	"net/http"
	"net/url"
	gotime "time"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/time"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

type providerImpl struct {
	RequestID               string
	BaseURL                 *url.URL
	Store                   Store
	AuthContext             auth.ContextGetter
	TimeProvider            time.Provider
	AuthInfoStore           authinfo.Store
	UserProfileStore        userprofile.Store
	Deliverer               Deliverer
	PersistentEventPayloads []event.Payload
}

func NewProvider(
	requestID string,
	request *http.Request,
	store Store,
	authContext auth.ContextGetter,
	timeProvider time.Provider,
	authInfoStore authinfo.Store,
	userProfileStore userprofile.Store,
	deliverer Deliverer,
) Provider {
	return &providerImpl{
		RequestID:        requestID,
		BaseURL:          getHookBaseURL(request),
		Store:            store,
		AuthContext:      authContext,
		TimeProvider:     timeProvider,
		AuthInfoStore:    authInfoStore,
		UserProfileStore: userProfileStore,
		Deliverer:        deliverer,
	}
}

func (provider *providerImpl) DispatchEvent(payload event.Payload, user *model.User) (err error) {
	var seq int64
	switch typedPayload := payload.(type) {
	case event.OperationPayload:
		if provider.Deliverer.WillDeliver(typedPayload.BeforeEventType()) {
			seq, err = provider.Store.NextSequenceNumber()
			if err != nil {
				return
			}
			event := event.NewBeforeEvent(seq, typedPayload, provider.makeContext())
			err = provider.Deliverer.DeliverBeforeEvent(provider.BaseURL, event, user)
			if err != nil {
				return err
			}

			// update payload since it may have been updated by mutations
			payload = event.Payload
		}

		provider.PersistentEventPayloads = append(provider.PersistentEventPayloads, payload)
		return

	case event.NotificationPayload:
		provider.PersistentEventPayloads = append(provider.PersistentEventPayloads, payload)
		err = nil
		return

	default:
		panic(invalidEventPayload{payload: payload})
	}
}

func (provider *providerImpl) WillCommitTx() error {
	err := provider.dispatchSyncUserEventIfNeeded()
	if err != nil {
		return err
	}

	events := []*event.Event{}
	for _, payload := range provider.PersistentEventPayloads {
		var ev *event.Event

		switch typedPayload := payload.(type) {
		case event.OperationPayload:
			if provider.Deliverer.WillDeliver(typedPayload.AfterEventType()) {
				seq, err := provider.Store.NextSequenceNumber()
				if err != nil {
					return err
				}
				ev = event.NewAfterEvent(seq, typedPayload, provider.makeContext())
			}

		case event.NotificationPayload:
			if provider.Deliverer.WillDeliver(typedPayload.EventType()) {
				seq, err := provider.Store.NextSequenceNumber()
				if err != nil {
					return err
				}
				ev = event.NewEvent(seq, typedPayload, provider.makeContext())
			}

		default:
			panic(invalidEventPayload{payload: payload})
		}

		if ev == nil {
			continue
		}
		events = append(events, ev)
	}

	err = provider.Store.AddEvents(events)
	if err != nil {
		return err
	}
	provider.PersistentEventPayloads = nil

	return nil
}

func (provider *providerImpl) DidCommitTx() {
	// TODO(webhook): deliver persisted events
	events, _ := provider.Store.GetEventsForDelivery()
	for _, event := range events {
		_ = provider.Deliverer.DeliverNonBeforeEvent(provider.BaseURL, event, 60*gotime.Second)
	}
}

func (provider *providerImpl) dispatchSyncUserEventIfNeeded() error {
	userIDToSync := []string{}

	for _, payload := range provider.PersistentEventPayloads {
		if _, isOperation := payload.(event.OperationPayload); !isOperation {
			continue
		}
		if userAwarePayload, ok := payload.(event.UserAwarePayload); ok {
			userIDToSync = append(userIDToSync, userAwarePayload.UserID())
		}
	}

	for _, userID := range userIDToSync {
		var authInfo authinfo.AuthInfo
		err := provider.AuthInfoStore.GetAuth(userID, &authInfo)
		if err != nil {
			return err
		}

		userProfile, err := provider.UserProfileStore.GetUserProfile(userID)
		if err != nil {
			return err
		}

		user := model.NewUser(authInfo, userProfile)
		payload := event.UserSyncEvent{User: user}
		err = provider.DispatchEvent(payload, &user)
		if err != nil {
			return err
		}
	}

	return nil
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

func getHookBaseURL(req *http.Request) *url.URL {
	if req == nil {
		return &url.URL{}
	}

	u := &url.URL{
		Host:   corehttp.GetHost(req),
		Scheme: corehttp.GetProto(req),
	}
	return u
}
