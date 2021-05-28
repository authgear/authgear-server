package event

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package event

type UserService interface {
	Get(id string) (*model.User, error)
}

type Database interface {
	UseHook(hook db.TransactionHook)
}

type Sink interface {
	ReceiveBlockingEvent(e *event.Event) error
	ReceiveNonBlockingEvent(e *event.Event) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("event")} }

type Service struct {
	Context      context.Context
	Logger       Logger
	Database     Database
	Clock        clock.Clock
	Users        UserService
	Localization *config.LocalizationConfig
	Sinks        []Sink

	NonBlockingEventPayloads []event.NonBlockingPayload `wire:"-"`
	DatabaseHooked           bool                       `wire:"-"`
	IsDispatchEventErr       bool                       `wire:"-"`
}

func (s *Service) DispatchEvent(payload event.Payload) (err error) {
	defer func() {
		if err != nil {
			s.IsDispatchEventErr = true
		}
	}()

	if !s.DatabaseHooked {
		s.Database.UseHook(s)
		s.DatabaseHooked = true
	}

	switch typedPayload := payload.(type) {
	case event.BlockingPayload:
		for _, sink := range s.Sinks {
			eventContext := s.makeContext(typedPayload)
			e := event.NewBlockingEvent(0, typedPayload, eventContext)
			err = sink.ReceiveBlockingEvent(e)
			if err != nil {
				return
			}
		}
	case event.NonBlockingPayload:
		s.NonBlockingEventPayloads = append(s.NonBlockingEventPayloads, typedPayload)
	default:
		panic(fmt.Sprintf("event: invalid event payload: %T", payload))
	}

	return
}

func (s *Service) WillCommitTx() error {
	// no-op
	return nil
}

func (s *Service) DidCommitTx() {
	// Skip non-blocking event if there is error during blocking event.
	if s.IsDispatchEventErr {
		return
	}

	var events []*event.Event
	for _, payload := range s.NonBlockingEventPayloads {
		eventContext := s.makeContext(payload)
		e := event.NewNonBlockingEvent(0, payload, eventContext)
		events = append(events, e)
	}

	for _, e := range events {
		for _, sink := range s.Sinks {
			err := sink.ReceiveNonBlockingEvent(e)
			if err != nil {
				s.Logger.WithError(err).Error("failed to dispatch non blocking event")
			}
		}
	}
}

func (s *Service) makeContext(payload event.Payload) event.Context {
	userID := session.GetUserID(s.Context)
	preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)
	resolvedLanguageIdx, _ := intl.Resolve(
		preferredLanguageTags,
		*s.Localization.FallbackLanguage,
		s.Localization.SupportedLanguages,
	)

	resolvedLanguage := *s.Localization.FallbackLanguage
	if resolvedLanguageIdx != -1 {
		resolvedLanguage = s.Localization.SupportedLanguages[resolvedLanguageIdx]
	}

	triggeredBy := event.TriggeredByTypeUser
	if payload.IsAdminAPI() {
		triggeredBy = event.TriggeredByTypeAdminAPI
	}

	ctx := &event.Context{
		Timestamp:          s.Clock.NowUTC().Unix(),
		UserID:             userID,
		PreferredLanguages: preferredLanguageTags,
		Language:           resolvedLanguage,
		TriggeredBy:        triggeredBy,
	}

	payload.FillContext(ctx)

	return *ctx
}
