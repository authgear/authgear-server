# 08 — Site Admin Audit Log

## Goal / Scope

Emit a non-blocking audit log event for each of the four site admin mutations:

| Endpoint | Event type |
|---|---|
| `POST /api/v1/apps/{app_id}/plan` | `site_admin.app.plan.updated` |
| `POST /api/v1/apps/{app_id}/collaborators` | `site_admin.app.collaborator.added` |
| `DELETE /api/v1/apps/{app_id}/collaborators/{id}` | `site_admin.app.collaborator.deleted` |
| `POST /api/v1/apps/{app_id}/collaborators/{id}/promote` | `site_admin.app.collaborator.promoted` |

Events are written to the **portal app's** (`SITEADMIN_AUTHGEAR_APP_ID`) audit log in the global audit DB (configured via `AUDIT_DATABASE_URL`). The affected app ID is recorded in `data.payload.app_id`. This mirrors the pattern used by other portal-level audit events.

No schema migration is required. The audit log table (`_audit_log`) already exists.

---

## Design

### Event type naming

Types follow the `project.*` naming convention — `namespace.resource.verb` with simple past-tense verbs — rather than the verbose `admin_api.mutation.*.executed` pattern:

```
site_admin.app.plan.updated
site_admin.app.collaborator.added
site_admin.app.collaborator.deleted
site_admin.app.collaborator.promoted
```

Future site admin types should continue this pattern (e.g. `site_admin.app.deleted`, `site_admin.app.feature.updated`).

### Dispatch location

Event dispatch happens at the **end of each service method**, after the main `GlobalDatabase.WithTx` call succeeds. This mirrors the admin API pattern (dispatch after commit) while staying in the service layer where all payload data is already assembled.

### Audit writer

The site admin service layer already has a `*auditdb.ReadHandle` for usage queries (`AppService`). We extend the DI graph with a `*auditdb.WriteHandle` and introduce a new `SiteAdminAuditService` that writes one log entry per call.

All records are stored under the portal app ID (`AuthgearConfig.AppID`) so they land in the same `_audit_log` partition that portal events use. The affected app ID is already present in each payload struct's `app_id` JSON field and surfaces via `data->'payload'->>'app_id'` in queries.

### Log structure alignment with portal audit logs

Site admin audit log records are structured to match portal audit logs exactly:

| Field | Portal | Site Admin |
|---|---|---|
| `_audit_log.user_id` | `''` (empty) | `''` (empty) |
| `_audit_log.user_agent` | browser UA | browser UA |
| `context.user_id` | `null` | `null` |
| `context.user_agent` | browser UA | browser UA |
| `context.audit_context` | `{usage, actor_user_id, http_url, http_referer}` | same |
| `context.preferred_languages` | `[]` | `[]` |
| `id` / `seq` | from `_auth_event_sequence` | from `_auth_event_sequence` |

Key implementation notes:
- `_audit_log.ip_address` is type `inet` — inserting `""` causes a type error. Inject `httputil.RemoteIP`.
- The actor's user ID goes in `audit_context.actor_user_id`, **not** `context.user_id`.
- The event `id` is `fmt.Sprintf("%016x", seq)` — set by `libevent.NewNonBlockingEvent` from the sequence. Do **not** override it with `uuid.New()`.
- Obtain `seq` from `_auth_event_sequence` via the global DB before writing.

### Admin API protection

The admin API `auditLogs` query resolves `activityType` against a non-null GraphQL enum. If `site_admin.*` records reach the resolver, it panics on serialization. Fix: when no `activityTypes` filter is provided, default to `knownAuditLogActivityTypes` (all values defined in the `AuditLogActivityType` enum). Site admin types are not in the enum and are therefore automatically excluded.

### New `TriggeredBy` constant

`TriggeredBySiteAdmin = "site_admin"` is added to `pkg/api/event/context.go`. It is distinct from `TriggeredByPortal` and `TriggeredByTypeAdminAPI`, allowing callers and dashboards to filter by source.

---

## Step 1 — New `TriggeredBy` constant

**File:** `pkg/api/event/context.go`

