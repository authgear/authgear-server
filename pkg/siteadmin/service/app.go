package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	sq "github.com/Masterminds/squirrel"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

const maxPageSize = 20

// ---- Narrow interfaces -------------------------------------------------------

type AppServiceDatabase interface {
	WithTx(ctx context.Context, do func(ctx context.Context) error) error
}

type AppServiceConfigSourceStore interface {
	GetDatabaseSourceByAppID(ctx context.Context, appID string) (*configsource.DatabaseSource, error)
	CountAll(ctx context.Context) (int, error)
	ListPaged(ctx context.Context, limit int, offset int) ([]*configsource.DatabaseSource, error)
	GetManyByAppIDs(ctx context.Context, appIDs []string) ([]*configsource.DatabaseSource, error)
}

type AppServiceOwnerStore interface {
	GetOwnerByAppID(ctx context.Context, appID string) (string, error)
	GetOwnersByAppIDs(ctx context.Context, appIDs []string) (map[string]string, error)
	CountAppsByOwnerUserID(ctx context.Context, userID string) (int, error)
	ListAppIDsByOwnerUserIDPaged(ctx context.Context, userID string, limit int, offset int) ([]string, error)
}

type AppServiceAdminAPI interface {
	SelfDirector(ctx context.Context, actorUserID string, usage portalservice.Usage) (func(*http.Request), error)
}

type AppServiceHTTPClient struct {
	*http.Client
}

// ---- AppOwnerStore -----------------------------------------------------------

// AppOwnerStore is a minimal struct that queries _portal_app_collaborator for
// owner relationships.
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

func (s *AppOwnerStore) GetOwnersByAppIDs(ctx context.Context, appIDs []string) (map[string]string, error) {
	if len(appIDs) == 0 {
		return map[string]string{}, nil
	}

	q := s.SQLBuilder.
		Select("app_id", "user_id").
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where(sq.Eq{"app_id": appIDs, "role": "owner"})

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]string, len(appIDs))
	for rows.Next() {
		var appID, userID string
		if err := rows.Scan(&appID, &userID); err != nil {
			return nil, err
		}
		m[appID] = userID
	}
	return m, nil
}

func (s *AppOwnerStore) CountAppsByOwnerUserID(ctx context.Context, userID string) (int, error) {
	q := s.SQLBuilder.
		Select("COUNT(*)").
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("user_id = ? AND role = ?", userID, "owner")

	scanner, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return 0, err
	}

	var count int
	if err := scanner.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *AppOwnerStore) ListAppIDsByOwnerUserIDPaged(ctx context.Context, userID string, limit int, offset int) ([]string, error) {
	q := s.SQLBuilder.
		Select("app_id").
		From(s.SQLBuilder.TableName("_portal_app_collaborator")).
		Where("user_id = ? AND role = ?", userID, "owner").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appIDs []string
	for rows.Next() {
		var appID string
		if err := rows.Scan(&appID); err != nil {
			return nil, err
		}
		appIDs = append(appIDs, appID)
	}
	return appIDs, nil
}

// ---- AppService ----------------------------------------------------------------

type ListAppsParams struct {
	Page       int
	PageSize   int
	AppID      string
	OwnerEmail string
}

type ListAppsResult struct {
	Apps       []siteadmin.App
	TotalCount int
}

type AppService struct {
	GlobalDatabase    AppServiceDatabase
	ConfigSourceStore AppServiceConfigSourceStore
	OwnerStore        AppServiceOwnerStore
	AdminAPI          AppServiceAdminAPI
	AuditDatabase     *auditdb.ReadHandle
	AuditStore        *analytic.AuditDBReadStore
	HTTPClient        AppServiceHTTPClient
	Clock             clock.Clock
}

func (s *AppService) ListApps(ctx context.Context, params ListAppsParams) (*ListAppsResult, error) {
	if params.PageSize <= 0 || params.PageSize > maxPageSize {
		params.PageSize = maxPageSize
	}

	switch {
	case params.OwnerEmail != "" && params.AppID == "":
		return s.listAppsByOwnerEmail(ctx, params)
	case params.AppID != "":
		return s.listAppsByAppID(ctx, params)
	default:
		return s.listAppsPaged(ctx, params)
	}
}

