package hook

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	tenantdb "github.com/authgear/authgear-server/pkg/lib/infra/db/tenant"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
)

//go:generate mockgen -source=provider.go -destination=provider_mock_test.go -mock_names=deliverer=MockDeliverer,store=MockStore -package hook

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type deliverer interface {
	WillDeliverBlockingEvent(eventType event.Type) bool
	WillDeliverNonBlockingEvent(eventType event.Type) bool
	DeliverBlockingEvent(event *event.Event) error
	DeliverNonBlockingEvent(event *event.Event) error
}

type store interface {
	NextSequenceNumber() (int64, error)
	AddEvents(events []*event.Event) error
	GetEventsForDelivery() ([]*event.Event, error)
}

type DatabaseHandle interface {
	UseHook(hook tenantdb.TransactionHook)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("hook")} }

type Provider struct {
	Context      context.Context
	Logger       Logger
	Database     DatabaseHandle
	Clock        clock.Clock
	Users        UserProvider
	Store        store
	Deliverer    deliverer
	Localization *config.LocalizationConfig

	persistentEventPayloads []event.Payload `wire:"-"`
	dbHooked                bool            `wire:"-"`
	IsDispatchEventErr      bool            `wire:"-"`
}

func (provider *Provider) DispatchEvent(payload event.Payload) (err error) {
	var seq int64
	defer func() {
		if err != nil {
			provider.IsDispatchEventErr = true
		}
	}()

	if !provider.dbHooked {
		provider.Database.UseHook(provider)
		provider.dbHooked = true
	}

	switch typedPayload := payload.(type) {

	case event.BlockingPayload:
		if provider.Deliverer.WillDeliverBlockingEvent(typedPayload.BlockingEventType()) {
			seq, err = provider.Store.NextSequenceNumber()
			if err != nil {
				err = fmt.Errorf("failed to dispatch event: %w", err)
				return
			}
			event := event.NewBlockingEvent(seq, typedPayload, provider.makeContext(typedPayload.IsAdminAPI()))
			err = provider.Deliverer.DeliverBlockingEvent(event)
			if err != nil {
				if !apierrors.IsKind(err, WebHookDisallowed) {
					err = fmt.Errorf("failed to dispatch event: %w", err)
				}
				return
			}
		}

	case event.NonBlockingPayload:
		provider.persistentEventPayloads = append(provider.persistentEventPayloads, payload)
		err = nil

	default:
		panic(fmt.Sprintf("hook: invalid event payload: %T", payload))
	}

	return
}

func (provider *Provider) WillCommitTx() error {
	// should skip persistent events if there is error during dispatch events
	if provider.IsDispatchEventErr {
		return nil
	}

	events := []*event.Event{}
	for _, payload := range provider.persistentEventPayloads {
		var ev *event.Event

		switch typedPayload := payload.(type) {

		case event.NonBlockingPayload:
			if provider.Deliverer.WillDeliverNonBlockingEvent(typedPayload.NonBlockingEventType()) {
				seq, err := provider.Store.NextSequenceNumber()
				if err != nil {
					err = fmt.Errorf("failed to persist event: %w", err)
					return err
				}
				ev = event.NewNonBlockingEvent(seq, typedPayload, provider.makeContext(typedPayload.IsAdminAPI()))
			}
		default:
			panic(fmt.Sprintf("hook: invalid event payload: %T", payload))
		}

		if ev == nil {
			continue
		}
		events = append(events, ev)
	}

	err := provider.Store.AddEvents(events)
	if err != nil {
		err = fmt.Errorf("failed to persist event: %w", err)
		return err
	}
	provider.persistentEventPayloads = nil

	return nil
}

func (provider *Provider) DidCommitTx() {
	// should skip further dispatch events if there is error during dispatch events
	if provider.IsDispatchEventErr {
		return
	}

	// TODO(webhook): deliver persisted events
	events, _ := provider.Store.GetEventsForDelivery()
	for _, event := range events {
		if err := provider.Deliverer.DeliverNonBlockingEvent(event); err != nil {
			provider.Logger.WithError(err).Debug("Failed to dispatch non blocking event")
		}
	}
}

func (provider *Provider) makeContext(isAdminAPI bool) event.Context {
	userID := session.GetUserID(provider.Context)
	preferredLanguageTags := intl.GetPreferredLanguageTags(provider.Context)
	resolvedLanguageIdx, _ := intl.Resolve(
		preferredLanguageTags,
		*provider.Localization.FallbackLanguage,
		provider.Localization.SupportedLanguages,
	)

	resolvedLanguage := *provider.Localization.FallbackLanguage
	if resolvedLanguageIdx != -1 {
		resolvedLanguage = provider.Localization.SupportedLanguages[resolvedLanguageIdx]
	}

	triggeredBy := event.TriggeredByTypeUser
	if isAdminAPI {
		triggeredBy = event.TriggeredByTypeAdminAPI
	}

	return event.Context{
		Timestamp:          provider.Clock.NowUTC().Unix(),
		UserID:             userID,
		PreferredLanguages: preferredLanguageTags,
		Language:           resolvedLanguage,
		TriggeredBy:        triggeredBy,
	}
}
