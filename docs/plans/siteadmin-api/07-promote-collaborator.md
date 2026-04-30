# 07 — Promote Collaborator to Owner

## Goal / Scope

Implement `POST /api/v1/apps/{app_id}/collaborators/{collaborator_id}/promote`.

The target collaborator becomes owner; the current owner is demoted to editor. Both role changes happen in a single database transaction, with `updated_at` stamped on each updated row. Email resolution happens outside the transaction (Admin API call).

No request body. Returns the promoted collaborator as a `Collaborator` object.

---

## Schema Change

Add `updated_at` column to `_portal_app_collaborator`. Initial value is copied from `created_at`.

Migration file: `cmd/portal/cmd/cmddatabase/migrations/portal/20260429120000-add_collaborator_updated_at.sql`

```sql
-- +migrate Up

ALTER TABLE _portal_app_collaborator ADD COLUMN updated_at timestamp with time zone;
UPDATE _portal_app_collaborator SET updated_at = created_at;
ALTER TABLE _portal_app_collaborator ALTER COLUMN updated_at SET NOT NULL;

-- +migrate Down

ALTER TABLE _portal_app_collaborator DROP COLUMN updated_at;
```

---

## Existing Code Audit

No model change. INSERT paths set `updated_at = created_at` using the existing `CreatedAt` field already on the struct (or the `now` variable already in scope). UPDATE paths stamp `now`.

| File | Operation | Change |
|---|---|---|
| `pkg/portal/service/collaborator.go` `CreateCollaborator` | INSERT | Add `"updated_at"` column with value `c.CreatedAt` |
| `pkg/siteadmin/service/collaborator.go` `CreateCollaborator` | INSERT | Add `"updated_at"` column with value `c.CreatedAt` |
| `cmd/portal/internal/collaborator.go` `insertCollaborator` | INSERT | Add `"updated_at"` column with value `now` (already in scope) |
| `cmd/portal/internal/collaborator.go` `updateCollaboratorRole` | UPDATE | Add `Set("updated_at", time.Now().UTC())` |

`model.Collaborator`, `NewCollaborator`, `selectCollaborator`, and `scanCollaborator` are **not changed**.

### `pkg/portal/service/collaborator.go` — `CreateCollaborator`

```go
Columns("id", "app_id", "user_id", "created_at", "updated_at", "role").
Values(c.ID, c.AppID, c.UserID, c.CreatedAt, c.CreatedAt, c.Role)
```

### `pkg/siteadmin/service/collaborator.go` — `CreateCollaborator`

```go
s.SQLBuilder.
    Insert(s.SQLBuilder.TableName("_portal_app_collaborator")).
    Columns("id", "app_id", "user_id", "created_at", "updated_at", "role").
    Values(c.ID, c.AppID, c.UserID, c.CreatedAt, c.CreatedAt, c.Role)
```

### `cmd/portal/internal/collaborator.go` — `insertCollaborator` + `updateCollaboratorRole`

`insertCollaborator` (`now` already declared in the function):
```go
Columns("id", "app_id", "user_id", "created_at", "updated_at", "role").
Values(id, appID, userID, now, now, role)
```

`updateCollaboratorRole`:
```go
func updateCollaboratorRole(ctx context.Context, tx *sql.Tx, id string, role model.CollaboratorRole) error {
    now := time.Now().UTC()
    builder := newSQLBuilder().Update(pq.QuoteIdentifier("_portal_app_collaborator")).
        Set("role", role).
        Set("updated_at", now).
        Where("id = ?", id)
    ...
}
```

---

## New Service Methods

### `pkg/siteadmin/service/collaborator.go`

**New error:**
```go
var ErrCollaboratorAlreadyOwner = apierrors.AlreadyExists.WithReason("CollaboratorAlreadyOwner").New("collaborator is already the owner")
```