// listAppsByOwnerEmail resolves the owner_email to a user ID via Admin API,
// then fetches apps owned by that user using DB-level pagination.
//
// Assumption: getUsersByStandardAttribute returns at most one user because email
// is unique within an Authgear app. We therefore treat the result as a single
// user and apply LIMIT/OFFSET directly against that user's owned apps.
func (s *AppService) listAppsByOwnerEmail(ctx context.Context, params ListAppsParams) (*ListAppsResult, error) {
	// Admin API call — must happen outside a DB transaction.
	userIDs, err := s.findUserIDsByEmail(ctx, params.OwnerEmail)
	if err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		return &ListAppsResult{Apps: []siteadmin.App{}, TotalCount: 0}, nil
	}

	// Email is unique — take the first (and expected only) match.
	userID := userIDs[0]

	var totalCount int
	var sources []*configsource.DatabaseSource
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		var e error
		totalCount, e = s.OwnerStore.CountAppsByOwnerUserID(ctx, userID)
		if e != nil {
			return e
		}
		if totalCount == 0 {
			return nil
		}
		offset := (params.Page - 1) * params.PageSize
		appIDs, e := s.OwnerStore.ListAppIDsByOwnerUserIDPaged(ctx, userID, params.PageSize, offset)
		if e != nil {
			return e
		}
		sources, e = s.ConfigSourceStore.GetManyByAppIDs(ctx, appIDs)
		return e
	})
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &ListAppsResult{Apps: []siteadmin.App{}, TotalCount: 0}, nil
	}

	apps := make([]siteadmin.App, len(sources))
	for i, src := range sources {
		apps[i] = siteadmin.App{
			Id:         src.AppID,
			OwnerEmail: params.OwnerEmail, // already known — no extra GraphQL call
			Plan:       src.PlanName,
			CreatedAt:  src.CreatedAt,
		}
	}

	return &ListAppsResult{Apps: apps, TotalCount: totalCount}, nil
}

// listAppsByAppID fetches a single app and optionally verifies owner_email.
func (s *AppService) listAppsByAppID(ctx context.Context, params ListAppsParams) (*ListAppsResult, error) {
	var src *configsource.DatabaseSource
	var ownerUserID string
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		var e error
		src, e = s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, params.AppID)
		if e != nil {
			return e
		}
		ownerUserID, e = s.OwnerStore.GetOwnerByAppID(ctx, params.AppID)
		if errors.Is(e, ErrOwnerNotFound) {
			return nil
		}
		return e
	})
	if err != nil {
		if errors.Is(err, configsource.ErrAppNotFound) {
			return &ListAppsResult{Apps: []siteadmin.App{}, TotalCount: 0}, nil
		}
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

	if params.OwnerEmail != "" && !strings.EqualFold(ownerEmail, params.OwnerEmail) {
		return &ListAppsResult{Apps: []siteadmin.App{}, TotalCount: 0}, nil
	}

	app := siteadmin.App{
		Id:         src.AppID,
		OwnerEmail: ownerEmail,
		Plan:       src.PlanName,
		CreatedAt:  src.CreatedAt,
	}
	return &ListAppsResult{Apps: []siteadmin.App{app}, TotalCount: 1}, nil
}

