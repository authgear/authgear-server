# Part 4: Site Admin API — Collaborators API Real Data

## Context

Replace dummy data in `CollaboratorsListHandler`, `CollaboratorAddHandler`, and
`CollaboratorRemoveHandler` with real data from the global database and the portal Admin API.

The three endpoints affected are:

- `GET  /api/v1/apps/:appID/collaborators` — list all collaborators for an app
- `POST /api/v1/apps/:appID/collaborators` — add a collaborator by email
- `DELETE /api/v1/apps/:appID/collaborators/:collaboratorID` — remove a collaborator

**Key data sources:**

| Field | Source |
|---|---|
| `id`, `app_id`, `user_id`, `created_at`, `role` | `_portal_app_collaborator` table via `siteadminservice.CollaboratorStore` |
| `user_email` | Admin API GraphQL (`getUserNodes` for list; `getUsersByStandardAttribute` for add) |

**Design decisions:**

- A new `siteadminservice.CollaboratorService` holds the business logic, following the
  same pattern as `siteadminservice.AppService`.
- Email resolution for list uses a batch `getUserNodes` Admin API call (same as
  `AppService.resolveUserEmails`).
- `AddCollaborator` looks up the user by email via `getUsersByStandardAttribute`. If no
  user exists for that email, it returns `404 Not Found`. Unlike the portal flow, the
  siteadmin API does **not** create accounts or send invitation emails — it requires the
  user to already exist.
- `RemoveCollaborator` verifies the collaborator belongs to the given `appID` before
  deleting (guards against cross-app ID manipulation in the URL).
- The actor user ID for Admin API calls is obtained from the validated session in the
  request context (`session.GetValidSessionInfo(ctx).UserID`), identical to how the
  `AuthzMiddleware` works.
- A shared `siteadminservice.AdminAPIService` owns all Site Admin API GraphQL calls.
  `AppService` and `CollaboratorService` both delegate to it instead of embedding their
  own request construction and response parsing logic.
- `CollaboratorService` owns the transaction boundary. Store reads and writes run inside
  `GlobalDatabase.WithTx`, but Admin API calls (`resolveUserEmails`,
  `getUsersByStandardAttribute`) must run after the transaction is closed, so the site
  admin handler does not hold a global DB connection while synchronously waiting for the
  Admin API handler to acquire one.
- The siteadmin collaborators path does **not** reuse `portalservice.CollaboratorService`
  for CRUD. That portal service already opens its own transactions, so wrapping it in
  siteadmin service transactions would reintroduce nested transaction behavior. Instead,
  Site Admin uses a local `CollaboratorStore` that operates on the existing transaction
  context directly.

---

## Architecture Overview

```
CollaboratorsListHandler / CollaboratorAddHandler / CollaboratorRemoveHandler (transport)
    │  depends on
    ▼
CollaboratorService (pkg/siteadmin/service/collaborator.go)
    │  depends on
    ├── CollaboratorServiceStore → *siteadminservice.CollaboratorStore  (SQL CRUD)
    └── AdminAPIService          → shared siteadmin Admin API helper    (GraphQL)
```

### `ListCollaborators` flow

```
1. GlobalDatabase.WithTx + CollaboratorStore.ListCollaborators(ctx, appID) → []*model.Collaborator
2. Collect all userIDs from result
3. After the DB work is finished, Admin API batch call: getUserNodes(userIDs) → map[userID]email
4. Map each *model.Collaborator → siteadmin.Collaborator (fill UserEmail)
5. Return []siteadmin.Collaborator
```

### `AddCollaborator` flow

```
1. session.GetValidSessionInfo(ctx) → actorUserID
2. Before opening a DB transaction, Admin API: getUsersByStandardAttribute("email", userEmail) → []userID
3. If empty → 404 Not Found ("user not found")
4. Take first userID (email is unique within an Authgear app)
5. GlobalDatabase.WithTx:
   a. CollaboratorStore.GetCollaboratorByAppAndUser(appID, userID) — if found → 409 Duplicate
   b. CollaboratorStore.NewCollaborator(appID, userID, "editor")
   c. CollaboratorStore.CreateCollaborator(c)
6. Return siteadmin.Collaborator (UserEmail filled from input)
```

### `RemoveCollaborator` flow