**`UpdateCollaborator` added to `CollaboratorServiceStore` interface:**
```go
UpdateCollaborator(ctx context.Context, c *model.Collaborator) error
```

**`CollaboratorStore.UpdateCollaborator` implementation:**
```go
func (s *CollaboratorStore) UpdateCollaborator(ctx context.Context, c *model.Collaborator) error {
    now := s.Clock.NowUTC()
    _, err := s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
        Update(s.SQLBuilder.TableName("_portal_app_collaborator")).
        Set("role", c.Role).
        Set("updated_at", now).
        Where("id = ?", c.ID),
    )
    return err
}
```

**`CollaboratorService.PromoteCollaborator`:**
```go
func (s *CollaboratorService) PromoteCollaborator(ctx context.Context, appID string, collaboratorID string) (*siteadmin.Collaborator, error) {
    var promoted *model.Collaborator
    err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
        target, err := s.Store.GetCollaborator(ctx, collaboratorID)
        if err != nil {
            return err
        }
        if target.AppID != appID {
            return portalservice.ErrCollaboratorNotFound
        }
        if target.Role == model.CollaboratorRoleOwner {
            return ErrCollaboratorAlreadyOwner
        }

        all, err := s.Store.ListCollaborators(ctx, appID)
        if err != nil {
            return err
        }
        var currentOwner *model.Collaborator
        for _, c := range all {
            if c.Role == model.CollaboratorRoleOwner {
                currentOwner = c
                break
            }
        }

        target.Role = model.CollaboratorRoleOwner
        if err := s.Store.UpdateCollaborator(ctx, target); err != nil {
            return err
        }
        if currentOwner != nil {
            currentOwner.Role = model.CollaboratorRoleEditor
            if err := s.Store.UpdateCollaborator(ctx, currentOwner); err != nil {
                return err
            }
        }

        promoted = target
        return nil
    })
    if err != nil {
        return nil, err
    }

    emailMap, err := s.AdminAPI.ResolveUserEmails(ctx, []string{promoted.UserID})
    if err != nil {
        return nil, err
    }

    return &siteadmin.Collaborator{
        Id:        promoted.ID,
        AppId:     promoted.AppID,
        UserId:    promoted.UserID,
        UserEmail: emailMap[promoted.UserID],
        Role:      siteadmin.CollaboratorRole(promoted.Role),
        CreatedAt: promoted.CreatedAt,
    }, nil
}
```

---

## Transport Handler

`pkg/siteadmin/transport/handler_collaborator_promote.go` — replace stub:

```go
package transport

import (
    "context"
    "net/http"

    "github.com/authgear/authgear-server/pkg/api/siteadmin"
    "github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorPromoteRoute(route httproute.Route) httproute.Route {
    return route.WithMethods("OPTIONS", "POST").
        WithPathPattern("/api/v1/apps/:appID/collaborators/:collaboratorID/promote")
}

type CollaboratorPromoteService interface {
    PromoteCollaborator(ctx context.Context, appID string, collaboratorID string) (*siteadmin.Collaborator, error)
}

type CollaboratorPromoteHandler struct {
    Service CollaboratorPromoteService
}

type CollaboratorPromoteParams struct {
    AppID          string
    CollaboratorID string
}

func parseCollaboratorPromoteParams(r *http.Request) CollaboratorPromoteParams {
    return CollaboratorPromoteParams{
        AppID:          httproute.GetParam(r, "appID"),
        CollaboratorID: httproute.GetParam(r, "collaboratorID"),
    }
}

func (h *CollaboratorPromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    params := parseCollaboratorPromoteParams(r)

    collaborator, err := h.Service.PromoteCollaborator(r.Context(), params.AppID, params.CollaboratorID)
    if err != nil {
        writeError(w, r, err)
        return
    }

    SiteAdminAPISuccessResponse{Body: collaborator}.WriteTo(w)
}
```

## DI Wiring

