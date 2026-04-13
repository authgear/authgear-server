# Part 3: Site Admin API — Apps API Real Data

## Context

Replace dummy data in `AppsListHandler` and `AppGetHandler` with real data from the
global database and the portal Admin API. The two endpoints affected are:

- `GET /api/v1/apps` — lists all apps with optional filtering and pagination
- `GET /api/v1/apps/{app_id}` — returns detailed info for one app

**Key data sources:**

| Field | Source |
|---|---|
| `app_id`, `plan`, `created_at` | `_portal_config_source` table |
| owner `user_id` | `_portal_app_collaborator` table |
| `owner_email` | Admin API GraphQL (`getUserNodes` / `getUsersByStandardAttribute`) |
| `user_count` (AppDetail only) | `_audit_analytic_count` table — `CumulativeUserCountType` for yesterday |

**Design decisions:**
- Owner email is NOT stored in the database; it is resolved via Admin API GraphQL.
- Filtering by `owner_email` goes through the Admin API first
  (`getUsersByStandardAttribute`) to find matching user IDs, then queries the collaborator
  table for their owned apps. This avoids loading all user emails into memory.
- Pagination is pushed to the database level (SQL `LIMIT`/`OFFSET`) for the no-filter case.
  Email resolution is limited to the current page (e.g., 20 rows), not all apps.
- A new `pkg/siteadmin/service/` package holds the business logic.

---

## Architecture Overview

```
AppsListHandler / AppGetHandler (transport)
    │  depends on
    ▼
AppService (pkg/siteadmin/service/app.go)
    │  depends on
    ├── AppServiceConfigSourceStore  → *configsource.Store           (app metadata — 3 new methods added)
    ├── AppServiceOwnerStore         → *AppOwnerStore                 (SQL: owner queries)
    ├── AppServiceAdminAPI           → *portalservice.AdminAPIService (resolve email / search by email)
    ├── AuditDatabase              → *auditdb.ReadHandle            (nil if audit DB not configured)
    ├── AuditStore                 → *analytic.AuditDBReadStore     (cumulative user count)
    ├── AppServiceHTTPClient                                          (HTTP for GraphQL)
    └── clock.Clock
```

### `ListApps` branching logic

```
┌─ owner_email filter only ──────────────────────────────────────────────┐
│  1. Admin API: getUsersByStandardAttribute("email", v) → []userID      │
│  2. AppOwnerStore: GetAppIDsByOwnerUserID(userID)     → []appID        │
│  3. ConfigSourceStore: GetManyByAppIDs(appIDs)        → []source       │
│  4. Build page (in-memory; bounded by one user's app count)            │
│  5. owner_email already known — no extra GraphQL needed                │
└────────────────────────────────────────────────────────────────────────┘

┌─ app_id filter (with or without owner_email) ──────────────────────────┐
│  1. ConfigSourceStore: GetDatabaseSourceByAppID(appID)                 │
│  2. AppOwnerStore: GetOwnerByAppID(appID)         → ownerUserID        │
│  3. Admin API: resolveUserEmails([ownerUserID])   → email              │
│  4. If owner_email filter also set: check email match                  │
└────────────────────────────────────────────────────────────────────────┘

┌─ no filter (default) ──────────────────────────────────────────────────┐
│  1. ConfigSourceStore: CountAll()                 → totalCount         │
│  2. ConfigSourceStore: ListPaged(limit, offset)   → []source (page)   │
│  3. AppOwnerStore: GetOwnersByAppIDs(appIDs)      → map[appID]userID  │
│  4. Admin API: resolveUserEmails(userIDs)         → map[userID]email  │
│     (only page_size emails resolved, e.g., 20)                        │
└────────────────────────────────────────────────────────────────────────┘
```

---

## Key Dependencies

