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
| `id`, `app_id`, `user_id`, `created_at`, `role` | `_portal_app_collaborator` table via `portalservice.CollaboratorService` |
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
- `CollaboratorService` owns the transaction boundary. Store reads and writes run inside
  `GlobalDatabase.WithTx`, but Admin API calls (`resolveUserEmails`,
  `getUsersByStandardAttribute`) must run after the transaction is closed, so the site
  admin handler does not hold a global DB connection while synchronously waiting for the
  Admin API handler to acquire one.
- SQL operations are delegated to `portalservice.CollaboratorService` via a narrow
  interface, reusing the existing partial struct already wired for AuthzMiddleware (we
  add `Clock` to the partial struct to support `NewCollaborator`).

---

## Architecture Overview

```
CollaboratorsListHandler / CollaboratorAddHandler / CollaboratorRemoveHandler (transport)
    │  depends on
    ▼
CollaboratorService (pkg/siteadmin/service/collaborator.go)
    │  depends on
    ├── CollaboratorServiceStore  → *portalservice.CollaboratorService  (SQL CRUD)
    ├── CollaboratorServiceAdminAPI → *portalservice.AdminAPIService    (email resolution)
    └── CollaboratorServiceHTTPClient                                   (HTTP for GraphQL)
```

### `ListCollaborators` flow

```
1. portalservice.CollaboratorService.ListCollaborators(ctx, appID) → []*model.Collaborator
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
   a. GetCollaboratorByAppAndUser(appID, userID) — if found → 409 Duplicate
   b. NewCollaborator(appID, userID, "editor")
   c. CreateCollaborator(c)
6. Return siteadmin.Collaborator (UserEmail filled from input)
```

### `RemoveCollaborator` flow

```
1. GlobalDatabase.WithTx:
   a. GetCollaborator(collaboratorID) — if not found → 404
   b. Verify collaborator.AppID == appID → if mismatch → 404
   c. DeleteCollaborator(collaborator)
2. Return {} (empty JSON object)
```

---

## Key Dependencies

| What | Where |
|---|---|
| `portalservice.CollaboratorService` | `pkg/portal/service/collaborator.go` |
| `portalservice.CollaboratorService.ListCollaborators` | list all collaborators for an app |
| `portalservice.CollaboratorService.GetCollaboratorByAppAndUser` | check duplicate on add |
| `portalservice.CollaboratorService.NewCollaborator` | build collaborator struct (needs `Clock`) |
| `portalservice.CollaboratorService.CreateCollaborator` | persist new collaborator |
| `portalservice.CollaboratorService.GetCollaborator` | fetch by ID on remove |
| `portalservice.CollaboratorService.DeleteCollaborator` | remove from DB |
| `portalservice.ErrCollaboratorNotFound` | `pkg/portal/service/collaborator.go` |
| `portalservice.ErrCollaboratorDuplicate` | `pkg/portal/service/collaborator.go` |
| `portalservice.AdminAPIService.SelfDirector` | `pkg/portal/service/admin_api.go` |
| `session.GetValidSessionInfo` | `pkg/portal/session/context.go` |
| `relay.ToGlobalID` / `relay.FromGlobalID` | `pkg/graphqlgo/relay` |
| `graphqlutil.DoParams` / `graphqlutil.HTTPDo` | `pkg/util/graphqlutil/http_do.go` |
| `model.CollaboratorRoleEditor` | `pkg/portal/model/collaborator.go` |
| `siteadmin.Collaborator` | `pkg/api/siteadmin/gen.go` |

---

## Files to Create

### 1. `pkg/siteadmin/service/collaborator.go`

