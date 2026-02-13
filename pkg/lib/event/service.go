package event

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

//go:generate go tool mockgen -source=service.go -destination=service_mock_test.go -package event

type Database interface {
	UseHook(ctx context.Context, hook db.TransactionHook)
}

type Sink interface {
	ReceiveBlockingEvent(ctx context.Context, e *event.Event) error
	ReceiveNonBlockingEvent(ctx context.Context, e *event.Event) error
}

type Store interface {
	NextSequenceNumber(ctx context.Context) (int64, error)
}

type Resolver interface {
	Resolve(ctx context.Context, anything interface{}) (err error)
}

var EventLogger = slogutil.NewLogger("event")

type Service struct {
	AppID           config.AppID
	RemoteIP        httputil.RemoteIP
	UserAgentString httputil.UserAgentString
	HTTPRequestURL  httputil.HTTPRequestURL
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

// DispatchEventOnCommit dispatches the event according to the tranaction lifecycle.
func (s *Service) DispatchEventOnCommit(ctx context.Context, payload event.Payload) (err error) {
	defer func() {
		if err != nil {
			s.IsDispatchEventErr = true
		}
	}()

	if !s.DatabaseHooked {
		s.Database.UseHook(ctx, s)
		s.DatabaseHooked = true
	}

	// Resolve refs once here
	// If the event is about entity deletion,
	// then it is not possible to resolve the entity in DidCommitTx.
	err = s.Resolver.Resolve(ctx, payload)
	if err != nil {
		return
	}

	switch typedPayload := payload.(type) {
	case event.BlockingPayload:
		eventContext := s.makeContext(ctx, payload)
		var seq int64
		seq, err = s.nextSeq(ctx)
		if err != nil {
			return
		}
		e := newBlockingEvent(seq, typedPayload, eventContext)
		for _, sink := range s.Sinks {
			err = sink.ReceiveBlockingEvent(ctx, e)
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

// DispatchEventImmediately dispatches the event immediately.
func (s *Service) DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) (err error) {
	logger := EventLogger.GetLogger(ctx)
	// Resolve refs once here
	// If the event is about entity deletion,
	// then it is not possible to resolve the entity in DidRollbackTx.
	err = s.Resolver.Resolve(ctx, payload)
	if err != nil {
		return
	}

	e, err := s.resolveNonBlockingEvent(ctx, payload)
	if err != nil {
		return err
	}

	for _, sink := range s.Sinks {
		err = sink.ReceiveNonBlockingEvent(ctx, e)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to dispatch nonblocking error event")
		}
	}

	return
}

// DispatchEventWithoutTx dispatches the blocking event immediately without transaction.
func (s *Service) DispatchEventWithoutTx(ctx context.Context, e *event.Event) (err error) {
	for _, sink := range s.Sinks {
		err = sink.ReceiveBlockingEvent(ctx, e)
		if err != nil {
			return
		}
	}
	return
}

func (s *Service) PrepareBlockingEventWithTx(ctx context.Context, payload event.BlockingPayload) (e *event.Event, err error) {
	eventContext := s.makeContext(ctx, payload)
	var seq int64
	seq, err = s.nextSeq(ctx)
	if err != nil {
		return
	}
	err = s.Resolver.Resolve(ctx, payload)
	if err != nil {
		return
	}
	e = newBlockingEvent(seq, payload, eventContext)
	return
}

func (s *Service) WillCommitTx(ctx context.Context) (err error) {
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
		e, err := s.resolveNonBlockingEvent(ctx, payload)
		if err != nil {
			return err
		}
		s.NonBlockingEvents = append(s.NonBlockingEvents, e)
	}

	return
}

func (s *Service) DidCommitTx(ctx context.Context) {
	logger := EventLogger.GetLogger(ctx)

	// To avoid triggering the events multiple times
	// reset s.NonBlockingEvents when we start processing the events
	nonBlockingEvents := s.NonBlockingEvents
	s.NonBlockingEvents = nil

	for _, e := range nonBlockingEvents {
		for _, sink := range s.Sinks {
			err := sink.ReceiveNonBlockingEvent(ctx, e)
			if err != nil {
				logger.WithError(err).Error(ctx, "failed to dispatch nonblocking event")
			}
		}
	}
}

func (s *Service) nextSeq(ctx context.Context) (seq int64, err error) {
	seq, err = s.Store.NextSequenceNumber(ctx)
	if err != nil {
		return
	}
	return
}

func (s *Service) makeContext(ctx context.Context, payload event.Payload) event.Context {
	var userID *string

	uid := payload.UserID()
	if uid != "" {
		userID = &uid
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(ctx)
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

	uiParam := uiparam.GetUIParam(ctx)
	auditCtx := adminauthz.GetAdminAuthzAudit(ctx)
	auditCtx = event.NewAuditContext(string(s.HTTPRequestURL), auditCtx)
	clientID := uiParam.ClientID

	var oauthContext *event.OAuthContext
	if uiParam.State != "" || uiParam.XState != "" {
		oauthContext = &event.OAuthContext{
			State:  uiParam.State,
			XState: uiParam.XState,
		}
	}

	var geoIPCountryCode *string
	geoipInfo, ok := geoip.IPString(string(s.RemoteIP))
	if ok && geoipInfo.CountryCode != "" {
		geoIPCountryCode = &geoipInfo.CountryCode
	}

	eventCtx := &event.Context{
		Timestamp:          s.Clock.NowUTC().Unix(),
		UserID:             userID,
		TriggeredBy:        triggeredBy,
		AuditContext:       auditCtx,
		PreferredLanguages: preferredLanguageTags,
		Language:           resolvedLanguage,
		IPAddress:          string(s.RemoteIP),
		GeoLocationCode:    geoIPCountryCode,
		UserAgent:          string(s.UserAgentString),
		AppID:              string(s.AppID),
		ClientID:           clientID,
		OAuth:              oauthContext,
		TrackingID:         errorutil.FormatTrackingID(ctx),
	}

	payload.FillContext(eventCtx)

	return *eventCtx
}

func (s *Service) resolveNonBlockingEvent(ctx context.Context, payload event.NonBlockingPayload) (*event.Event, error) {
	eventContext := s.makeContext(ctx, payload)
	seq, err := s.nextSeq(ctx)
	if err != nil {
		return nil, err
	}
	err = s.Resolver.Resolve(ctx, payload)
	if err != nil {
		return nil, err
	}
	return NewNonBlockingEvent(seq, payload, eventContext), nil
}