| What | Where |
|---|---|
| `configsource.Store` | `pkg/lib/config/configsource/store.go` |
| `portalservice.AdminAPIService` | `pkg/portal/service/admin_api.go` |
| `portalservice.DefaultDomainService` | `pkg/portal/service/default_domain.go` |
| `authz.Adder` | `pkg/lib/admin/authz/adder.go` |
| `analytic.AuditDBReadStore` | `pkg/lib/analytic/auditdb_read_store.go` |
| `analytic.CumulativeUserCountType` | `pkg/lib/analytic/count.go` — `"cumulative.user"` |
| `analytic.ErrAnalyticCountNotFound` | `pkg/lib/analytic/auditdb_read_store.go` |
| `auditdb.ReadHandle` | `pkg/lib/infra/db/auditdb/` — nil when audit DB not configured |
| `timeutil.TruncateToDate` | `pkg/util/timeutil/` |
| `graphqlutil.DoParams` / `graphqlutil.HTTPDo` | `pkg/util/graphqlutil/http_do.go` |
| `relay.ToGlobalID` / `relay.FromGlobalID` | `pkg/graphqlgo/relay` |
| `portalservice.UsageInternal` | `pkg/portal/service/admin_api.go` |
| `getUsersByStandardAttribute` query | `pkg/portal/service/collaborator.go:824` (existing pattern) |

---

## Files to Create

### 1. `pkg/siteadmin/service/app.go`

```go
package service

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "net/http"
    "strings"

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

type AppServiceConfigSourceStore interface {
    GetDatabaseSourceByAppID(ctx context.Context, appID string) (*configsource.DatabaseSource, error)
    CountAll(ctx context.Context) (int, error)
    ListPaged(ctx context.Context, limit int, offset int) ([]*configsource.DatabaseSource, error)
    GetManyByAppIDs(ctx context.Context, appIDs []string) ([]*configsource.DatabaseSource, error)
}

type AppServiceOwnerStore interface {
    GetOwnerByAppID(ctx context.Context, appID string) (string, error)                              // returns userID; ErrOwnerNotFound if none
    GetOwnersByAppIDs(ctx context.Context, appIDs []string) (map[string]string, error)              // map[appID]userID
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
// owner relationships. It lives here to avoid touching pkg/lib/.
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
    userIDs, err := s.findUserIDsByEmail(ctx, params.OwnerEmail)
    if err != nil {
        return nil, err
    }

    if len(userIDs) == 0 {
        // User not found by email — return empty result.
        return &ListAppsResult{Apps: []siteadmin.App{}, TotalCount: 0}, nil
    }

    // Email is unique — take the first (and expected only) match.
    userID := userIDs[0]

    totalCount, err := s.OwnerStore.CountAppsByOwnerUserID(ctx, userID)
    if err != nil {
        return nil, err
    }

    if totalCount == 0 {
        return &ListAppsResult{Apps: []siteadmin.App{}, TotalCount: 0}, nil
    }

    offset := (params.Page - 1) * params.PageSize
    appIDs, err := s.OwnerStore.ListAppIDsByOwnerUserIDPaged(ctx, userID, params.PageSize, offset)
    if err != nil {
        return nil, err
    }

    sources, err := s.ConfigSourceStore.GetManyByAppIDs(ctx, appIDs)
    if err != nil {
        return nil, err
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
    src, err := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, params.AppID)
    if err != nil {
        if errors.Is(err, configsource.ErrAppNotFound) {
            return &ListAppsResult{Apps: []siteadmin.App{}, TotalCount: 0}, nil
        }
        return nil, err
    }

    ownerUserID, err := s.OwnerStore.GetOwnerByAppID(ctx, params.AppID)
    if err != nil && !errors.Is(err, ErrOwnerNotFound) {
        return nil, err
    }

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
    totalCount, err := s.ConfigSourceStore.CountAll(ctx)
    if err != nil {
        return nil, err
    }

    offset := (params.Page - 1) * params.PageSize
    sources, err := s.ConfigSourceStore.ListPaged(ctx, params.PageSize, offset)
    if err != nil {
        return nil, err
    }

    appIDs := make([]string, len(sources))
    for i, src := range sources {
        appIDs[i] = src.AppID
    }

    ownerMap, err := s.OwnerStore.GetOwnersByAppIDs(ctx, appIDs)
    if err != nil {
        return nil, err
    }

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
    src, err := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
    if err != nil {
        return nil, err
    }

    ownerUserID, err := s.OwnerStore.GetOwnerByAppID(ctx, appID)
    if err != nil && !errors.Is(err, ErrOwnerNotFound) {
        return nil, err
    }

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
        id := relay.FromGlobalID(globalID).ID
        if id == "" {
            // relay.FromGlobalID failed to parse — skip this entry.
            continue
        }
        ids = append(ids, id)
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
```