```go
package service

import (
	"context"
	"net/http"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

// ---- Narrow interfaces -------------------------------------------------------

type CollaboratorServiceStore interface {
	ListCollaborators(ctx context.Context, appID string) ([]*model.Collaborator, error)
	GetCollaborator(ctx context.Context, id string) (*model.Collaborator, error)
	GetCollaboratorByAppAndUser(ctx context.Context, appID string, userID string) (*model.Collaborator, error)
	NewCollaborator(appID string, userID string, role model.CollaboratorRole) *model.Collaborator
	CreateCollaborator(ctx context.Context, c *model.Collaborator) error
	DeleteCollaborator(ctx context.Context, c *model.Collaborator) error
}

type CollaboratorServiceAdminAPI interface {
	SelfDirector(ctx context.Context, actorUserID string, usage portalservice.Usage) (func(*http.Request), error)
}

// ---- CollaboratorService -----------------------------------------------------

type CollaboratorService struct {
	GlobalDatabase AppServiceDatabase
	Store          CollaboratorServiceStore
	AdminAPI       CollaboratorServiceAdminAPI
	HTTPClient     AppServiceHTTPClient
}

func (s *CollaboratorService) ListCollaborators(ctx context.Context, appID string) ([]siteadmin.Collaborator, error) {
	collaborators, err := s.Store.ListCollaborators(ctx, appID)
	if err != nil {
		return nil, err
	}
	if len(collaborators) == 0 {
		return []siteadmin.Collaborator{}, nil
	}

	userIDs := make([]string, len(collaborators))
	for i, c := range collaborators {
		userIDs[i] = c.UserID
	}

	// Admin API call — outside a DB transaction.
	sessionInfo := session.GetValidSessionInfo(ctx)
	actorUserID := sessionInfo.UserID
	emailMap, err := s.resolveUserEmails(ctx, actorUserID, userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]siteadmin.Collaborator, len(collaborators))
	for i, c := range collaborators {
		result[i] = siteadmin.Collaborator{
			Id:        c.ID,
			AppId:     c.AppID,
			UserId:    c.UserID,
			UserEmail: emailMap[c.UserID],
			Role:      siteadmin.CollaboratorRole(c.Role),
			CreatedAt: c.CreatedAt,
		}
	}
	return result, nil
}

func (s *CollaboratorService) AddCollaborator(ctx context.Context, appID string, userEmail string) (*siteadmin.Collaborator, error) {
	sessionInfo := session.GetValidSessionInfo(ctx)
	actorUserID := sessionInfo.UserID

	// Admin API call — must happen outside a DB transaction.
	userIDs, err := s.findUserIDsByEmail(ctx, actorUserID, userEmail)
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		return nil, portalservice.ErrCollaboratorNotFound
	}
	targetUserID := userIDs[0]

	var newCollab *model.Collaborator
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		_, e := s.Store.GetCollaboratorByAppAndUser(ctx, appID, targetUserID)
		if e == nil {
			return portalservice.ErrCollaboratorDuplicate
		}
		if !isNotFound(e) {
			return e
		}

		newCollab = s.Store.NewCollaborator(appID, targetUserID, model.CollaboratorRoleEditor)
		return s.Store.CreateCollaborator(ctx, newCollab)
	})
	if err != nil {
		return nil, err
	}

	out := siteadmin.Collaborator{
		Id:        newCollab.ID,
		AppId:     newCollab.AppID,
		UserId:    newCollab.UserID,
		UserEmail: userEmail, // already known from input
		Role:      siteadmin.CollaboratorRole(newCollab.Role),
		CreatedAt: newCollab.CreatedAt,
	}
	return &out, nil
}

func (s *CollaboratorService) RemoveCollaborator(ctx context.Context, appID string, collaboratorID string) error {
	return s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		c, err := s.Store.GetCollaborator(ctx, collaboratorID)
		if err != nil {
			return err
		}
		if c.AppID != appID {
			// Treat a cross-app mismatch as not found to avoid leaking info.
			return portalservice.ErrCollaboratorNotFound
		}
		return s.Store.DeleteCollaborator(ctx, c)
	})
}

// resolveUserEmails batch-fetches emails for the given user IDs via Admin API.
// Reuses the same getUserNodes GraphQL query as AppService.resolveUserEmails.
func (s *CollaboratorService) resolveUserEmails(ctx context.Context, actorUserID string, userIDs []string) (map[string]string, error) {
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
		return nil, fmt.Errorf("unexpected graphql errors: %v", result.Errors)
	}

	emailMap := make(map[string]string, len(userIDs))
	data := result.Data.(map[string]interface{})
	nodes, _ := data["nodes"].([]interface{})
	for _, node := range nodes {
		n, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		globalID, _ := n["id"].(string)
		resolved := relay.FromGlobalID(globalID)
		if resolved == nil || resolved.ID == "" {
			continue
		}
		attrs, _ := n["standardAttributes"].(map[string]interface{})
		email, _ := attrs["email"].(string)
		emailMap[resolved.ID] = email
	}
	return emailMap, nil
}

// findUserIDsByEmail calls Admin API getUsersByStandardAttribute to find users
// matching the given email. Returns their raw (non-global) user IDs.
func (s *CollaboratorService) findUserIDsByEmail(ctx context.Context, actorUserID string, email string) ([]string, error) {
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
		return nil, fmt.Errorf("unexpected graphql errors: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	users, _ := data["users"].([]interface{})
	var ids []string
	for _, u := range users {
		m, ok := u.(map[string]interface{})
		if !ok {
			continue
		}
		globalID, _ := m["id"].(string)
		resolved := relay.FromGlobalID(globalID)
		if resolved == nil || resolved.ID == "" {
			continue
		}
		ids = append(ids, resolved.ID)
	}
	return ids, nil
}

// isNotFound returns true for ErrCollaboratorNotFound so the add flow can
// distinguish "not found" (expected on fresh add) from other errors.
func isNotFound(err error) bool {
	return errors.Is(err, portalservice.ErrCollaboratorNotFound)
}
```

