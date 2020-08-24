package hook

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

//go:generate mockgen -source=provider.go -destination=provider_mock_test.go -mock_names=deliverer=MockDeliverer,store=MockStore -package hook

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type deliverer interface {
	WillDeliver(eventType event.Type) bool
	DeliverBeforeEvent(event *event.Event) error
	DeliverNonBeforeEvent(event *event.Event) error
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

func (provider *Provider) DispatchEvent(payload event.Payload) (err error) {
	var seq int64
	switch typedPayload := payload.(type) {
	case event.OperationPayload:
		if provider.Deliverer.WillDeliver(typedPayload.BeforeEventType()) {
			seq, err = provider.Store.NextSequenceNumber()
			if err != nil {
				err = errorutil.HandledWithMessage(err, "failed to dispatch event")
				return
			}
			event := event.NewBeforeEvent(seq, typedPayload, provider.makeContext())
			err = provider.Deliverer.DeliverBeforeEvent(event)
			if err != nil {
				if !apierrors.IsKind(err, WebHookDisallowed) {
					err = errorutil.HandledWithMessage(err, "failed to dispatch event")
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
					err = errorutil.HandledWithMessage(err, "failed to persist event")
					return err
				}
				ev = event.NewAfterEvent(seq, typedPayload, provider.makeContext())
			}

		case event.NotificationPayload:
			if provider.Deliverer.WillDeliver(typedPayload.EventType()) {
				seq, err := provider.Store.NextSequenceNumber()
				if err != nil {
					err = errorutil.HandledWithMessage(err, "failed to persist event")
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
		err = errorutil.HandledWithMessage(err, "failed to persist event")
		return err
	}
	provider.persistentEventPayloads = nil

	return nil
}

func (provider *Provider) DidCommitTx() {
	// TODO(webhook): deliver persisted events
	events, _ := provider.Store.GetEventsForDelivery()
	for _, event := range events {
		err := provider.Deliverer.DeliverNonBeforeEvent(event)
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
		userIDToSync = append(userIDToSync, payload.UserID())
	}

	for _, userID := range userIDToSync {
		user, err := provider.Users.Get(userID)
		if err != nil {
			return err
		}

		payload := &event.UserSyncEvent{User: *user}
		err = provider.DispatchEvent(payload)
		if err != nil {
			return err
		}
	}

	return nil
}

func (provider *Provider) makeContext() event.Context {
	userID := session.GetUserID(provider.Context)

	return event.Context{
		Timestamp: provider.Clock.NowUTC().Unix(),
		UserID:    userID,
	}
}
