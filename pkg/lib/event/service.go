package event

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/clientid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package event

type Database interface {
	UseHook(hook db.TransactionHook)
}

type Sink interface {
	ReceiveBlockingEvent(e *event.Event) error
	ReceiveNonBlockingEvent(e *event.Event) error
}

type Store interface {
	NextSequenceNumber() (int64, error)
}

type Resolver interface {
	Resolve(anything interface{}) (err error)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("event")} }

type Service struct {
	Context         context.Context
	RemoteIP        httputil.RemoteIP
	UserAgentString httputil.UserAgentString
	Logger          Logger
	Database        Database
	Clock           clock.Clock
	Localization    *config.LocalizationConfig
	Store           Store
	Resolver        Resolver
	Sinks           []Sink

	NonBlockingPayloads []event.NonBlockingPayload `wire:"-"`
	NonBlockingEvents   []*event.Event             `wire:"-"`
	DatabaseHooked      bool                       `wire:"-"`
	IsDispatchEventErr  bool                       `wire:"-"`
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

	// Resolve refs once here
	// If the event is about entity deletion,
	// then it is not possible to resolve the entity in DidCommitTx.
	err = s.Resolver.Resolve(payload)
	if err != nil {
		return
	}

	switch typedPayload := payload.(type) {
	case event.BlockingPayload:
		eventContext := s.makeContext(payload)
		var seq int64
		seq, err = s.nextSeq()
		if err != nil {
			return
		}
		e := newBlockingEvent(seq, typedPayload, eventContext)
		for _, sink := range s.Sinks {
			err = sink.ReceiveBlockingEvent(e)
			if err != nil {
				return
			}
		}
	case event.NonBlockingPayload:
		s.NonBlockingPayloads = append(s.NonBlockingPayloads, typedPayload)
	default:
		panic(fmt.Sprintf("event: invalid event payload: %T", payload))
	}

	return
}

func (s *Service) WillCommitTx() (err error) {
	defer func() {
		s.NonBlockingPayloads = nil
	}()

	// Skip non-blocking event if there is error during blocking event.
	if s.IsDispatchEventErr {
		return
	}

	// We have to prepare the event here because we need an ongoing transaction
	// to get the seq number, as well as resolving refs.

	for _, payload := range s.NonBlockingPayloads {
		eventContext := s.makeContext(payload)
		seq, err := s.nextSeq()
		if err != nil {
			return err
		}
		err = s.Resolver.Resolve(payload)
		if err != nil {
			return err
		}
		e := newNonBlockingEvent(seq, payload, eventContext)
		s.NonBlockingEvents = append(s.NonBlockingEvents, e)
	}

	return
}

func (s *Service) DidCommitTx() {
	defer func() {
		s.NonBlockingEvents = nil
	}()

	for _, e := range s.NonBlockingEvents {
		for _, sink := range s.Sinks {
			err := sink.ReceiveNonBlockingEvent(e)
			if err != nil {
				s.Logger.WithError(err).Error("failed to dispatch nonblocking event")
			}
		}
	}
}

func (s *Service) nextSeq() (seq int64, err error) {
	seq, err = s.Store.NextSequenceNumber()
	if err != nil {
		return
	}
	return
}

func (s *Service) makeContext(payload event.Payload) event.Context {
	userID := session.GetUserID(s.Context)
	if userID == nil {
		uid := payload.UserID()
		if uid != "" {
			userID = &uid
		}
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)
	// Initialize this to an empty slice so that it is always present in the JSON.
	if preferredLanguageTags == nil {
		preferredLanguageTags = []string{}
	}
	resolvedLanguageIdx, _ := intl.Resolve(
		preferredLanguageTags,
		*s.Localization.FallbackLanguage,
		s.Localization.SupportedLanguages,
	)

	resolvedLanguage := ""
	if resolvedLanguageIdx != -1 {
		resolvedLanguage = s.Localization.SupportedLanguages[resolvedLanguageIdx]
	}

	triggeredBy := payload.GetTriggeredBy()

	clientID := clientid.GetClientID(s.Context)

	ctx := &event.Context{
		Timestamp:          s.Clock.NowUTC().Unix(),
		UserID:             userID,
		TriggeredBy:        triggeredBy,
		PreferredLanguages: preferredLanguageTags,
		Language:           resolvedLanguage,
		IPAddress:          string(s.RemoteIP),
		UserAgent:          string(s.UserAgentString),
		ClientID:           clientID,
	}

	payload.FillContext(ctx)

	return *ctx
}