```go
// TriggeredBySiteAdmin means the event originates from the Site Admin API.
TriggeredBySiteAdmin TriggeredByType = "site_admin"
```

---

## Step 2 — New event payload types

Four new files under `pkg/api/event/nonblocking/`. All four share:
- `ForHook()` → `false`
- `ForAudit()` → `true`
- `GetTriggeredBy()` → `event.TriggeredBySiteAdmin`
- `UserID()` → `""` (actor is in `audit_context.actor_user_id`)

### `siteadmin_app_plan_updated.go`

```go
const SiteAdminAppPlanUpdated event.Type = "site_admin.app.plan.updated"

type SiteAdminAppPlanUpdatedEventPayload struct {
    AppID   string `json:"app_id"`
    OldPlan string `json:"old_plan"`
    NewPlan string `json:"new_plan"`
}
```

`OldPlan` is captured from the config source before the plan is overwritten (see Step 4).

### `siteadmin_app_collaborator_added.go`

```go
const SiteAdminAppCollaboratorAdded event.Type = "site_admin.app.collaborator.added"

type SiteAdminAppCollaboratorAddedEventPayload struct {
    AppID              string `json:"app_id"`
    CollaboratorID     string `json:"collaborator_id"`
    CollaboratorUserID string `json:"user_id"`
    UserEmail          string `json:"user_email"`
    Role               string `json:"role"`
}
```

### `siteadmin_app_collaborator_deleted.go`

```go
const SiteAdminAppCollaboratorDeleted event.Type = "site_admin.app.collaborator.deleted"

type SiteAdminAppCollaboratorDeletedEventPayload struct {
    AppID                 string `json:"app_id"`
    CollaboratorID        string `json:"collaborator_id"`
    CollaboratorUserID    string `json:"user_id"`
    CollaboratorUserEmail string `json:"user_email"`
}
```

### `siteadmin_app_collaborator_promoted.go`

```go
const SiteAdminAppCollaboratorPromoted event.Type = "site_admin.app.collaborator.promoted"

type SiteAdminAppCollaboratorPromotedEventPayload struct {
    AppID                  string `json:"app_id"`
    NewOwnerCollaboratorID string `json:"new_owner_collaborator_id"`
    NewOwnerUserID         string `json:"new_owner_user_id"`
    NewOwnerUserEmail      string `json:"new_owner_user_email"`
    // DemotedEditor* fields are omitted when the app had no previous owner.
    DemotedEditorCollaboratorID string `json:"demoted_editor_collaborator_id,omitempty"`
    DemotedEditorUserID         string `json:"demoted_editor_user_id,omitempty"`
    DemotedEditorUserEmail      string `json:"demoted_editor_user_email,omitempty"`
}
```

---

## Step 3 — `SiteAdminAuditService`

**New file:** `pkg/siteadmin/service/audit.go`

```go
package service

import (
    "context"
    "fmt"
    "net/http"

    libevent "github.com/authgear/authgear-server/pkg/lib/event"
    "github.com/authgear/authgear-server/pkg/api/event"
    "github.com/authgear/authgear-server/pkg/lib/audit"
    "github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
    "github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
    portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
    portalservice "github.com/authgear/authgear-server/pkg/portal/service"
    "github.com/authgear/authgear-server/pkg/portal/session"
    "github.com/authgear/authgear-server/pkg/util/clock"
    "github.com/authgear/authgear-server/pkg/util/httputil"
)

// SiteAdminAuditService writes audit log entries for site admin mutations to
// the global audit DB. All records are stored under the portal app ID
// (SITEADMIN_AUTHGEAR_APP_ID); the affected app ID is in the payload.
//
// The log structure mirrors portal audit logs: the actor's user ID is in
// audit_context.actor_user_id (not in context.user_id), user_agent and
// http_url / http_referer are populated from the HTTP request, and seq
// is obtained from _auth_event_sequence (same as portal events).
type SiteAdminAuditService struct {
    AuditDatabase     *auditdb.WriteHandle    // nil when audit DB is not configured
    SQLBuilder        *auditdb.SQLBuilder     // global (not app-scoped)
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

// nextSeq returns the next value from _auth_event_sequence, mirroring
// portal/service.AuditService.
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

    // Obtain a real sequence number, mirroring portal audit logs.
    var seq int64
    err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
        var e error
        seq, e = s.nextSeq(ctx)
        return e
    })
    if err != nil {
        return err
    }

    // Build audit_context mirroring portal/service.AuditService.makeContext:
    // actor goes in audit_context.actor_user_id, NOT in context.user_id.
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

    // NewNonBlockingEvent sets ID = fmt.Sprintf("%016x", seq) — do not override.
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
```