> **Import note**: `sq "github.com/Masterminds/squirrel"` is needed for `sq.Eq` in
> `GetOwnersByAppIDs`. Check the import alias used in other store files in `pkg/lib/`.

---

### 2. `pkg/siteadmin/service/deps.go`

```go
package service

import (
    "net/http"
    "time"

    "github.com/google/wire"

    "github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewHTTPClient() AppServiceHTTPClient {
    return AppServiceHTTPClient{
        Client: httputil.NewExternalClient(5 * time.Second),
    }
}

var DependencySet = wire.NewSet(
    wire.Struct(new(AppOwnerStore), "*"),
    wire.Bind(new(AppServiceOwnerStore), new(*AppOwnerStore)),
    wire.Struct(new(AppService), "*"),
    NewHTTPClient,
)
```

---

## Files to Modify

### `pkg/lib/config/configsource/store.go`

Add three new methods. They follow the exact same SQL builder pattern as the existing
`ListAll` and `GetDatabaseSourceByAppID` methods in the same file.

```go
func (s *Store) CountAll(ctx context.Context) (int, error) {
    q := s.SQLBuilder.
        Select("COUNT(*)").
        From(s.SQLBuilder.TableName("_portal_config_source"))

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

func (s *Store) ListPaged(ctx context.Context, limit int, offset int) ([]*DatabaseSource, error) {
    builder := s.selectConfigSourceQuery().
        OrderBy("created_at DESC").
        Limit(uint64(limit)).
        Offset(uint64(offset))

    rows, err := s.SQLExecutor.QueryWith(ctx, builder)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []*DatabaseSource
    for rows.Next() {
        item, err := s.scanConfigSource(rows)
        if err != nil {
            return nil, err
        }
        items = append(items, item)
    }
    return items, nil
}

func (s *Store) GetManyByAppIDs(ctx context.Context, appIDs []string) ([]*DatabaseSource, error) {
    if len(appIDs) == 0 {
        return nil, nil
    }

    builder := s.selectConfigSourceQuery().
        Where(sq.Eq{"app_id": appIDs}).
        OrderBy("created_at DESC")

    rows, err := s.SQLExecutor.QueryWith(ctx, builder)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []*DatabaseSource
    for rows.Next() {
        item, err := s.scanConfigSource(rows)
        if err != nil {
            return nil, err
        }
        items = append(items, item)
    }
    return items, nil
}
```

> **Import note**: `sq "github.com/Masterminds/squirrel"` is needed for `sq.Eq`. Check the
> existing import alias in `store.go` — the file may already import squirrel under a different
> alias. Align with whatever is already used.

---

### `pkg/siteadmin/deps.go` (additions across commits 3–5)

The lines below are added incrementally — do **not** touch existing items from plans 01/02.

**Commit 3** adds:

