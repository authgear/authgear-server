package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

const MaxPageSize = 20

// ---- Narrow interfaces -------------------------------------------------------

type AppServiceDatabase interface {
	WithTx(ctx context.Context, do func(ctx context.Context) error) error
}

type AppServiceConfigSourceStore interface {
	GetDatabaseSourceByAppID(ctx context.Context, appID string) (*configsource.DatabaseSource, error)
	CountAll(ctx context.Context) (int, error)
	ListPaged(ctx context.Context, limit uint64, offset uint64) ([]*configsource.DatabaseSource, error)
	GetManyByAppIDs(ctx context.Context, appIDs []string) ([]*configsource.DatabaseSource, error)
}

type AppServiceOwnerStore interface {
	GetOwnerByAppID(ctx context.Context, appID string) (string, error)
	ListAppsWithStats(ctx context.Context, params ListAppsStoreParams) ([]AppStoreRow, int, error)
}

// ---- Store types -------------------------------------------------------------

// ListAppsStoreParams parameterises the unified ListAppsWithStats query.
type ListAppsStoreParams struct {
	Page           uint64
	PageSize       uint64
	AppID          string                        // optional; if set, WHERE cs.app_id = ?
	PlanName       string                        // optional; if set, WHERE cs.plan_name = ?
	OwnerUserID    string                        // optional; if set, WHERE ac.user_id = ?
	Sort           siteadmin.ListAppsParamsSort  // "created_at" | "mau"
	Order          siteadmin.ListAppsParamsOrder // "asc" | "desc"
	LastMonthStart time.Time                     // exact start_time value for the usage record JOIN
}

// AppStoreRow is a single row returned by ListAppsWithStats.
type AppStoreRow struct {
	AppID        string
	PlanName     string
	CreatedAt    time.Time
	OwnerUserID  string // empty string if no owner row
	LastMonthMAU int    // COALESCE(ur.count, 0)
}

// ---- AppOwnerStore -----------------------------------------------------------

// AppOwnerStore queries _portal_app_collaborator and related tables for owner
// relationships and aggregated app statistics.
type AppOwnerStore struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

var ErrOwnerNotFound = errors.New("app owner not found")

func (s *AppOwnerStore) GetOwnerByAppID(ctx context.Context, appID string) (string, error) {
	q := s.SQLBuilder.
		Select("user_id").
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("app_id = ? AND role = ?", appID, "owner").
		Limit(1)

	scanner, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return "", err
	}

	var userID string
	if err := scanner.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrOwnerNotFound
		}
		return "", err
	}
	return userID, nil
}

