package hook

import (
	"context"
	"fmt"
	gotime "time"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type providerImpl struct {
	RequestID               string
	Store                   Store
	Context                 context.Context
	TxContext               db.TxContext
	TimeProvider            time.Provider
	AuthInfoStore           authinfo.Store
	UserProfileStore        userprofile.Store
	Deliverer               Deliverer
	PersistentEventPayloads []event.Payload
	Logger                  *logrus.Entry

	txHooked bool
}

func NewProvider(
	ctx context.Context,
	requestID string,
	store Store,
	txContext db.TxContext,
	timeProvider time.Provider,
	authInfoStore authinfo.Store,
	userProfileStore userprofile.Store,
	deliverer Deliverer,
	loggerFactory logging.Factory,
) Provider {
	return &providerImpl{
		Context:          ctx,
		RequestID:        requestID,
		Store:            store,
		TxContext:        txContext,
		TimeProvider:     timeProvider,
		AuthInfoStore:    authInfoStore,
		UserProfileStore: userProfileStore,
		Deliverer:        deliverer,
		Logger:           loggerFactory.NewLogger("hook"),
	}
}

func (provider *providerImpl) DispatchEvent(payload event.Payload, user *model.User) (err error) {
	var seq int64
	switch typedPayload := payload.(type) {
	case event.OperationPayload:
		if provider.Deliverer.WillDeliver(typedPayload.BeforeEventType()) {
			seq, err = provider.Store.NextSequenceNumber()
			if err != nil {
				err = errors.HandledWithMessage(err, "failed to dispatch event")
				return
			}
			event := event.NewBeforeEvent(seq, typedPayload, provider.makeContext())
			err = provider.Deliverer.DeliverBeforeEvent(event, user)
			if err != nil {
				if !skyerr.IsKind(err, WebHookDisallowed) {
					err = errors.HandledWithMessage(err, "failed to dispatch event")
				}
				return
			}

			// update payload since it may have been updated by mutations
			payload = event.Payload
		}

		provider.PersistentEventPayloads = append(provider.PersistentEventPayloads, payload)

	case event.NotificationPayload:
		provider.PersistentEventPayloads = append(provider.PersistentEventPayloads, payload)
		err = nil

	default:
		panic(fmt.Sprintf("hook: invalid event payload: %T", payload))
	}

	if !provider.txHooked {
		provider.TxContext.UseHook(provider)
		provider.txHooked = true
	}
	return
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
					err = errors.HandledWithMessage(err, "failed to persist event")
					return err
				}
				ev = event.NewAfterEvent(seq, typedPayload, provider.makeContext())
			}

		case event.NotificationPayload:
			if provider.Deliverer.WillDeliver(typedPayload.EventType()) {
				seq, err := provider.Store.NextSequenceNumber()
				if err != nil {
					err = errors.HandledWithMessage(err, "failed to persist event")
					return err
				}
				ev = event.NewEvent(seq, typedPayload, provider.makeContext())
			}

		default:
			panic(fmt.Sprintf("hook: invalid event payload: %T", payload))
		}

		if ev == nil {
			continue
		}
		events = append(events, ev)
	}

	err = provider.Store.AddEvents(events)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to persist event")
		return err
	}
	provider.PersistentEventPayloads = nil

	return nil
}

func (provider *providerImpl) DidCommitTx() {
	// TODO(webhook): deliver persisted events
	events, _ := provider.Store.GetEventsForDelivery()
	for _, event := range events {
		err := provider.Deliverer.DeliverNonBeforeEvent(event, 60*gotime.Second)
		if err != nil {
			provider.Logger.WithError(err).Debug("Failed to dispatch event")
		}
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
	var requestID, userID *string
	var session *model.Session

	if provider.RequestID == "" {
		requestID = nil
	} else {
		requestID = &provider.RequestID
	}

	user := authn.GetUser(provider.Context)
	sess := authn.GetSession(provider.Context)
	if user == nil {
		userID = nil
		session = nil
	} else {
		userID = &user.ID
		session = sess.(auth.AuthSession).ToAPIModel()
	}

	return event.Context{
		Timestamp: provider.TimeProvider.NowUTC().Unix(),
		RequestID: requestID,
		UserID:    userID,
		Session:   session,
	}
}