```
1. GlobalDatabase.WithTx:
   a. CollaboratorStore.GetCollaborator(collaboratorID) — if not found → 404
   b. Verify collaborator.AppID == appID → if mismatch → 404
   c. CollaboratorStore.DeleteCollaborator(collaborator)
2. Return {} (empty JSON object)
```

---

## Key Dependencies

| What | Where |
|---|---|
| `siteadminservice.AdminAPIService` | `pkg/siteadmin/service/admin_api.go` |
| `siteadminservice.CollaboratorStore` | `pkg/siteadmin/service/collaborator.go` |
| `siteadminservice.CollaboratorService` | `pkg/siteadmin/service/collaborator.go` |
| `portalservice.ErrCollaboratorNotFound` | `pkg/portal/service/collaborator.go` |
| `portalservice.ErrCollaboratorDuplicate` | `pkg/portal/service/collaborator.go` |
| `portalservice.AdminAPIService.SelfDirector` | `pkg/portal/service/admin_api.go` |
| `session.GetValidSessionInfo` | `pkg/portal/session/context.go` |
| `model.CollaboratorRoleEditor` | `pkg/portal/model/collaborator.go` |
| `siteadmin.Collaborator` | `pkg/api/siteadmin/gen.go` |

---

## Implemented Files

### Created

- `pkg/siteadmin/service/admin_api.go` — shared Site Admin `AdminAPIService` with
  `FindUserIDsByEmail` and `ResolveUserEmails`
- `pkg/siteadmin/service/collaborator.go` — `CollaboratorStore`,
  `CollaboratorService`, collaborator scanning helpers, and duplicate detection
- `pkg/siteadmin/service/collaborator_test.go` — service-level tests for list, add,
  remove, and transaction-boundary behavior

### Modified

- `pkg/siteadmin/service/app.go` — delegates Admin API lookups to the shared
  `AdminAPIService`
- `pkg/siteadmin/service/deps.go` — wires `AdminAPIService`, `CollaboratorStore`,
  `CollaboratorService`, and the shared HTTP client
- `pkg/siteadmin/deps.go` — binds `portalservice.AdminAPIService` into
  `siteadminservice.SiteAdminAdminAPI` and adds transport bindings for the collaborator
  handlers
- `pkg/siteadmin/transport/handler_collaborators_list.go` — real list handler via
  `CollaboratorsListService`
- `pkg/siteadmin/transport/handler_collaborator_add.go` — real add handler via
  `CollaboratorAddService`, removing all dummy data
- `pkg/siteadmin/transport/handler_collaborator_remove.go` — real remove handler via
  `CollaboratorRemoveService`
- `pkg/siteadmin/wire_gen.go` — regenerated after DI changes

### Implementation Notes

- The shared Site Admin HTTP client type is `SiteAdminHTTPClient`, not
  `AppServiceHTTPClient`.
- `AppService` and `CollaboratorService` both depend on `*siteadminservice.AdminAPIService`.
- `AuthzMiddleware` still reuses `*portalservice.CollaboratorService` for the existing
  authorization check. The new collaborators CRUD path is separate and uses
  `CollaboratorStore`.
- `AddCollaborator` still does a pre-check for an existing collaborator, but
  `CollaboratorStore.CreateCollaborator` also maps the database unique constraint to
  `portalservice.ErrCollaboratorDuplicate` so the write path is safe against races.

---

## Implementation Roadmap: 5 Atomic Commits

### **Commit 1: Extract shared `AdminAPIService` helper**

**Files Created:**
- `pkg/siteadmin/service/admin_api.go`

**Files Modified:**
- `pkg/siteadmin/service/app.go` — replace inline `findUserIDsByEmail` / `resolveUserEmails` logic with calls to the shared `AdminAPIService`

**Scope:** Refactor only. No behavior change, no collaborator feature code yet.
This commit prepares `CollaboratorService` to reuse the same generic Admin API
service as `AppService` instead of duplicating GraphQL request code.

**Commit Message:** `"Extract shared siteadmin AdminAPIService"`

---

### **Commit 2: Add `CollaboratorService` to siteadmin service layer**

**Files Created:**
- `pkg/siteadmin/service/collaborator.go`

**Files Modified:**
- `pkg/siteadmin/service/deps.go` — add `CollaboratorStore` binding, `AdminAPIService`,
  and `wire.Struct(new(CollaboratorService), "*")`

**Scope:** Business logic only — no DI wiring, no handler changes.

**Commit Message:** `"Add siteadmin collaborator service layer"`