> **Note:** `fmt` and `errors` imports are required. `AppServiceDatabase` and
> `AppServiceHTTPClient` are already defined in `app.go` in the same package — no
> duplication needed.

---

## Files to Modify

### `pkg/siteadmin/service/deps.go`

Add `CollaboratorService` to the dependency set:

```go
var DependencySet = wire.NewSet(
	wire.Struct(new(AppOwnerStore), "*"),
	wire.Bind(new(AppServiceOwnerStore), new(*AppOwnerStore)),
	wire.Struct(new(AppService), "*"),
	NewHTTPClient,
	wire.Struct(new(CollaboratorService), "*"),  // NEW
)
```

### `pkg/siteadmin/deps.go`

Three changes:

**1. Add `Clock` to the partial `portalservice.CollaboratorService` struct** (needed by
`NewCollaborator`):

```go
// Before:
wire.Struct(new(portalservice.CollaboratorService), "SQLBuilder", "SQLExecutor", "GlobalDatabase"),

// After:
wire.Struct(new(portalservice.CollaboratorService), "SQLBuilder", "SQLExecutor", "GlobalDatabase", "Clock"),
```

**2. Add second binding** for the siteadmin service layer interface:

```go
wire.Bind(new(transport.AuthzCollaboratorService), new(*portalservice.CollaboratorService)),
wire.Bind(new(siteadminservice.CollaboratorServiceStore), new(*portalservice.CollaboratorService)),  // NEW
```

**3. Add transport bindings** for the three handler interfaces and the admin API binding:

```go
// transport bindings
wire.Bind(new(transport.AppsListService), new(*siteadminservice.AppService)),
wire.Bind(new(transport.AppGetService), new(*siteadminservice.AppService)),
wire.Bind(new(transport.CollaboratorsListService), new(*siteadminservice.CollaboratorService)),  // NEW
wire.Bind(new(transport.CollaboratorAddService), new(*siteadminservice.CollaboratorService)),    // NEW
wire.Bind(new(transport.CollaboratorRemoveService), new(*siteadminservice.CollaboratorService)), // NEW

// adminAPI binding for CollaboratorService (reuse same AdminAPIService)
wire.Bind(new(siteadminservice.CollaboratorServiceAdminAPI), new(*portalservice.AdminAPIService)), // NEW
```

### `pkg/siteadmin/transport/handler_collaborators_list.go`

Replace the stub with a real implementation:

```go
package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorsListRoute(route httproute.Route) httproute.Route {
	// The OPTIONS request is handled in CollaboratorAddRoute
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/apps/:appID/collaborators")
}

type CollaboratorsListService interface {
	ListCollaborators(ctx context.Context, appID string) ([]siteadmin.Collaborator, error)
}

type CollaboratorsListHandler struct {
	Service CollaboratorsListService
}

type CollaboratorsListParams struct {
	AppID string
}

func parseCollaboratorsListParams(r *http.Request) CollaboratorsListParams {
	return CollaboratorsListParams{
		AppID: httproute.GetParam(r, "appID"),
	}
}

func (h *CollaboratorsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseCollaboratorsListParams(r)

	collaborators, err := h.Service.ListCollaborators(r.Context(), params.AppID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	response := siteadmin.CollaboratorsListResponse{
		Collaborators: collaborators,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
```

### `pkg/siteadmin/transport/handler_collaborator_add.go`

Replace the dummy data and stub with a real implementation. Remove all `dummyCollaborators`
map and related helpers (they will no longer be referenced by any handler):

```go
package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func ConfigureCollaboratorAddRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").
		WithPathPattern("/api/v1/apps/:appID/collaborators")
}

var CollaboratorAddRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"user_email": { "type": "string", "format": "email" }
		},
		"required": ["user_email"]
	}
`)

type CollaboratorAddService interface {
	AddCollaborator(ctx context.Context, appID string, userEmail string) (*siteadmin.Collaborator, error)
}

type CollaboratorAddHandler struct {
	Service CollaboratorAddService
}