// ListAppsWithStats issues a three-table LEFT JOIN across _portal_config_source,
// _portal_app_collaborator, and _portal_usage_record. It supports all filter,
// sort, and pagination combinations in a single DB round-trip pair (count + page).
func (s *AppOwnerStore) ListAppsWithStats(ctx context.Context, params ListAppsStoreParams) ([]AppStoreRow, int, error) {
	configTable := s.SQLBuilder.TableName("_portal_config_source")
	collaboratorTable := s.SQLBuilder.TableName("_portal_app_collaborator")
	usageTable := s.SQLBuilder.TableName("_portal_usage_record")

	// withFilters applies optional WHERE clauses shared by both queries.
	withFilters := func(q sq.SelectBuilder) sq.SelectBuilder {
		if params.PlanName != "" {
			q = q.Where("cs.plan_name = ?", params.PlanName)
		}
		if params.OwnerUserID != "" {
			q = q.Where("ac.user_id = ?", params.OwnerUserID)
		}
		if params.AppID != "" {
			q = q.Where("cs.app_id = ?", params.AppID)
		}
		return q
	}

	// Count query — no usage record join needed.
	countQ := withFilters(
		s.SQLBuilder.Select("COUNT(*)").
			From(configTable + " cs").
			LeftJoin(collaboratorTable + " ac ON ac.app_id = cs.app_id AND ac.role = 'owner'"),
	)
	scanner, err := s.SQLExecutor.QueryRowWith(ctx, countQ)
	if err != nil {
		return nil, 0, err
	}
	var totalCount int
	if err := scanner.Scan(&totalCount); err != nil {
		return nil, 0, err
	}
	if totalCount == 0 {
		return nil, 0, nil
	}

	// Page query.
	dirSQL := "DESC"
	if params.Order == siteadmin.Asc {
		dirSQL = "ASC"
	}
	var primaryExpr string
	switch params.Sort {
	case siteadmin.Mau:
		primaryExpr = "COALESCE(ur.count, 0)"
	default:
		primaryExpr = "cs.created_at"
	}
	orderExpr := primaryExpr + " " + dirSQL + ", cs.app_id ASC"

	offset := (params.Page - 1) * params.PageSize
	pageQ := withFilters(
		s.SQLBuilder.Select(
			"cs.app_id",
			"cs.plan_name",
			"cs.created_at",
			"COALESCE(ac.user_id, '') AS owner_user_id",
			"COALESCE(ur.count, 0) AS last_month_mau",
		).
			From(configTable+" cs").
			LeftJoin(collaboratorTable+" ac ON ac.app_id = cs.app_id AND ac.role = 'owner'").
			LeftJoin(
				usageTable+" ur ON ur.app_id = cs.app_id AND ur.name = ? AND ur.period = ? AND ur.start_time = ?",
				usage.RecordNameActiveUser, periodical.Monthly, params.LastMonthStart,
			).
			OrderBy(orderExpr).
			Limit(params.PageSize).
			Offset(offset),
	)

	rows, err := s.SQLExecutor.QueryWith(ctx, pageQ)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []AppStoreRow
	for rows.Next() {
		var row AppStoreRow
		if err := rows.Scan(&row.AppID, &row.PlanName, &row.CreatedAt, &row.OwnerUserID, &row.LastMonthMAU); err != nil {
			return nil, 0, err
		}
		result = append(result, row)
	}
	return result, totalCount, nil
}

// ---- AppService ----------------------------------------------------------------

type ListAppsParams struct {
	Page       uint64
	PageSize   uint64
	AppID      string
	OwnerEmail string
	Plan       string                        // exact plan name filter
	Sort       siteadmin.ListAppsParamsSort  // "created_at" (default) | "mau"
	Order      siteadmin.ListAppsParamsOrder // "desc" (default) | "asc"
}

type ListAppsResult struct {
	Apps       []siteadmin.App
	TotalCount int
}

type AppService struct {
	GlobalDatabase    AppServiceDatabase
	ConfigSourceStore AppServiceConfigSourceStore
	OwnerStore        AppServiceOwnerStore
	AdminAPI          *AdminAPIService
	AuditDatabase     *auditdb.ReadHandle
	AuditStore        *analytic.AuditDBReadStore
	Clock             clock.Clock
}

