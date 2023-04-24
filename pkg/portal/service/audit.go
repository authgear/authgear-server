package service

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	libevent "github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalsession "github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type AuditStore interface {
	NextSequenceNumber() (int64, error)
}

type AuditService struct {
	Context                   context.Context
	RemoteIP                  httputil.RemoteIP
	UserAgentString           httputil.UserAgentString
	Clock                     clock.Clock
	DBPool                    *db.Pool
	DatabaseEnvironmentConfig *config.DatabaseEnvironmentConfig
	LoggerFactory             *log.Factory
}

func (s *AuditService) Log(app *model.App, payload event.NonBlockingPayload) (err error) {
	cfg := app.Context.Config
	loggerFactory := s.LoggerFactory.ReplaceHooks(
		log.NewDefaultMaskLogHook(),
		config.NewSecretMaskLogHook(cfg.SecretConfig),
		sentry.NewLogHookFromContext(s.Context),
	)
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID
	auditDatabaseCredentials := cfg.SecretConfig.LookupData(
		config.AuditDatabaseCredentialsKey).(*config.AuditDatabaseCredentials)
	appDbCredentials := cfg.SecretConfig.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials)
	appDatabase := appdb.NewHandle(
		s.Context,
		s.DBPool,
		s.DatabaseEnvironmentConfig,
		appDbCredentials,
		loggerFactory,
	)
	appSQLBuilder := appdb.NewSQLBuilder(appDbCredentials)
	appSQLExecutor := appdb.NewSQLExecutor(s.Context, appDatabase)
	auditdbSQLBuilderApp := auditdb.NewSQLBuilderApp(auditDatabaseCredentials, cfg.AppConfig.ID)
	auditWriteHandle := auditdb.NewWriteHandle(
		s.Context,
		s.DBPool,
		s.DatabaseEnvironmentConfig,
		auditDatabaseCredentials,
		loggerFactory,
	)
	writeSQLExecutor := auditdb.NewWriteSQLExecutor(s.Context, auditWriteHandle)
	writeStore := &audit.WriteStore{
		SQLBuilder:  auditdbSQLBuilderApp,
		SQLExecutor: writeSQLExecutor,
	}
	auditSink := &audit.Sink{
		Logger:   audit.NewLogger(loggerFactory),
		Database: auditWriteHandle,
		Store:    writeStore,
	}

	return appDatabase.WithTx(func() error {
		eventStore := libevent.NewStoreImpl(
			appSQLBuilder,
			appSQLExecutor,
		)
		e, err := s.resolveNonBlockingEvent(eventStore, payload)
		if err != nil {
			return err
		}
		return auditSink.ReceiveNonBlockingEvent(e)
	})
}

func (s *AuditService) nextSeq(store AuditStore) (seq int64, err error) {
	seq, err = store.NextSequenceNumber()
	if err != nil {
		return
	}
	return
}

func (s *AuditService) makeContext(payload event.Payload) event.Context {
	portalSession := portalsession.GetValidSessionInfo(s.Context)
	var userID *string

	if portalSession != nil {
		userID = &portalSession.UserID
	}

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

	triggeredBy := payload.GetTriggeredBy()

	uiParam := uiparam.GetUIParam(s.Context)
	auditCtx := adminauthz.GetAdminAuthzAudit(s.Context)
	clientID := uiParam.ClientID

	ctx := &event.Context{
		Timestamp:          s.Clock.NowUTC().Unix(),
		UserID:             userID,
		TriggeredBy:        triggeredBy,
		AuditContext:       auditCtx,
		PreferredLanguages: preferredLanguageTags,
		Language:           "",
		IPAddress:          string(s.RemoteIP),
		UserAgent:          string(s.UserAgentString),
		ClientID:           clientID,
	}

	payload.FillContext(ctx)

	return *ctx
}

func (s *AuditService) resolveNonBlockingEvent(store AuditStore, payload event.NonBlockingPayload) (*event.Event, error) {
	eventContext := s.makeContext(payload)
	seq, err := s.nextSeq(store)
	if err != nil {
		return nil, err
	}

	return libevent.NewNonBlockingEvent(seq, payload, eventContext), nil
}