**Add to `pkg/siteadmin/service/deps.go`:**

```go
wire.Struct(new(SiteAdminAuditService), "*"),
```

---

## Step 4 — Wire in `auditdb.WriteHandle` and additional dependencies

**`pkg/siteadmin/deps.go`** — extend the existing `auditdb` block and add interface bindings:

```go
// Audit DB (optional — nil when not configured)
auditdb.DependencySet,
auditdb.NewReadHandle,
auditdb.NewWriteHandle,          // ← add
wire.Struct(new(analytic.AuditDBReadStore), "*"),

// SiteAdminAuditService interface bindings
wire.Bind(new(siteadminservice.PlanServiceAuditService), new(*siteadminservice.SiteAdminAuditService)),
wire.Bind(new(siteadminservice.CollaboratorServiceAuditService), new(*siteadminservice.SiteAdminAuditService)),
```

`auditdb.DependencySet` already wires `NewWriteSQLExecutor`. `RemoteIP`, `UserAgentString`, `HTTPRequestURL`, `Request`, `GlobalDatabase`, `GlobalSQLBuilder`, and `GlobalSQLExecutor` are already in the graph via `deps.DependencySet`.

After all deps changes: `make generate`.

---

## Step 5 — Inject `SiteAdminAuditService` into mutation services

Both services use a narrow interface so tests can inject a fake without
needing the full audit service:

```go
// in plan.go
type PlanServiceAuditService interface {
    LogEvent(ctx context.Context, appID string, payload event.NonBlockingPayload) error
}

// in collaborator.go
type CollaboratorServiceAuditService interface {
    LogEvent(ctx context.Context, appID string, payload event.NonBlockingPayload) error
}
```

### `PlanService`

Add field `AuditService PlanServiceAuditService`. In `ChangeAppPlan`, capture `oldPlanName` before overwriting, then after `GlobalDatabase.WithTx` succeeds:

```go
if s.AuditService != nil {
    _ = s.AuditService.LogEvent(ctx, appID, &nonblocking.SiteAdminAppPlanUpdatedEventPayload{
        AppID:   appID,
        OldPlan: oldPlanName,
        NewPlan: planName,
    })
}
```

### `CollaboratorService`

Add field `AuditService CollaboratorServiceAuditService`.

**`AddCollaborator`** — after email resolution:

```go
if s.AuditService != nil {
    _ = s.AuditService.LogEvent(ctx, appID, &nonblocking.SiteAdminAppCollaboratorAddedEventPayload{
        AppID:              appID,
        CollaboratorID:     newCollaborator.ID,
        CollaboratorUserID: newCollaborator.UserID,
        UserEmail:          userEmail,
        Role:               string(newCollaborator.Role),
    })
}
```

**`RemoveCollaborator`** — capture `deleted` before returning, then resolve email and log:

```go
emailMap, _ := s.AdminAPI.ResolveUserEmails(ctx, []string{deleted.UserID})
if s.AuditService != nil {
    _ = s.AuditService.LogEvent(ctx, deleted.AppID, &nonblocking.SiteAdminAppCollaboratorDeletedEventPayload{
        AppID:                 deleted.AppID,
        CollaboratorID:        deleted.ID,
        CollaboratorUserID:    deleted.UserID,
        CollaboratorUserEmail: emailMap[deleted.UserID],
    })
}
```

**`PromoteCollaborator`** — capture `demotedOwner` (may be nil), resolve emails for both, then log:

