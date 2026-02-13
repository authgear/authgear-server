package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	libevent "github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalsession "github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type AuditServiceAppService interface {
	Get(ctx context.Context, id string) (*model.App, error)
}

type AuditService struct {
	RemoteIP        httputil.RemoteIP
	UserAgentString httputil.UserAgentString
	HTTPRequestURL  httputil.HTTPRequestURL
	Request         *http.Request

	Apps     AuditServiceAppService
	Authgear *portalconfig.AuthgearConfig

	Database                  *db.Pool
	DatabaseEnvironmentConfig *config.DatabaseEnvironmentConfig

	DenoEndpoint config.DenoEndpoint

	GlobalSQLBuilder  *globaldb.SQLBuilder
	GlobalSQLExecutor *globaldb.SQLExecutor
	GlobalDatabase    *globaldb.Handle

	Clock clock.Clock
}

func (s *AuditService) Log(ctx context.Context, app *model.App, payload event.NonBlockingPayload) (err error) {

	authgearApp, err := s.Apps.Get(ctx, s.Authgear.AppID)
	if err != nil {
		return
	}

	// Legacy logging setup
	cfg := app.Context.Config

	// Modern logging setup
	ctx = slogutil.AddMaskPatterns(ctx, config.NewMaskPatternFromSecretConfig(cfg.SecretConfig))
	logger := slogutil.GetContextLogger(ctx)
	logger = logger.With(slog.String("app", string(cfg.AppConfig.ID)))
	ctx = slogutil.SetContextLogger(ctx, logger)

	// AuditSink is app specific.
	// The records MUST have correct app_id.
	// We have construct audit sink with the target app.
	auditSink := newAuditSink(app, s.Database, s.DatabaseEnvironmentConfig)
	// The portal uses its Authgear to deliver hooks.
	// We have construct hook sink with the Authgear app.
	hookSink := newHookSink(authgearApp, s.DenoEndpoint)

	// Use the target app ID.
	e, err := s.resolveNonBlockingEvent(ctx, app.ID, payload)
	if err != nil {
		return err
	}

	err = auditSink.ReceiveNonBlockingEvent(ctx, e)
	if err != nil {
		return err
	}

	err = hookSink.ReceiveNonBlockingEvent(ctx, e)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuditService) nextSeq(ctx context.Context) (seq int64, err error) {
	builder := s.GlobalSQLBuilder.
		Select(fmt.Sprintf("nextval('%s')", s.GlobalSQLBuilder.TableName("_auth_event_sequence")))
	row, err := s.GlobalSQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return
	}
	err = row.Scan(&seq)
	return
}

func (s *AuditService) makeContext(ctx context.Context, appID string, payload event.Payload) event.Context {
	var userIDStr string
	portalSession := portalsession.GetValidSessionInfo(ctx)
	if portalSession != nil {
		userIDStr = portalSession.UserID
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(ctx)
	// Initialize this to an empty slice so that it is always present in the JSON.
	if preferredLanguageTags == nil {
		preferredLanguageTags = []string{}
	}

	triggeredBy := payload.GetTriggeredBy()

	uiParam := uiparam.GetUIParam(ctx)
	clientID := uiParam.ClientID
	// This audit context must be constructed here.
	// We cannot use GetAdminAuthzAudit because that is for Admin API to audit context.
	portalAuditCtx := PortalAdminAPIAuthContext{
		Usage:       UsageInternal,
		ActorUserID: userIDStr,
		HTTPReferer: s.Request.Header.Get("Referer"),
	}

	auditCtx := portalAuditCtx.ToAuditContext(s.HTTPRequestURL)

	var geoIPCountryCode *string
	geoipInfo, ok := geoip.IPString(string(s.RemoteIP))
	if ok && geoipInfo.CountryCode != "" {
		geoIPCountryCode = &geoipInfo.CountryCode
	}

	eventCtx := &event.Context{
		Timestamp: s.Clock.NowUTC().Unix(),
		// We do not populate UserID because the event is not about UserID.
		TriggeredBy:        triggeredBy,
		AuditContext:       auditCtx,
		PreferredLanguages: preferredLanguageTags,
		Language:           "",
		IPAddress:          string(s.RemoteIP),
		GeoLocationCode:    geoIPCountryCode,
		UserAgent:          string(s.UserAgentString),
		ClientID:           clientID,
		AppID:              appID,
		TrackingID:         errorutil.FormatTrackingID(ctx),
	}

	payload.FillContext(eventCtx)

	return *eventCtx
}

func (s *AuditService) resolveNonBlockingEvent(ctx context.Context, appID string, payload event.NonBlockingPayload) (*event.Event, error) {
	eventContext := s.makeContext(ctx, appID, payload)
	var seq int64
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		seq, err = s.nextSeq(ctx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return libevent.NewNonBlockingEvent(seq, payload, eventContext), nil
}
