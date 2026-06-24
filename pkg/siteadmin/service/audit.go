package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	libevent "github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var AuditServiceLogger = slogutil.NewLogger("siteadmin-audit-service")

// SiteAdminAuditService writes audit log entries for site admin mutations to
// the global audit DB. All records are stored under the portal app ID
// (SITEADMIN_AUTHGEAR_APP_ID); the affected app ID is in the payload.
//
// The log structure mirrors portal audit logs: the actor's user ID is in
// audit_context.actor_user_id (not in context.user_id), and user_agent /
// http_url / http_referer are populated from the HTTP request.
type SiteAdminAuditService struct {
	AuditDatabase     *auditdb.WriteHandle // nil when audit DB is not configured
	SQLBuilder        *auditdb.SQLBuilder  // global (not app-scoped)
	WriteSQLExecutor  *auditdb.WriteSQLExecutor
	Clock             clock.Clock
	AuthgearConfig    *portalconfig.AuthgearConfig
	RemoteIP          httputil.RemoteIP
	UserAgentString   httputil.UserAgentString
	HTTPRequestURL    httputil.HTTPRequestURL
	Request           *http.Request
	GlobalDatabase    *globaldb.Handle
	GlobalSQLBuilder  *globaldb.SQLBuilder
	GlobalSQLExecutor *globaldb.SQLExecutor
}

// nextSeq returns the next value of _auth_event_sequence from the global DB,
// mirroring the portal audit service pattern.
func (s *SiteAdminAuditService) nextSeq(ctx context.Context) (seq int64, err error) {
	builder := s.GlobalSQLBuilder.
		Select(fmt.Sprintf("nextval('%s')", s.GlobalSQLBuilder.TableName("_auth_event_sequence")))
	row, err := s.GlobalSQLExecutor.QueryRowWith(ctx, builder)
	if err != nil {
		return
	}
	err = row.Scan(&seq)
	return
}

// LogEvent writes one audit log entry under the portal app ID.
// The affected appID is already embedded in the payload; it is passed here
// only for informational context.
// If the audit database is not configured the call is a no-op.
func (s *SiteAdminAuditService) LogEvent(ctx context.Context, appID string, payload event.NonBlockingPayload) error {
	if s.AuditDatabase == nil {
		return nil
	}

	var actorUserID string
	if info := session.GetValidSessionInfo(ctx); info != nil {
		actorUserID = info.UserID
	}

	// All siteadmin audit records belong to the portal app in the DB.
	portalAppID := s.AuthgearConfig.AppID

	// Get a real sequence number, mirroring portal audit logs.
	var seq int64
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		var e error
		seq, e = s.nextSeq(ctx)
		return e
	})
	if err != nil {
		return err
	}

	// Build audit context mirroring portal/service.AuditService.makeContext:
	// actor goes in audit_context.actor_user_id, not in context.user_id.
	referer := ""
	if s.Request != nil {
		referer = s.Request.Header.Get("Referer")
	}
	auditCtx := event.NewAuditContext(string(s.HTTPRequestURL), map[string]any{
		"usage":         portalservice.UsageInternal,
		"actor_user_id": actorUserID,
		"http_referer":  referer,
	})

	now := s.Clock.NowUTC()
	eventCtx := event.Context{
		Timestamp:          now.Unix(),
		TriggeredBy:        payload.GetTriggeredBy(),
		UserID:             nil, // actor is in audit_context, not user_id
		AppID:              portalAppID,
		IPAddress:          string(s.RemoteIP),
		UserAgent:          string(s.UserAgentString),
		AuditContext:       auditCtx,
		PreferredLanguages: []string{},
	}

	e := libevent.NewNonBlockingEvent(seq, payload, eventCtx)

	logEntry, err := audit.NewLog(e)
	if err != nil {
		return err
	}

	return s.AuditDatabase.WithTx(ctx, func(ctx context.Context) error {
		store := &audit.WriteStore{
			SQLBuilder:  s.SQLBuilder.WithAppID(portalAppID),
			SQLExecutor: s.WriteSQLExecutor,
		}
		return store.PersistLog(ctx, logEntry)
	})
}