```go
    // siteadmin service package
    siteadminservice.DependencySet,

    // AdminAPIService — needed by AppService for email resolution via GraphQL
    wire.Struct(new(authz.Adder), "Clock"),
    wire.Bind(new(portalservice.AuthzAdder), new(*authz.Adder)),
    wire.Struct(new(portalservice.DefaultDomainService), "AppHostSuffixes", "AppConfig"),
    wire.Bind(new(portalservice.AdminAPIDefaultDomainService), new(*portalservice.DefaultDomainService)),
    wire.Struct(new(portalservice.AdminAPIService), "AuthgearConfig", "AdminAPIConfig", "ConfigSource", "AuthzAdder", "DefaultDomains"),
    wire.Bind(new(siteadminservice.AppServiceAdminAPI), new(*portalservice.AdminAPIService)),

    // configsource.Store — app metadata (AppID, PlanName, CreatedAt)
    wire.Struct(new(configsource.Store), "*"),
    wire.Bind(new(siteadminservice.AppServiceConfigSourceStore), new(*configsource.Store)),

    // Audit DB — cumulative user count for AppDetail
    auditdb.DependencySet,
    wire.Struct(new(analytic.AuditDBReadStore), "*"),
```

**Commit 4** adds:

```go
    wire.Bind(new(transport.AppsListService), new(*siteadminservice.AppService)),
```

**Commit 5** adds:

```go
    wire.Bind(new(transport.AppGetService), new(*siteadminservice.AppService)),
```

> **New import aliases** (add in commit 3):
> - `siteadminservice "github.com/authgear/authgear-server/pkg/siteadmin/service"`
> - `portalservice "github.com/authgear/authgear-server/pkg/portal/service"`
> - `configsource "github.com/authgear/authgear-server/pkg/lib/config/configsource"`
> - `authz "github.com/authgear/authgear-server/pkg/lib/admin/authz"`
> - `"github.com/authgear/authgear-server/pkg/lib/analytic"`
> - `"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"`

---

### `pkg/siteadmin/transport/handler_apps_list.go`

Add the service interface and field; replace `ServeHTTP`:

```go
type AppsListService interface {
    ListApps(ctx context.Context, params service.ListAppsParams) (*service.ListAppsResult, error)
}

type AppsListHandler struct {
    AppsList AppsListService
}
```

Updated `ServeHTTP`:

```go
func (h *AppsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    params := parseAppsListParams(r)

    result, err := h.AppsList.ListApps(r.Context(), service.ListAppsParams{
        Page:       params.Page,
        PageSize:   params.PageSize,
        AppID:      params.AppID,
        OwnerEmail: params.OwnerEmail,
    })
    if err != nil {
        writeError(w, r, err)
        return
    }

    response := siteadmin.AppsListResponse{
        Apps:       result.Apps,
        TotalCount: result.TotalCount,
        Page:       params.Page,
        PageSize:   params.PageSize,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(response)
}
```

Remove the `dummyApps` variable and unused imports (`"time"`, `"strings"`).
Add: `service "github.com/authgear/authgear-server/pkg/siteadmin/service"`.

---

### `pkg/siteadmin/transport/handler_app_get.go`

Add the service interface and field; replace `ServeHTTP`:

```go
type AppGetService interface {
    GetApp(ctx context.Context, appID string) (*siteadmin.AppDetail, error)
}

type AppGetHandler struct {
    AppGet AppGetService
}
```

Updated `ServeHTTP`:

```go
func (h *AppGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    appID := httproute.GetParam(r, "app_id")

    detail, err := h.AppGet.GetApp(r.Context(), appID)
    if err != nil {
        writeError(w, r, err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(detail)
}
```

---

### `pkg/siteadmin/transport/deps.go`

No changes needed — `wire.Struct(new(AppsListHandler), "*")` and
`wire.Struct(new(AppGetHandler), "*")` already use `"*"`, so wire will pick up the new
`AppsList` and `AppGet` fields automatically once they are bound in `pkg/siteadmin/deps.go`.

---

### `pkg/siteadmin/wire_gen.go`

Run after all changes:

```bash
wire gen ./pkg/siteadmin/...
go build ./pkg/siteadmin/...
```

Do **not** hand-edit `wire_gen.go`.

---

## Implementation Roadmap: 5 Atomic Commits

Each component is wired into DI in the same commit it is added. Every commit leaves the
codebase in a working, fully-wired state.

### **Commit 1: Add CountAll, ListPaged, GetManyByAppIDs to configsource.Store**

**Files modified:**
- `pkg/lib/config/configsource/store.go`

