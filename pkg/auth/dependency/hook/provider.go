package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/log"
)

//go:generate mockgen -source=provider.go -destination=provider_mock_test.go -mock_names=deliverer=MockDeliverer,store=MockStore -package hook

type UserProvider interface {
	Get(id string) (*model.User, error)
	UpdateMetadata(user *model.User, metadata map[string]interface{}) error
}

type deliverer interface {
	WillDeliver(eventType event.Type) bool
	DeliverBeforeEvent(event *event.Event, user *model.User) error
	DeliverNonBeforeEvent(event *event.Event, timeout time.Duration) error
}

type store interface {
	NextSequenceNumber() (int64, error)
	AddEvents(events []*event.Event) error
	GetEventsForDelivery() ([]*event.Event, error)
}

type DatabaseHandle interface {
	UseHook(hook db.TransactionHook)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("hook")} }

type Provider struct {
	Context   context.Context
	Logger    Logger
	Database  DatabaseHandle
	Clock     clock.Clock
	Users     UserProvider
	Store     store
	Deliverer deliverer

	persistentEventPayloads []event.Payload `wire:"-"`
	dbHooked                bool            `wire:"-"`
}

func (provider *Provider) DispatchEvent(payload event.Payload, user *model.User) (err error) {
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

		provider.persistentEventPayloads = append(provider.persistentEventPayloads, payload)

	case event.NotificationPayload:
		provider.persistentEventPayloads = append(provider.persistentEventPayloads, payload)
		err = nil

	default:
		panic(fmt.Sprintf("hook: invalid event payload: %T", payload))
	}

	if !provider.dbHooked {
		provider.Database.UseHook(provider)
		provider.dbHooked = true
	}
	return
}

func (provider *Provider) WillCommitTx() error {
	err := provider.dispatchSyncUserEventIfNeeded()
	if err != nil {
		return err
	}

	events := []*event.Event{}
	for _, payload := range provider.persistentEventPayloads {
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
	provider.persistentEventPayloads = nil

	return nil
}

func (provider *Provider) DidCommitTx() {
	// TODO(webhook): deliver persisted events
	events, _ := provider.Store.GetEventsForDelivery()
	for _, event := range events {
		err := provider.Deliverer.DeliverNonBeforeEvent(event, 60*time.Second)
		if err != nil {
			provider.Logger.WithError(err).Debug("Failed to dispatch event")
		}
	}
}

func (provider *Provider) dispatchSyncUserEventIfNeeded() error {
	userIDToSync := []string{}

	for _, payload := range provider.persistentEventPayloads {
		if _, isOperation := payload.(event.OperationPayload); !isOperation {
			continue
		}
		if userAwarePayload, ok := payload.(event.UserAwarePayload); ok {
			userIDToSync = append(userIDToSync, userAwarePayload.UserID())
		}
	}

	for _, userID := range userIDToSync {
		user, err := provider.Users.Get(userID)
		if err != nil {
			return err
		}

		payload := event.UserSyncEvent{User: *user}
		err = provider.DispatchEvent(payload, user)
		if err != nil {
			return err
		}
	}

	return nil
}

func (provider *Provider) makeContext() event.Context {
	userID := authn.GetUserID(provider.Context)

	return event.Context{
		Timestamp: provider.Clock.NowUTC().Unix(),
		UserID:    userID,
	}
}