type CollaboratorAddParams struct {
	AppID string
	siteadmin.AddCollaboratorRequest
}

func parseCollaboratorAddParams(r *http.Request) (CollaboratorAddParams, error) {
	var body siteadmin.AddCollaboratorRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return CollaboratorAddParams{}, err
	}

	if err := CollaboratorAddRequestSchema.Validator().ValidateValue(r.Context(), body); err != nil {
		return CollaboratorAddParams{}, err
	}

	return CollaboratorAddParams{
		AppID:                  httproute.GetParam(r, "appID"),
		AddCollaboratorRequest: body,
	}, nil
}

func (h *CollaboratorAddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseCollaboratorAddParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	collaborator, err := h.Service.AddCollaborator(r.Context(), params.AppID, params.UserEmail)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(collaborator)
}
```

### `pkg/siteadmin/transport/handler_collaborator_remove.go`

Replace the stub with a real implementation:

```go
package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureCollaboratorRemoveRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "DELETE").
		WithPathPattern("/api/v1/apps/:appID/collaborators/:collaboratorID")
}

type CollaboratorRemoveService interface {
	RemoveCollaborator(ctx context.Context, appID string, collaboratorID string) error
}

type CollaboratorRemoveHandler struct {
	Service CollaboratorRemoveService
}

type CollaboratorRemoveParams struct {
	AppID          string
	CollaboratorID string
}

func parseCollaboratorRemoveParams(r *http.Request) CollaboratorRemoveParams {
	return CollaboratorRemoveParams{
		AppID:          httproute.GetParam(r, "appID"),
		CollaboratorID: httproute.GetParam(r, "collaboratorID"),
	}
}

func (h *CollaboratorRemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := parseCollaboratorRemoveParams(r)

	if err := h.Service.RemoveCollaborator(r.Context(), params.AppID, params.CollaboratorID); err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(struct{}{})
}
```

### `pkg/siteadmin/transport/deps.go`

Update structs to use `"*"` now that `Service` fields exist:

```go
var DependencySet = wire.NewSet(
	wire.Struct(new(AppsListHandler), "*"),
	wire.Struct(new(AppGetHandler), "*"),
	wire.Struct(new(CollaboratorsListHandler), "*"),
	wire.Struct(new(CollaboratorAddHandler), "*"),
	wire.Struct(new(CollaboratorRemoveHandler), "*"),
	wire.Struct(new(MessagingUsageHandler), "*"),
	wire.Struct(new(MonthlyActiveUsersUsageHandler), "*"),
	wire.Struct(new(AuthzMiddleware), "*"),
)
```

> The struct entry format does not change (already `"*"`). The wiring becomes valid once
> `Service` fields are present on the handler structs and bound in `pkg/siteadmin/deps.go`.

### `pkg/siteadmin/wire_gen.go`

Regenerated via `wire gen ./pkg/siteadmin/...` after the deps changes. Do **not** hand-edit.

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
- `pkg/siteadmin/service/deps.go` — add `wire.Struct(new(CollaboratorService), "*")`

**Scope:** Business logic only — no DI wiring, no handler changes.

**Commit Message:** `"Add siteadmin CollaboratorService with list/add/remove"`

---

### **Commit 3: Add service interfaces to transport handlers**

**Files Modified:**
- `pkg/siteadmin/transport/handler_collaborators_list.go` — add `CollaboratorsListService` interface + `Service` field; remove `TODO` dummy call
- `pkg/siteadmin/transport/handler_collaborator_add.go` — add `CollaboratorAddService` interface + `Service` field; remove all dummy data (`dummyCollaborators` map, `dummyCollaboratorsForApp` helper)
- `pkg/siteadmin/transport/handler_collaborator_remove.go` — add `CollaboratorRemoveService` interface + `Service` field; remove dummy-data lookup

**Scope:** Handler structs updated; service interfaces declared. The handlers now compile
only when a `Service` field is provided. `deps.go` bindings are added in the next commit.

**Build note:** `go build ./pkg/siteadmin/transport/...` will fail until Commit 3
provides the wire bindings; `pkg/siteadmin/transport/...` itself compiles fine in
isolation.

**Commit Message:** `"Wire service interfaces into collaborator transport handlers"`

---

### **Commit 4: Wire CollaboratorService into DI and regenerate**

**Files Modified:**
- `pkg/siteadmin/deps.go` — extend partial `portalservice.CollaboratorService` with `Clock`; add `CollaboratorServiceStore` binding; add `CollaboratorServiceAdminAPI` binding; add transport bindings for the three handler service interfaces
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