```go
if s.AuditService != nil {
    payload := &nonblocking.SiteAdminAppCollaboratorPromotedEventPayload{
        AppID:                  appID,
        NewOwnerCollaboratorID: promoted.ID,
        NewOwnerUserID:         promoted.UserID,
        NewOwnerUserEmail:      emailMap[promoted.UserID],
    }
    if demotedOwner != nil {
        payload.DemotedEditorCollaboratorID = demotedOwner.ID
        payload.DemotedEditorUserID = demotedOwner.UserID
        payload.DemotedEditorUserEmail = emailMap[demotedOwner.UserID]
    }
    _ = s.AuditService.LogEvent(ctx, appID, payload)
}
```

### Error handling convention

Audit log failures must not affect mutation success. Silently ignore the error (`_ = err`). No logging is added — the siteadmin service package has no logger.

---

## Step 6 — Unit tests

Tests are added in `plan_test.go` and `collaborator_test.go`. A `fakeAuditService` records events:

```go
type fakeAuditService struct {
    logged []event.NonBlockingPayload
}

func (f *fakeAuditService) LogEvent(_ context.Context, _ string, p event.NonBlockingPayload) error {
    f.logged = append(f.logged, p)
    return nil
}
```

| Service | Method | Assertion |
|---|---|---|
| `PlanService` | `ChangeAppPlan` success | payload is `SiteAdminAppPlanUpdatedEventPayload` with correct `OldPlan`/`NewPlan` |
| `CollaboratorService` | `AddCollaborator` success | payload has correct `CollaboratorID`, `UserEmail`, `Role="editor"` |
| `CollaboratorService` | `RemoveCollaborator` success | payload has correct `CollaboratorID`, `CollaboratorUserID` |
| `CollaboratorService` | `PromoteCollaborator` success | payload has correct new owner + demoted editor fields |
| `CollaboratorService` | `PromoteCollaborator` (no prior owner) | `DemotedEditor*` fields are empty |
| Any | `AuditService` is nil | Mutation still succeeds |

---

## Step 7 — E2E tests

All siteadmin audit records are stored under `app_id = SITEADMIN_AUTHGEAR_APP_ID` (`e2e-portal` in e2e). The affected app is identified via `data->'payload'->>'app_id'`.

### `e2e/tests/siteadmin/plans.test.yaml` — after `change_plan_to_enterprise`

```yaml
- name: audit_change_plan
  action: audit_query
  audit_query: |
    SELECT
      activity_type,
      data->'payload'->>'app_id'   AS affected_app_id,
      data->'payload'->>'old_plan' AS old_plan,
      data->'payload'->>'new_plan' AS new_plan
    FROM _audit_log
    WHERE app_id = 'e2e-portal'
      AND activity_type = 'site_admin.app.plan.updated'
      AND data->'payload'->>'app_id' = 'e2e-siteadmin-app-beta'
    ORDER BY created_at DESC
    LIMIT 1
  audit_query_output:
    rows: |
      [
        {
          "activity_type": "site_admin.app.plan.updated",
          "affected_app_id": "e2e-siteadmin-app-beta",
          "old_plan": "[[string]]",
          "new_plan": "enterprise"
        }
      ]
```

### `e2e/tests/siteadmin/collaborators.test.yaml`

After `add_collaborator`:

```yaml
- name: audit_add_collaborator
  action: audit_query
  audit_query: |
    SELECT activity_type,
      data->'payload'->>'collaborator_id' AS collaborator_id,
      data->'payload'->>'user_email'      AS user_email,
      data->'payload'->>'role'            AS role
    FROM _audit_log
    WHERE app_id = 'e2e-portal'
      AND activity_type = 'site_admin.app.collaborator.added'
      AND data->'payload'->>'app_id' = 'e2e-collab-beta'
    ORDER BY created_at DESC LIMIT 1
  audit_query_output:
    rows: |
      [{"activity_type":"site_admin.app.collaborator.added","collaborator_id":"[[string]]","user_email":"owner@example.com","role":"editor"}]
```

After `remove editor`:

