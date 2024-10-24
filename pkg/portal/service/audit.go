package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	libevent "github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalsession "github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type AuditServiceAppService interface {
	Get(id string) (*model.App, error)
}

type AuditService struct {
	Context         context.Context
	RemoteIP        httputil.RemoteIP
	UserAgentString httputil.UserAgentString
	Request         *http.Request

	Apps     AuditServiceAppService
	Authgear *portalconfig.AuthgearConfig

	DenoEndpoint config.DenoEndpoint

	GlobalSQLBuilder  *globaldb.SQLBuilder
	GlobalSQLExecutor *globaldb.SQLExecutor
	GlobalDatabase    *globaldb.Handle

	AuditDatabase *auditdb.WriteHandle

	Clock         clock.Clock
	LoggerFactory *log.Factory
}

func (s *AuditService) Log(app *model.App, payload event.NonBlockingPayload) (err error) {
	if s.AuditDatabase == nil {
		return
	}

	authgearApp, err := s.Apps.Get(s.Authgear.AppID)
	if err != nil {
		return
	}

	cfg := app.Context.Config
	loggerFactory := s.LoggerFactory.ReplaceHooks(
		apierrors.SkipLoggingHook{},
		log.NewDefaultMaskLogHook(),
		config.NewSecretMaskLogHook(cfg.SecretConfig),
		sentry.NewLogHookFromContext(s.Context),
	)
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID

	// AuditSink is app specific.
	// The records MUST have correct app_id.
	// We have construct audit sink with the target app.
	auditSink := newAuditSink(s.Context, app, s.AuditDatabase, loggerFactory)
	// The portal uses its Authgear to deliver hooks.
	// We have construct hook sink with the Authgear app.
	hookSink := newHookSink(s.Context, authgearApp, s.DenoEndpoint, loggerFactory)

	// Use the target app ID.
	e, err := s.resolveNonBlockingEvent(app.ID, payload)
	if err != nil {
		return err
	}

	err = auditSink.ReceiveNonBlockingEvent(e)
	if err != nil {
		return err
	}

	err = hookSink.ReceiveNonBlockingEvent(e)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuditService) nextSeq() (seq int64, err error) {
	builder := s.GlobalSQLBuilder.
		Select(fmt.Sprintf("nextval('%s')", s.GlobalSQLBuilder.TableName("_auth_event_sequence")))
	row, err := s.GlobalSQLExecutor.QueryRowWith(builder)
	if err != nil {
		return
	}
	err = row.Scan(&seq)
	return
}

func (s *AuditService) makeContext(appID string, payload event.Payload) event.Context {
	var userIDStr string
	portalSession := portalsession.GetValidSessionInfo(s.Context)
	if portalSession != nil {
		userIDStr = portalSession.UserID
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)
	// Initialize this to an empty slice so that it is always present in the JSON.
	if preferredLanguageTags == nil {
		preferredLanguageTags = []string{}
	}

	triggeredBy := payload.GetTriggeredBy()

	uiParam := uiparam.GetUIParam(s.Context)
	clientID := uiParam.ClientID
	// This audit context must be constructed here.
	// We cannot use GetAdminAuthzAudit because that is for Admin API to audit context.
	auditCtx := PortalAdminAPIAuthContext{
		Usage:       UsageInternal,
		ActorUserID: userIDStr,
		HTTPReferer: s.Request.Header.Get("Referer"),
	}

	ctx := &event.Context{
		Timestamp: s.Clock.NowUTC().Unix(),
		// We do not populate UserID because the event is not about UserID.
		TriggeredBy:        triggeredBy,
		AuditContext:       auditCtx,
		PreferredLanguages: preferredLanguageTags,
		Language:           "",
		IPAddress:          string(s.RemoteIP),
		UserAgent:          string(s.UserAgentString),
		ClientID:           clientID,
		AppID:              appID,
	}

	payload.FillContext(ctx)

	return *ctx
}

func (s *AuditService) resolveNonBlockingEvent(appID string, payload event.NonBlockingPayload) (*event.Event, error) {
	eventContext := s.makeContext(appID, payload)
	var seq int64
	var err error
	err = s.GlobalDatabase.WithTx(func() error {
		seq, err = s.nextSeq()
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