`pkg/siteadmin/deps.go` — add:
```go
wire.Bind(new(transport.CollaboratorPromoteService), new(*siteadminservice.CollaboratorService)),
```

---

## Test Plan

Add `UpdateCollaborator` to `fakeCollaboratorStore` in `pkg/siteadmin/service/collaborator_test.go`:

```go
updated []*portalmodel.Collaborator

func (f *fakeCollaboratorStore) UpdateCollaborator(_ context.Context, c *portalmodel.Collaborator) error {
    f.updated = append(f.updated, c)
    if existing, ok := f.existingByID[c.ID]; ok {
        existing.Role = c.Role
    }
    return nil
}
```

Four new test cases under `TestCollaboratorService`:

1. **PromoteCollaborator succeeds** — target is editor, current owner exists; after call: target role is owner, old owner role is editor; returned collaborator has resolved email; AdminAPI called outside TX.
2. **PromoteCollaborator returns not found when collaboratorID is missing** — `existingByID` is empty.
3. **PromoteCollaborator returns not found on cross-app access** — collaborator exists but `AppID != appID`.
4. **PromoteCollaborator returns AlreadyOwner when target is already owner** — collaborator has `Role == owner`.

---

## Error Table

| Condition | Error | HTTP |
|---|---|---|
| collaboratorID not in DB | `portalservice.ErrCollaboratorNotFound` | 404 |
| collaborator belongs to different app | `portalservice.ErrCollaboratorNotFound` | 404 |
| collaborator is already owner | `ErrCollaboratorAlreadyOwner` | 409 |

---

## Atomic Commit Plan

### Commit 1 — DB migration: add updated_at to _portal_app_collaborator

Files:
- `cmd/portal/cmd/cmddatabase/migrations/portal/20260429120000-add_collaborator_updated_at.sql`

Standalone SQL-only commit.

### Commit 2 — Fix all INSERT and UPDATE paths for updated_at (no model change)

Files:
- `pkg/portal/service/collaborator.go` — update `CreateCollaborator` (add `updated_at = c.CreatedAt`)
- `pkg/siteadmin/service/collaborator.go` — update `CreateCollaborator` (add `updated_at = c.CreatedAt`)
- `cmd/portal/internal/collaborator.go` — update `insertCollaborator` (add `updated_at = now`), `updateCollaboratorRole` (add `Set("updated_at", now)`)

Verification: `go build ./pkg/portal/... ./pkg/siteadmin/... ./cmd/portal/...` · `make fmt`

### Commit 3 — Siteadmin service: UpdateCollaborator + ErrCollaboratorAlreadyOwner + PromoteCollaborator + tests

Files:
- `pkg/siteadmin/service/collaborator.go` — add `ErrCollaboratorAlreadyOwner`, `UpdateCollaborator` to interface and store, implement `CollaboratorService.PromoteCollaborator`
- `pkg/siteadmin/service/collaborator_test.go` — add `updated` field + `UpdateCollaborator` fake, add 4 test cases

Verification: `go test ./pkg/siteadmin/service/...` · `make fmt`

### Commit 4 — Transport interface + handler body + DI wiring + wire gen + check-tidy

The handler stub was scaffolded in Stage 3. This commit replaces it with the real implementation and wires everything up.

Files:
- `pkg/siteadmin/transport/handler_collaborator_promote.go` — add `CollaboratorPromoteService` interface, `Service` field, and `ServeHTTP` body
- `pkg/siteadmin/deps.go` — add `wire.Bind(new(transport.CollaboratorPromoteService), new(*siteadminservice.CollaboratorService))`
- `pkg/siteadmin/wire_gen.go` — regenerated via `go generate ./pkg/siteadmin/...`
- `.vettedpositions` — add new `r.Context()` positions, run `make sort-vettedpositions`

Run `make check-tidy`; stage regenerated/formatted output files and include in this commit.

Verification: `go test ./pkg/siteadmin/...` · `make fmt` · `make lint` · `make check-tidy`