func (s *AppService) ListApps(ctx context.Context, params ListAppsParams) (*ListAppsResult, error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 || params.PageSize > MaxPageSize {
		params.PageSize = MaxPageSize
	}
	if !params.Sort.Valid() {
		params.Sort = siteadmin.CreatedAt
	}
	if !params.Order.Valid() {
		params.Order = siteadmin.Desc
	}

	// 1. Resolve owner_email → owner_user_id via Admin API (outside DB transaction).
	var ownerUserID string
	if params.OwnerEmail != "" {
		userIDs, err := s.AdminAPI.FindUserIDsByEmail(ctx, params.OwnerEmail)
		if err != nil {
			return nil, err
		}
		if len(userIDs) == 0 {
			return &ListAppsResult{Apps: []siteadmin.App{}, TotalCount: 0}, nil
		}
		ownerUserID = userIDs[0]
	}

	// 2. Compute last-month start.
	// Go normalises time.Date(y, 0, ...) → time.Date(y-1, 12, ...) when m = January.
	now := s.Clock.NowUTC()
	y, m, _ := now.Date()
	lastMonthStart := time.Date(y, m-1, 1, 0, 0, 0, 0, time.UTC)

	// 3. Unified DB query.
	var rows []AppStoreRow
	var totalCount int
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		var e error
		rows, totalCount, e = s.OwnerStore.ListAppsWithStats(ctx, ListAppsStoreParams{
			Page:           params.Page,
			PageSize:       params.PageSize,
			AppID:          params.AppID,
			PlanName:       params.Plan,
			OwnerUserID:    ownerUserID,
			Sort:           params.Sort,
			Order:          params.Order,
			LastMonthStart: lastMonthStart,
		})
		return e
	})
	if err != nil {
		return nil, err
	}

	// 4. Resolve owner emails for the current page via Admin API.
	// Optimisation: when owner_email was the filter, every row shares the same
	// owner_user_id — reuse the known email without an extra API call.
	var emailMap map[string]string
	if params.OwnerEmail != "" && ownerUserID != "" {
		emailMap = map[string]string{ownerUserID: params.OwnerEmail}
	} else {
		seen := make(map[string]struct{}, len(rows))
		uniqueIDs := make([]string, 0, len(rows))
		for _, row := range rows {
			if row.OwnerUserID != "" {
				if _, ok := seen[row.OwnerUserID]; !ok {
					seen[row.OwnerUserID] = struct{}{}
					uniqueIDs = append(uniqueIDs, row.OwnerUserID)
				}
			}
		}
		emailMap, err = s.AdminAPI.ResolveUserEmails(ctx, uniqueIDs)
		if err != nil {
			return nil, err
		}
	}

	// 5. Build response.
	apps := make([]siteadmin.App, len(rows))
	for i, row := range rows {
		apps[i] = siteadmin.App{
			Id:           row.AppID,
			OwnerEmail:   emailMap[row.OwnerUserID],
			Plan:         row.PlanName,
			CreatedAt:    row.CreatedAt,
			LastMonthMau: row.LastMonthMAU,
		}
	}
	return &ListAppsResult{Apps: apps, TotalCount: totalCount}, nil
}

func (s *AppService) GetApp(ctx context.Context, appID string) (*siteadmin.AppDetail, error) {
	var src *configsource.DatabaseSource
	var ownerUserID string
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		var e error
		src, e = s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
		if e != nil {
			return e
		}
		ownerUserID, e = s.OwnerStore.GetOwnerByAppID(ctx, appID)
		if errors.Is(e, ErrOwnerNotFound) {
			return nil
		}
		return e
	})
	if err != nil {
		return nil, err
	}

	// Admin API call — outside the DB transaction.
	ownerEmail := ""
	if ownerUserID != "" {
		emailMap, err := s.resolveUserEmails(ctx, []string{ownerUserID})
		if err != nil {
			return nil, err
		}
		ownerEmail = emailMap[ownerUserID]
	}

	userCount, err := s.fetchTotalUserCount(ctx, appID)
	if err != nil {
		return nil, err
	}

	return &siteadmin.AppDetail{
		Id:         src.AppID,
		OwnerEmail: ownerEmail,
		Plan:       src.PlanName,
		CreatedAt:  src.CreatedAt,
		UserCount:  userCount,
	}, nil
}

// ---- Private helpers ---------------------------------------------------------

// fetchTotalUserCount returns the cumulative total user count for the given app
// from the audit DB. Returns 0 if the audit DB is not configured or no data exists
// for yesterday. Mirrors the pattern in analytic.ChartService.GetTotalUserCountChart.
func (s *AppService) fetchTotalUserCount(ctx context.Context, appID string) (int, error) {
	if s.AuditDatabase == nil {
		return 0, nil
	}

	now := s.Clock.NowUTC()
	yesterday := timeutil.TruncateToDate(now).AddDate(0, 0, -1)

	var userCount int
	err := s.AuditDatabase.WithTx(ctx, func(ctx context.Context) error {
		c, err := s.AuditStore.GetAnalyticCountByType(ctx, appID, analytic.CumulativeUserCountType, &yesterday)
		if errors.Is(err, analytic.ErrAnalyticCountNotFound) {
			userCount = 0
			return nil
		}
		if err != nil {
			return err
		}
		userCount = c.Count
		return nil
	})
	return userCount, err
}

// resolveUserEmails batch-fetches emails for the given user IDs via Admin API.
func (s *AppService) resolveUserEmails(ctx context.Context, userIDs []string) (map[string]string, error) {
	return s.AdminAPI.ResolveUserEmails(ctx, userIDs)
}
