package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
)

const AuditLogsMaxPageSize uint64 = 100

// ---- Domain types -----------------------------------------------------------

type AuditLogEntry struct {
	ID            string
	CreatedAt     time.Time
	ActivityType  string
	IPAddress     string
	UserAgent     string
	ActorUserID   string
	AffectedAppID string
}

type AuditLogEntryDetail struct {
	AuditLogEntry
	Data map[string]any
}

type ListAuditLogsParams struct {
	Page          uint64
	PageSize      uint64
	AffectedAppID string
	Order         siteadmin.OrderDirection // "asc" | "desc"; defaults to "desc"
}

type ListAuditLogsResult struct {
	Entries    []AuditLogEntry
	TotalCount int
}

// ---- Store ------------------------------------------------------------------

type SiteAdminAuditLogStoreIface interface {
	Count(ctx context.Context, affectedAppID string) (int, error)
	List(ctx context.Context, affectedAppID string, order siteadmin.OrderDirection, limit, offset uint64) ([]AuditLogEntry, error)
	Get(ctx context.Context, id string) (*AuditLogEntryDetail, error)
}

type SiteAdminAuditLogStore struct {
	SQLBuilder     *auditdb.SQLBuilder
	SQLExecutor    *auditdb.ReadSQLExecutor
	AuthgearConfig *portalconfig.AuthgearConfig
}

func (s *SiteAdminAuditLogStore) scopedBuilder() *auditdb.SQLBuilderApp {
	return s.SQLBuilder.WithAppID(s.AuthgearConfig.AppID)
}

func (s *SiteAdminAuditLogStore) Count(ctx context.Context, affectedAppID string) (int, error) {
	sb := s.scopedBuilder()
	q := sb.
		Select("COUNT(*)").
		From(sb.TableName("_audit_log")).
		Where("activity_type LIKE ?", "site_admin.%")
	if affectedAppID != "" {
		q = q.Where("data->'payload'->>'app_id' = ?", affectedAppID)
	}

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return 0, err
	}
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *SiteAdminAuditLogStore) List(ctx context.Context, affectedAppID string, order siteadmin.OrderDirection, limit, offset uint64) ([]AuditLogEntry, error) {
	dir := "DESC"
	if order == siteadmin.Asc {
		dir = "ASC"
	}

	sb := s.scopedBuilder()
	q := sb.
		Select(
			"id", "created_at", "activity_type",
			"ip_address::text", "user_agent",
			"data->'context'->'audit_context'->>'actor_user_id'",
			"data->'payload'->>'app_id'",
		).
		From(sb.TableName("_audit_log")).
		Where("activity_type LIKE ?", "site_admin.%").
		OrderBy("created_at " + dir).
		Limit(limit).
		Offset(offset)
	if affectedAppID != "" {
		q = q.Where("data->'payload'->>'app_id' = ?", affectedAppID)
	}

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AuditLogEntry
	for rows.Next() {
		var e AuditLogEntry
		var ipAddress, userAgent, actorUserID, affectedAppIDVal sql.NullString
		if err := rows.Scan(&e.ID, &e.CreatedAt, &e.ActivityType,
			&ipAddress, &userAgent, &actorUserID, &affectedAppIDVal); err != nil {
			return nil, err
		}
		e.IPAddress = ipAddress.String
		e.UserAgent = userAgent.String
		e.ActorUserID = actorUserID.String
		e.AffectedAppID = affectedAppIDVal.String
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *SiteAdminAuditLogStore) Get(ctx context.Context, id string) (*AuditLogEntryDetail, error) {
	sb := s.scopedBuilder()
	q := sb.
		Select(
			"id", "created_at", "activity_type",
			"ip_address::text", "user_agent",
			"data->'context'->'audit_context'->>'actor_user_id'",
			"data->'payload'->>'app_id'",
			"data",
		).
		From(sb.TableName("_audit_log")).
		Where("id = ?", id).
		Where("activity_type LIKE ?", "site_admin.%")

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	var detail AuditLogEntryDetail
	var ipAddress, userAgent, actorUserID, affectedAppIDVal sql.NullString
	var dataBytes []byte
	if err := row.Scan(&detail.ID, &detail.CreatedAt, &detail.ActivityType,
		&ipAddress, &userAgent, &actorUserID, &affectedAppIDVal, &dataBytes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.NewNotFound("audit log not found")
		}
		return nil, err
	}
	detail.IPAddress = ipAddress.String
	detail.UserAgent = userAgent.String
	detail.ActorUserID = actorUserID.String
	detail.AffectedAppID = affectedAppIDVal.String

	if err := json.Unmarshal(dataBytes, &detail.Data); err != nil {
		return nil, err
	}
	return &detail, nil
}

// ---- Service ----------------------------------------------------------------

type SiteAdminAuditReadDatabase interface {
	ReadOnly(ctx context.Context, do func(context.Context) error) error
}

type SiteAdminAuditReadService struct {
	AuditDatabase SiteAdminAuditReadDatabase // nil when audit DB is not configured
	Store         SiteAdminAuditLogStoreIface
}

func (s *SiteAdminAuditReadService) ListAuditLogs(ctx context.Context, params ListAuditLogsParams) (*ListAuditLogsResult, error) {
	if s.AuditDatabase == nil {
		return &ListAuditLogsResult{Entries: []AuditLogEntry{}}, nil
	}

	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 || params.PageSize > AuditLogsMaxPageSize {
		params.PageSize = AuditLogsMaxPageSize
	}
	if !params.Order.Valid() {
		params.Order = siteadmin.Desc
	}

	var entries []AuditLogEntry
	var totalCount int
	err := s.AuditDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		var e error
		totalCount, e = s.Store.Count(ctx, params.AffectedAppID)
		if e != nil {
			return e
		}
		if totalCount == 0 {
			return nil
		}
		offset := (params.Page - 1) * params.PageSize
		entries, e = s.Store.List(ctx, params.AffectedAppID, params.Order, params.PageSize, offset)
		return e
	})
	if err != nil {
		return nil, err
	}

	if entries == nil {
		entries = []AuditLogEntry{}
	}
	return &ListAuditLogsResult{Entries: entries, TotalCount: totalCount}, nil
}

func (s *SiteAdminAuditReadService) GetAuditLog(ctx context.Context, id string) (*AuditLogEntryDetail, error) {
	if s.AuditDatabase == nil {
		return nil, apierrors.NewNotFound("audit log not found")
	}

	var entry *AuditLogEntryDetail
	err := s.AuditDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		var e error
		entry, e = s.Store.Get(ctx, id)
		return e
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}