Pure data-layer additions, no callers yet.

**Commit message:** `Add CountAll, ListPaged, GetManyByAppIDs to configsource.Store`

---

### **Commit 2: Add AppOwnerStore and AppService**

**Files created:**
- `pkg/siteadmin/service/app.go`
- `pkg/siteadmin/service/app_test.go`

All business logic, narrow interfaces, and unit tests. No wire setup yet — this package has
no `init`/side-effects so it compiles and is safe to land without being wired in.

```bash
go test ./pkg/siteadmin/service/...
```

**Commit message:** `Add AppOwnerStore and AppService`

---

### **Commit 3: Wire AppService into siteadmin DI**

**Files created:**
- `pkg/siteadmin/service/deps.go`

**Files modified:**
- `pkg/siteadmin/deps.go` — add `siteadminservice.DependencySet`, `AdminAPIService`,
  `configsource.Store`, `auditdb.DependencySet`, and `analytic.AuditDBReadStore` bindings
  (delta only; see section above)
- `pkg/siteadmin/wire_gen.go` — regenerated

`AppService` is now available in the DI graph. No handler uses it yet.

```bash
wire gen ./pkg/siteadmin/...
go build ./pkg/siteadmin/...
```

**Commit message:** `Wire AppService into siteadmin DI`

---

### **Commit 4: Replace dummy data in AppsListHandler**

**Files modified:**
- `pkg/siteadmin/transport/handler_apps_list.go` — add `AppsListService` interface,
  `AppsList` field, updated `ServeHTTP`; remove `dummyApps` and unused imports
- `pkg/siteadmin/deps.go` — add `wire.Bind(new(transport.AppsListService), new(*siteadminservice.AppService))`
- `pkg/siteadmin/wire_gen.go` — regenerated

Handler is fully wired after this commit.

```bash
wire gen ./pkg/siteadmin/...
go build ./pkg/siteadmin/...
```

**Commit message:** `Replace dummy data in AppsListHandler`

---

### **Commit 5: Replace dummy data in AppGetHandler**

**Files modified:**
- `pkg/siteadmin/transport/handler_app_get.go` — add `AppGetService` interface,
  `AppGet` field, updated `ServeHTTP`; remove dummy variables and unused imports
- `pkg/siteadmin/deps.go` — add `wire.Bind(new(transport.AppGetService), new(*siteadminservice.AppService))`
- `pkg/siteadmin/wire_gen.go` — regenerated

Handler is fully wired after this commit.

```bash
wire gen ./pkg/siteadmin/...
go build ./pkg/siteadmin/...
```

**Commit message:** `Replace dummy data in AppGetHandler`

---

## Tests

### `pkg/siteadmin/service/app_test.go`

Added in **Commit 2** alongside `app.go`. Uses hand-written fakes and GoConvey, matching
the pattern in `pkg/siteadmin/transport/middleware_authz_test.go`.

#### Fakes