```yaml
- name: audit_delete_collaborator
  action: audit_query
  audit_query: |
    SELECT activity_type,
      data->'payload'->>'collaborator_id' AS collaborator_id,
      data->'payload'->>'user_email'      AS user_email
    FROM _audit_log
    WHERE app_id = 'e2e-portal'
      AND activity_type = 'site_admin.app.collaborator.deleted'
      AND data->'payload'->>'app_id' = 'e2e-collab-beta'
    ORDER BY created_at DESC LIMIT 1
  audit_query_output:
    rows: |
      [{"activity_type":"site_admin.app.collaborator.deleted","collaborator_id":"[[string]]","user_email":"owner@example.com"}]
```

After `promote editor to owner`:

```yaml
- name: audit_promote_collaborator
  action: audit_query
  audit_query: |
    SELECT activity_type,
      data->'payload'->>'new_owner_user_email'      AS new_owner_email,
      data->'payload'->>'demoted_editor_user_email' AS demoted_email
    FROM _audit_log
    WHERE app_id = 'e2e-portal'
      AND activity_type = 'site_admin.app.collaborator.promoted'
      AND data->'payload'->>'app_id' = 'e2e-collab-gamma'
    ORDER BY created_at DESC LIMIT 1
  audit_query_output:
    rows: |
      [{"activity_type":"site_admin.app.collaborator.promoted","new_owner_email":"editor@example.com","demoted_email":"gamma-owner@example.com"}]
```

---

## Step 8 — Filter admin API audit query to known types

**Files:** `pkg/admin/graphql/audit_log.go`, `pkg/admin/graphql/query.go`

The admin API `auditLogs` query resolves `activityType` against the non-null `AuditLogActivityType` enum. If rows with unknown types (e.g. `site_admin.*`) reach the resolver, enum serialization panics.

Fix: derive a `knownAuditLogActivityTypes []string` from the enum at init time, and default the query to that whitelist when no `activityTypes` argument is provided.

```go
// audit_log.go — after the enum definition
var knownAuditLogActivityTypes = func() []string {
    values := auditLogActivityType.Values()
    result := make([]string, 0, len(values))
    for _, v := range values {
        result = append(result, v.Value.(string))
    }
    return result
}()
```

```go
// query.go — in the auditLogs resolver
var activityTypes []string
if arr, ok := p.Args["activityTypes"].([]any); ok {
    for _, v := range arr {
        if s, ok := v.(string); ok {
            activityTypes = append(activityTypes, s)
        }
    }
}
if len(activityTypes) == 0 {
    activityTypes = knownAuditLogActivityTypes
}
```

Site admin types are not in the `AuditLogActivityType` enum and are therefore automatically excluded. The whitelist stays in sync with the enum — no manual maintenance needed.

---

## Atomic commit plan

| # | Commit message | Files |
|---|---|---|
| 1 | `Add implementation plan for site admin audit log` | `docs/plans/siteadmin-api/08-audit-log.md` |
| 2 | `Add TriggeredBySiteAdmin event constant` | `pkg/api/event/context.go` |
| 3 | `Add site admin audit event payload types` | `pkg/api/event/nonblocking/siteadmin_app_*.go` |
| 4 | `Add SiteAdminAuditService and wire audit write handle` | `pkg/siteadmin/service/audit.go`, `pkg/siteadmin/service/deps.go`, `pkg/siteadmin/deps.go`, `pkg/siteadmin/wire_gen.go` |
| 5 | `Emit audit log from PlanService.ChangeAppPlan` | `pkg/siteadmin/service/plan.go`, `pkg/siteadmin/service/plan_test.go` |
| 6 | `Emit audit log from CollaboratorService mutations` | `pkg/siteadmin/service/collaborator.go`, `pkg/siteadmin/service/collaborator_test.go` |
| 7 | `Add e2e audit assertions for site admin mutations` | `e2e/tests/siteadmin/plans.test.yaml`, `e2e/tests/siteadmin/collaborators.test.yaml` |
| 8 | `Filter audit log query to known activity types by default` | `pkg/admin/graphql/audit_log.go`, `pkg/admin/graphql/query.go` |

Run after commit 4: `go build ./pkg/siteadmin/...`  
Run after commits 5–6: `go test ./pkg/siteadmin/...`  
Run after commit 7: `make -C e2e run`