---

### **Commit 3: Add service interfaces to transport handlers**

**Files Modified:**
- `pkg/siteadmin/transport/handler_collaborators_list.go` — add `CollaboratorsListService` interface + `Service` field; remove `TODO` dummy call
- `pkg/siteadmin/transport/handler_collaborator_add.go` — add `CollaboratorAddService` interface + `Service` field; remove all dummy data (`dummyCollaborators` map, `dummyCollaboratorsForApp` helper)
- `pkg/siteadmin/transport/handler_collaborator_remove.go` — add `CollaboratorRemoveService` interface + `Service` field; remove dummy-data lookup

**Scope:** Handler structs updated; service interfaces declared. The handlers now compile
only when a `Service` field is provided. `deps.go` bindings are added in the next commit.

**Build note:** `go build ./pkg/siteadmin/transport/...` will fail until Commit 4
provides the wire bindings; `pkg/siteadmin/transport/...` itself compiles fine in
isolation.

**Commit Message:** `"Wire service interfaces into collaborator transport handlers"`

---

### **Commit 4: Wire CollaboratorService into DI and regenerate**

**Files Modified:**
- `pkg/siteadmin/deps.go` — bind `siteadminservice.SiteAdminAdminAPI` to `*portalservice.AdminAPIService`; add transport bindings for the three collaborator handler service interfaces
- `pkg/siteadmin/wire_gen.go` — regenerated

**Build Steps:**
```bash
wire gen ./pkg/siteadmin/...
go build ./pkg/siteadmin/...
go build ./cmd/portal/...
```

**Commit Message:** `"Wire CollaboratorService into siteadmin DI and regenerate"`

---

### **Commit 5: Replace handler bodies with real service calls**

**Files Modified:**
- `pkg/siteadmin/transport/handler_collaborators_list.go` — call `h.Service.ListCollaborators`
- `pkg/siteadmin/transport/handler_collaborator_add.go` — call `h.Service.AddCollaborator`
- `pkg/siteadmin/transport/handler_collaborator_remove.go` — call `h.Service.RemoveCollaborator`

**Scope:** Handler `ServeHTTP` bodies only — no interface or DI changes.

**Commit Message:** `"Replace collaborator handler stubs with real service calls"`

---

## Dependency Graph

```
Commit 1 (shared AdminAPIService)
    ↓
Commit 2 (CollaboratorService + service/deps.go)
    ↓
Commit 3 (transport handler interfaces + Service fields)
    ↓
Commit 4 (deps.go wiring + wire gen)
    ↓
Commit 5 (handler ServeHTTP bodies)
```

**Key Properties:**
- ✅ Each commit is independently reviewable
- ✅ No mixing of concerns (shared refactor → service → transport interfaces → DI wiring → handler bodies)
- ✅ Build passes after Commit 4
- ✅ Endpoints return real data after Commit 5

---

## Verification

### List collaborators (real data)

```bash
curl -s -H "Authorization: Bearer <token>" \
  http://localhost:3005/api/v1/apps/myapp/collaborators | jq .
```

Expected response:
```json
{
  "collaborators": [
    {
      "id": "...",
      "app_id": "myapp",
      "user_id": "...",
      "user_email": "alice@example.com",
      "role": "owner",
      "created_at": "2024-01-15T08:00:00Z"
    }
  ]
}
```

### Add collaborator (existing user)

```bash
curl -s -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_email":"bob@example.com"}' \
  http://localhost:3005/api/v1/apps/myapp/collaborators | jq .
```

Expected response: `200 OK` with `siteadmin.Collaborator` JSON (role = "editor").

### Add collaborator (user not found)

```bash
curl -i -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_email":"nobody@example.com"}' \
  http://localhost:3005/api/v1/apps/myapp/collaborators
```

Expected response: `404 Not Found`.

### Add collaborator (duplicate)

Repeat the same add request for an existing collaborator.

Expected response: `409 Conflict` with `reason: "CollaboratorDuplicate"`.

### Remove collaborator

```bash
curl -s -X DELETE \
  -H "Authorization: Bearer <token>" \
  http://localhost:3005/api/v1/apps/myapp/collaborators/<collaboratorID>
```

Expected response: `200 OK` with `{}`.

### Remove collaborator (wrong app)

Use a `collaboratorID` that belongs to a different app.

Expected response: `404 Not Found` (cross-app mismatch is treated as not found).