```go
package service

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "net/url"
    "testing"
    "time"

    . "github.com/smartystreets/goconvey/convey"

    "github.com/authgear/authgear-server/pkg/api/model"
    "github.com/authgear/authgear-server/pkg/lib/config/configsource"
    relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
    portalservice "github.com/authgear/authgear-server/pkg/portal/service"
    "github.com/authgear/authgear-server/pkg/portal/session"
    "github.com/authgear/authgear-server/pkg/util/clock"
)

// fakeAdminAPI rewrites every request to the given test server URL.
type fakeAdminAPI struct{ serverURL string }

func (f *fakeAdminAPI) SelfDirector(ctx context.Context, actorUserID string, usage portalservice.Usage) (func(*http.Request), error) {
    target, _ := url.Parse(f.serverURL)
    return func(r *http.Request) {
        r.URL.Scheme = target.Scheme
        r.URL.Host = target.Host
    }, nil
}

// adminAPIServer returns a test server that serves a fixed JSON response for every request.
func adminAPIServer(response interface{}) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(response)
    }))
}

// fakeOwnerStore implements AppServiceOwnerStore.
// capturedLimit records the limit argument passed to ListAppIDsByOwnerUserIDPaged.
type fakeOwnerStore struct {
    countResult   int
    appIDsResult  []string
    ownerResult   string
    ownersResult  map[string]string
    capturedLimit int
}

func (f *fakeOwnerStore) CountAppsByOwnerUserID(_ context.Context, _ string) (int, error) {
    return f.countResult, nil
}
func (f *fakeOwnerStore) ListAppIDsByOwnerUserIDPaged(_ context.Context, _ string, limit int, _ int) ([]string, error) {
    f.capturedLimit = limit
    return f.appIDsResult, nil
}
func (f *fakeOwnerStore) GetOwnerByAppID(_ context.Context, _ string) (string, error) {
    if f.ownerResult == "" {
        return "", ErrOwnerNotFound
    }
    return f.ownerResult, nil
}
func (f *fakeOwnerStore) GetOwnersByAppIDs(_ context.Context, _ []string) (map[string]string, error) {
    return f.ownersResult, nil
}

// fakeConfigSourceStore implements AppServiceConfigSourceStore.
type fakeConfigSourceStore struct {
    byAppID map[string]*configsource.DatabaseSource
    all     []*configsource.DatabaseSource
    total   int
}

func (f *fakeConfigSourceStore) GetDatabaseSourceByAppID(_ context.Context, appID string) (*configsource.DatabaseSource, error) {
    src, ok := f.byAppID[appID]
    if !ok {
        return nil, configsource.ErrAppNotFound
    }
    return src, nil
}
func (f *fakeConfigSourceStore) CountAll(_ context.Context) (int, error) { return f.total, nil }
func (f *fakeConfigSourceStore) ListPaged(_ context.Context, _, _ int) ([]*configsource.DatabaseSource, error) {
    return f.all, nil
}
func (f *fakeConfigSourceStore) GetManyByAppIDs(_ context.Context, appIDs []string) ([]*configsource.DatabaseSource, error) {
    result := make([]*configsource.DatabaseSource, 0, len(appIDs))
    for _, id := range appIDs {
        if src, ok := f.byAppID[id]; ok {
            result = append(result, src)
        }
    }
    return result, nil
}

// getUsersByEmailResponse builds the JSON payload for getUsersByStandardAttribute.
func getUsersByEmailResponse(globalIDs ...string) map[string]interface{} {
    users := make([]interface{}, len(globalIDs))
    for i, id := range globalIDs {
        users[i] = map[string]interface{}{"id": id}
    }
    return map[string]interface{}{"data": map[string]interface{}{"users": users}}
}
```

#### Test cases

