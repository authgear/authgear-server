package event

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
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

type Store interface {
	NextSequenceNumber() (int64, error)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("event")} }

type Service struct {
	Context      context.Context
	Request      *http.Request
	TrustProxy   config.TrustProxy
	Logger       Logger
	Database     Database
	Clock        clock.Clock
	Users        UserService
	Localization *config.LocalizationConfig
	Store        Store
	Sinks        []Sink

	NonBlockingEvents  []*event.Event `wire:"-"`
	DatabaseHooked     bool           `wire:"-"`
	IsDispatchEventErr bool           `wire:"-"`
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

	eventContext := s.makeContext(payload)
	var seq int64
	seq, err = s.Store.NextSequenceNumber()
	if err != nil {
		return
	}

	switch typedPayload := payload.(type) {
	case event.BlockingPayload:
		e := event.NewBlockingEvent(seq, typedPayload, eventContext)
		for _, sink := range s.Sinks {
			err = sink.ReceiveBlockingEvent(e)
			if err != nil {
				return
			}
		}
	case event.NonBlockingPayload:
		e := event.NewNonBlockingEvent(seq, typedPayload, eventContext)
		s.NonBlockingEvents = append(s.NonBlockingEvents, e)
	default:
		panic(fmt.Sprintf("event: invalid event payload: %T", payload))
	}

	return
}

func (s *Service) WillCommitTx() (err error) {
	// no-op
	return
}

func (s *Service) DidCommitTx() {
	// Skip non-blocking event if there is error during blocking event.
	if s.IsDispatchEventErr {
		return
	}

	for _, e := range s.NonBlockingEvents {
		for _, sink := range s.Sinks {
			err := sink.ReceiveNonBlockingEvent(e)
			if err != nil {
				s.Logger.WithError(err).Error("failed to dispatch non blocking event")
			}
		}
	}
	s.NonBlockingEvents = nil
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

	ipAddress := httputil.GetIP(s.Request, bool(s.TrustProxy))
	userAgent := s.Request.UserAgent()
	clientID := clientid.GetClientID(s.Context)

	ctx := &event.Context{
		Timestamp:          s.Clock.NowUTC().Unix(),
		UserID:             userID,
		PreferredLanguages: preferredLanguageTags,
		Language:           resolvedLanguage,
		TriggeredBy:        triggeredBy,
		IPAddress:          ipAddress,
		UserAgent:          userAgent,
		ClientID:           clientID,
	}

	payload.FillContext(ctx)

	return *ctx
}