// listAppsPaged uses DB-level pagination; resolves emails only for the current page.
func (s *AppService) listAppsPaged(ctx context.Context, params ListAppsParams) (*ListAppsResult, error) {
	var totalCount int
	var sources []*configsource.DatabaseSource
	var ownerMap map[string]string
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		var e error
		totalCount, e = s.ConfigSourceStore.CountAll(ctx)
		if e != nil {
			return e
		}
		offset := (params.Page - 1) * params.PageSize
		sources, e = s.ConfigSourceStore.ListPaged(ctx, params.PageSize, offset)
		if e != nil {
			return e
		}
		appIDs := make([]string, len(sources))
		for i, src := range sources {
			appIDs[i] = src.AppID
		}
		ownerMap, e = s.OwnerStore.GetOwnersByAppIDs(ctx, appIDs)
		return e
	})
	if err != nil {
		return nil, err
	}

	// Admin API call — outside the DB transaction.
	userIDs := uniqueValues(ownerMap)
	emailMap, err := s.resolveUserEmails(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	apps := make([]siteadmin.App, len(sources))
	for i, src := range sources {
		ownerUserID := ownerMap[src.AppID]
		apps[i] = siteadmin.App{
			Id:         src.AppID,
			OwnerEmail: emailMap[ownerUserID],
			Plan:       src.PlanName,
			CreatedAt:  src.CreatedAt,
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

// findUserIDsByEmail calls Admin API getUsersByStandardAttribute to find users
// matching the given email. Returns their raw (non-global) user IDs.
func (s *AppService) findUserIDsByEmail(ctx context.Context, email string) ([]string, error) {
	params := graphqlutil.DoParams{
		OperationName: "getUsersByStandardAttribute",
		Query: `
		query getUsersByStandardAttribute($name: String!, $value: String!) {
			users: getUsersByStandardAttribute(attributeName: $name, attributeValue: $value) {
				id
			}
		}
		`,
		Variables: map[string]interface{}{
			"name":  "email",
			"value": email,
		},
	}

	r, err := http.NewRequestWithContext(ctx, "POST", "/graphql", nil)
	if err != nil {
		return nil, err
	}

	actorUserID := session.GetValidSessionInfo(ctx).UserID
	director, err := s.AdminAPI.SelfDirector(ctx, actorUserID, portalservice.UsageInternal)
	if err != nil {
		return nil, err
	}
	director(r)

	result, err := graphqlutil.HTTPDo(s.HTTPClient.Client, r, params)
	if err != nil {
		return nil, err
	}
	if result.HasErrors() {
		return nil, fmt.Errorf("failed to search users by email: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	users := data["users"].([]interface{})

	ids := make([]string, 0, len(users))
	for _, u := range users {
		userNode, ok := u.(map[string]interface{})
		if !ok {
			continue
		}
		globalID, _ := userNode["id"].(string)
		resolved := relay.FromGlobalID(globalID)
		if resolved == nil || resolved.ID == "" {
			// relay.FromGlobalID failed to parse — skip this entry.
			continue
		}
		ids = append(ids, resolved.ID)
	}
	return ids, nil
}

// resolveUserEmails batch-fetches emails for the given user IDs via Admin API.
func (s *AppService) resolveUserEmails(ctx context.Context, userIDs []string) (map[string]string, error) {
	if len(userIDs) == 0 {
		return map[string]string{}, nil
	}

	globalIDs := make([]string, len(userIDs))
	for i, id := range userIDs {
		globalIDs[i] = relay.ToGlobalID("User", id)
	}

	params := graphqlutil.DoParams{
		OperationName: "getUserNodes",
		Query: `
		query getUserNodes($ids: [ID!]!) {
			nodes(ids: $ids) {
				... on User {
					id
					standardAttributes
				}
			}
		}
		`,
		Variables: map[string]interface{}{
			"ids": globalIDs,
		},
	}

	r, err := http.NewRequestWithContext(ctx, "POST", "/graphql", nil)
	if err != nil {
		return nil, err
	}

	actorUserID := session.GetValidSessionInfo(ctx).UserID
	director, err := s.AdminAPI.SelfDirector(ctx, actorUserID, portalservice.UsageInternal)
	if err != nil {
		return nil, err
	}
	director(r)

	result, err := graphqlutil.HTTPDo(s.HTTPClient.Client, r, params)
	if err != nil {
		return nil, err
	}
	if result.HasErrors() {
		return nil, fmt.Errorf("failed to resolve user emails: %v", result.Errors)
	}

	emailMap := make(map[string]string, len(userIDs))
	data := result.Data.(map[string]interface{})
	nodes := data["nodes"].([]interface{})
	for _, node := range nodes {
		userNode, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		globalID, _ := userNode["id"].(string)
		resolvedID := relay.FromGlobalID(globalID)
		if resolvedID == nil || resolvedID.ID == "" {
			continue
		}
		attrs, ok := userNode["standardAttributes"].(map[string]interface{})
		if !ok {
			continue
		}
		email, _ := attrs["email"].(string)
		emailMap[resolvedID.ID] = email
	}
	return emailMap, nil
}

func uniqueValues(m map[string]string) []string {
	seen := make(map[string]struct{}, len(m))
	result := make([]string, 0, len(m))
	for _, v := range m {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