```go
func TestAppService(t *testing.T) {
    ownerEmail := "owner@example.com"
    fixedTime := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
    sessionCtx := session.WithSessionInfo(context.Background(), &model.SessionInfo{
        IsValid: true, UserID: "actor-user",
    })

    Convey("ListApps — owner_email filter path", t, func() {

        Convey("returns empty when no user has the given email", func() {
            srv := adminAPIServer(getUsersByEmailResponse( /* no IDs */ ))
            defer srv.Close()

            s := &AppService{
                ConfigSourceStore: &fakeConfigSourceStore{},
                OwnerStore:        &fakeOwnerStore{},
                AdminAPI:          &fakeAdminAPI{serverURL: srv.URL},
                HTTPClient:        AppServiceHTTPClient{Client: &http.Client{}},
                Clock:             clock.NewMockClockAtTime(fixedTime),
            }
            result, err := s.ListApps(sessionCtx, ListAppsParams{Page: 1, PageSize: 20, OwnerEmail: ownerEmail})
            So(err, ShouldBeNil)
            So(result.TotalCount, ShouldEqual, 0)
            So(result.Apps, ShouldBeEmpty)
        })

        Convey("returns empty when Admin API returns a user with an unparseable global ID", func() {
            // base64("no-colon-here") has no ":" separator so relay.FromGlobalID returns ID=""
            malformed := base64.StdEncoding.EncodeToString([]byte("no-colon-here"))
            srv := adminAPIServer(getUsersByEmailResponse(malformed))
            defer srv.Close()

            s := &AppService{
                ConfigSourceStore: &fakeConfigSourceStore{},
                OwnerStore:        &fakeOwnerStore{},
                AdminAPI:          &fakeAdminAPI{serverURL: srv.URL},
                HTTPClient:        AppServiceHTTPClient{Client: &http.Client{}},
                Clock:             clock.NewMockClockAtTime(fixedTime),
            }
            result, err := s.ListApps(sessionCtx, ListAppsParams{Page: 1, PageSize: 20, OwnerEmail: ownerEmail})
            So(err, ShouldBeNil)
            So(result.TotalCount, ShouldEqual, 0)
            So(result.Apps, ShouldBeEmpty)
        })

        Convey("returns empty when user owns no apps", func() {
            srv := adminAPIServer(getUsersByEmailResponse(relay.ToGlobalID("User", "user-1")))
            defer srv.Close()

            s := &AppService{
                ConfigSourceStore: &fakeConfigSourceStore{},
                OwnerStore:        &fakeOwnerStore{countResult: 0},
                AdminAPI:          &fakeAdminAPI{serverURL: srv.URL},
                HTTPClient:        AppServiceHTTPClient{Client: &http.Client{}},
                Clock:             clock.NewMockClockAtTime(fixedTime),
            }
            result, err := s.ListApps(sessionCtx, ListAppsParams{Page: 1, PageSize: 20, OwnerEmail: ownerEmail})
            So(err, ShouldBeNil)
            So(result.TotalCount, ShouldEqual, 0)
            So(result.Apps, ShouldBeEmpty)
        })

        Convey("returns paginated apps with correct TotalCount and OwnerEmail", func() {
            src1 := &configsource.DatabaseSource{AppID: "app-1", PlanName: "free", CreatedAt: fixedTime}
            src2 := &configsource.DatabaseSource{AppID: "app-2", PlanName: "startups", CreatedAt: fixedTime}

            srv := adminAPIServer(getUsersByEmailResponse(relay.ToGlobalID("User", "user-1")))
            defer srv.Close()

            s := &AppService{
                ConfigSourceStore: &fakeConfigSourceStore{
                    byAppID: map[string]*configsource.DatabaseSource{"app-1": src1, "app-2": src2},
                },
                OwnerStore: &fakeOwnerStore{
                    countResult:  2,
                    appIDsResult: []string{"app-1", "app-2"},
                },
                AdminAPI:   &fakeAdminAPI{serverURL: srv.URL},
                HTTPClient: AppServiceHTTPClient{Client: &http.Client{}},
                Clock:      clock.NewMockClockAtTime(fixedTime),
            }
            result, err := s.ListApps(sessionCtx, ListAppsParams{Page: 1, PageSize: 20, OwnerEmail: ownerEmail})
            So(err, ShouldBeNil)
            So(result.TotalCount, ShouldEqual, 2)
            So(result.Apps, ShouldHaveLength, 2)
            So(result.Apps[0].Id, ShouldEqual, "app-1")
            So(result.Apps[0].OwnerEmail, ShouldEqual, ownerEmail)
            So(result.Apps[1].Id, ShouldEqual, "app-2")
        })

        Convey("clamps PageSize exceeding maxPageSize", func() {
            ownerStore := &fakeOwnerStore{
                countResult:  1,
                appIDsResult: []string{"app-1"},
            }
            srv := adminAPIServer(getUsersByEmailResponse(relay.ToGlobalID("User", "user-1")))
            defer srv.Close()

            s := &AppService{
                ConfigSourceStore: &fakeConfigSourceStore{
                    byAppID: map[string]*configsource.DatabaseSource{"app-1": {AppID: "app-1"}},
                },
                OwnerStore: ownerStore,
                AdminAPI:   &fakeAdminAPI{serverURL: srv.URL},
                HTTPClient: AppServiceHTTPClient{Client: &http.Client{}},
                Clock:      clock.NewMockClockAtTime(fixedTime),
            }
            _, err := s.ListApps(sessionCtx, ListAppsParams{Page: 1, PageSize: 100, OwnerEmail: ownerEmail})
            So(err, ShouldBeNil)
            So(ownerStore.capturedLimit, ShouldEqual, maxPageSize)
        })
    })

    Convey("GetApp — audit DB nil", t, func() {
        Convey("returns UserCount 0 when AuditDatabase is nil", func() {
            src := &configsource.DatabaseSource{AppID: "app-1", PlanName: "free", CreatedAt: fixedTime}
            s := &AppService{
                ConfigSourceStore: &fakeConfigSourceStore{
                    byAppID: map[string]*configsource.DatabaseSource{"app-1": src},
                },
                OwnerStore: &fakeOwnerStore{ownerResult: "user-1"},
                AdminAPI: &fakeAdminAPI{serverURL: adminAPIServer(map[string]interface{}{
                    "data": map[string]interface{}{
                        "nodes": []interface{}{
                            map[string]interface{}{
                                "id":                 relay.ToGlobalID("User", "user-1"),
                                "standardAttributes": map[string]interface{}{"email": "owner@example.com"},
                            },
                        },
                    },
                }).URL},
                AuditDatabase: nil, // not configured
                HTTPClient:    AppServiceHTTPClient{Client: &http.Client{}},
                Clock:         clock.NewMockClockAtTime(fixedTime),
            }
            detail, err := s.GetApp(sessionCtx, "app-1")
            So(err, ShouldBeNil)
            So(detail.UserCount, ShouldEqual, 0)
        })
    })
}
```

> **Note on the `GetApp` test server**: the inline `adminAPIServer(...)` creates a server whose
> `Close()` is never called in this snippet. In the real test, assign it to a variable and call
> `defer srv.Close()`. Shown inline here for brevity.

---

## Dependency Graph

```
Commit 1  configsource.Store new methods   (pkg/lib — no callers yet)
    ↓
Commit 2  AppService logic            (pkg/siteadmin/service — not wired yet)
    ↓
Commit 3  Wire AppService into DI             (service/deps.go + root deps.go + wire_gen.go)
    ↓
Commit 4  AppsListHandler + wire binding   (transport handler + root deps.go + wire_gen.go)
    ↓
Commit 5  AppGetHandler + wire binding     (transport handler + root deps.go + wire_gen.go)
```

---

## Verification

### List all apps (paginated at DB level)

```bash
curl -s -H "Authorization: Bearer <token>" \
  "http://localhost:3005/api/v1/apps?page=1&page_size=20" | jq .
```

Expected: first 20 apps from the database with real owner emails resolved from Admin API.

### Filter by `owner_email` (Admin API search path)

```bash
curl -s -H "Authorization: Bearer <token>" \
  "http://localhost:3005/api/v1/apps?owner_email=alice@example.com" | jq .
```

Expected: only apps owned by `alice@example.com`. No in-memory email scan — email lookup
goes through `getUsersByStandardAttribute` first.

### Filter by `app_id`

```bash
curl -s -H "Authorization: Bearer <token>" \
  "http://localhost:3005/api/v1/apps?app_id=my-app" | jq .
```

### Get single app detail (with cumulative user count)

```bash
curl -s -H "Authorization: Bearer <token>" \
  "http://localhost:3005/api/v1/apps/my-app-id" | jq .
```

Expected: `AppDetail` with `user_count` from `_audit_analytic_count` for yesterday
(`CumulativeUserCountType`). Returns 0 if the audit DB is not configured or has no record
for yesterday.

### Unknown app → 404

```bash
curl -i -H "Authorization: Bearer <token>" \
  "http://localhost:3005/api/v1/apps/does-not-exist"
```

Expected: `404 Not Found` (propagated from `configsource.ErrAppNotFound`).
