# Part 2: Site Admin API Authorization

## Context

The Site Admin API must only be accessible to users who are collaborators of the portal's own Authgear app (`AuthgearConfig.AppID`). This plan adds two middleware layers:

1. **Session middleware** — reuse `pkg/portal/session/SessionInfoMiddleware` to parse the JWT/cookie and inject `*model.SessionInfo` into the request context.
2. **Authz middleware** — a new siteadmin-specific middleware that reads the session info and checks whether the user is a collaborator of `AuthgearConfig.AppID`. Returns `401 Unauthenticated` or `403 Forbidden` if not.

**Key design decisions:**
- Reuse the existing portal `SessionInfoMiddleware` unchanged — no forking.
- The collaborator check uses `CollaboratorService.GetCollaboratorByAppAndUser` (same as the portal's `AuthzService.CheckAccessOfViewer`).
- The app being checked is always `AuthgearConfig.AppID` (the portal's own Authgear app ID), not the app ID from the request URL.
- Error kinds reuse `service.ErrUnauthenticated` and `service.ErrForbidden` from `pkg/portal/service/authz.go`.

---

## Architecture Overview

**Updated middleware chain:**

```
Request
  1. OtelMiddleware            → OpenTelemetry tracing
  2. PanicMiddleware           → Panic recovery
  3. BodyLimitMiddleware       → Max request body size
  4. SentryMiddleware          → Error capture
  5. CORSMiddleware            → CORS headers
  6. SessionInfoMiddleware     → Parse JWT/cookie → inject *model.SessionInfo into ctx  (NEW)
  7. AuthzMiddleware           → Check collaborator of AuthgearConfig.AppID             (NEW)
  8. SecurityMiddleware        → Headers (XContentTypeOptions, XFrame, etc.)
  9. NoStore                   → Cache-Control: no-store
Handler (REST)
```

**Authorization logic:**

```
SessionInfoMiddleware
  ├── WebSDKSessionType == "refresh_token" → parse Authorization header JWT
  └── WebSDKSessionType == "cookie"        → parse x-authgear-session-info header

AuthzMiddleware
  ├── session.GetValidSessionInfo(ctx) == nil     → 401 Unauthenticated
  ├── GetCollaboratorByAppAndUser(appID, userID)
  │     ├── found                                  → proceed
  │     └── ErrCollaboratorNotFound                → 403 Forbidden
  └── other error                                  → 500
```

---

## Key Dependencies

| What | Where |
|---|---|
| `SessionInfoMiddleware` | `pkg/portal/session/middleware_session_info.go` |
| `session.GetValidSessionInfo` | `pkg/portal/session/context.go` |
| `CollaboratorService.GetCollaboratorByAppAndUser` | `pkg/portal/service/collaborator.go` |
| `ErrForbidden` / `ErrUnauthenticated` | `pkg/portal/service/authz.go` |
| `AuthgearConfig.AppID` | `pkg/portal/config/authgear.go` |
| `model.Collaborator` | `pkg/portal/model/` |

---

## Files to Create

### 1. `pkg/siteadmin/transport/middleware_authz.go`

```go
package transport

import (
    "context"
    "errors"
    "net/http"

    portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
    "github.com/authgear/authgear-server/pkg/portal/model"
    "github.com/authgear/authgear-server/pkg/portal/service"
    "github.com/authgear/authgear-server/pkg/portal/session"
)

type AuthzCollaboratorService interface {
    GetCollaboratorByAppAndUser(ctx context.Context, appID string, userID string) (*model.Collaborator, error)
}

type AuthzMiddleware struct {
    AuthgearConfig *portalconfig.AuthgearConfig
    Collaborators  AuthzCollaboratorService
}

func (m *AuthzMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        sessionInfo := session.GetValidSessionInfo(ctx)
        if sessionInfo == nil {
            writeError(w, r, service.ErrUnauthenticated)
            return
        }

        _, err := m.Collaborators.GetCollaboratorByAppAndUser(ctx, m.AuthgearConfig.AppID, sessionInfo.UserID)
        if errors.Is(err, service.ErrCollaboratorNotFound) {
            writeError(w, r, service.ErrForbidden)
            return
        } else if err != nil {
            writeError(w, r, err)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

## Files to Modify

### `pkg/siteadmin/transport/deps.go`

Add `AuthzMiddleware` to the dependency set:

```go
var DependencySet = wire.NewSet(
    wire.Struct(new(AppsListHandler), "*"),
    wire.Struct(new(AppGetHandler), "*"),
    wire.Struct(new(CollaboratorsListHandler), "*"),
    wire.Struct(new(CollaboratorAddHandler), "*"),
    wire.Struct(new(CollaboratorRemoveHandler), "*"),
    wire.Struct(new(MessagingUsageHandler), "*"),
    wire.Struct(new(MonthlyActiveUsersUsageHandler), "*"),
    wire.Struct(new(AuthzMiddleware), "*"),  // NEW
)
```

### `pkg/siteadmin/deps.go`

Add `session.DependencySet`, a partial `CollaboratorService` struct (only the 3 DB fields used by
`GetCollaboratorByAppAndUser`), and bind `AuthzCollaboratorService`:

```go
var DependencySet = wire.NewSet(
    deps.DependencySet,
    clock.DependencySet,
    globaldb.DependencySet,
    globalredis.DependencySet,
    session.DependencySet,                                                                        // NEW
    transport.DependencySet,
    wire.FieldsOf(new(*config.EnvironmentConfig), "CORSAllowedOrigins"),
    wire.Struct(new(CORSMatcher), "*"),
    wire.Bind(new(middleware.CORSOriginMatcher), new(*CORSMatcher)),
    wire.Struct(new(service.CollaboratorService), "SQLBuilder", "SQLExecutor", "GlobalDatabase"), // NEW: partial — only fields needed for GetCollaboratorByAppAndUser
    wire.Bind(new(transport.AuthzCollaboratorService), new(*service.CollaboratorService)),        // NEW
)
```

> **Why partial `CollaboratorService`?** Using `"SQLBuilder", "SQLExecutor", "GlobalDatabase"`
> avoids pulling in the full `service.DependencySet` dependency tree (smtp, template engine, admin
> API, etc.). `GetCollaboratorByAppAndUser` only touches those three fields.

### `pkg/siteadmin/wire.go`

Add injectors for both new middleware:

```go
func newSessionMiddleware(p *deps.RequestProvider) httproute.Middleware {
    panic(wire.Build(
        DependencySet,
        wire.Bind(new(httproute.Middleware), new(*session.SessionInfoMiddleware)),
    ))
}

func newAuthzMiddleware(p *deps.RequestProvider) httproute.Middleware {
    panic(wire.Build(
        DependencySet,
        wire.Bind(new(httproute.Middleware), new(*transport.AuthzMiddleware)),
    ))
}
```

### `pkg/siteadmin/wire_gen.go`

Run `wire gen ./pkg/siteadmin/...` after the deps changes. Do **not** hand-edit `wire_gen.go`.

### `pkg/siteadmin/routes.go` (Commit 3)

Add `SessionInfoMiddleware` to the chain only — no enforcement yet:

```go
apiChain := httproute.Chain(
    rootChain,
    p.Middleware(newSessionMiddleware),   // NEW: inject session info into ctx
    securityMiddleware,
    httproute.MiddlewareFunc(httputil.NoStore),
)
```

### `pkg/siteadmin/routes.go` (Commit 4)

Add `AuthzMiddleware` immediately after `SessionInfoMiddleware`:

```go
apiChain := httproute.Chain(
    rootChain,
    p.Middleware(newSessionMiddleware),   // injects session info
    p.Middleware(newAuthzMiddleware),     // NEW: enforce collaborator check
    securityMiddleware,
    httproute.MiddlewareFunc(httputil.NoStore),
)
```

---

## Implementation Roadmap: 5 Atomic Commits

### **Commit 1: Add SiteadminAuthgear config and separate RootProvider**

**Files Modified:**
- `cmd/portal/server/config.go` — add `SiteadminAuthgear portalconfig.AuthgearConfig` field and validate it:

```go
// SiteadminAuthgear configures Authgear for the Site Admin API.
// Allows the siteadmin server to authenticate against a different Authgear app than the portal.
SiteadminAuthgear portalconfig.AuthgearConfig `envconfig:"SITEADMIN_AUTHGEAR"`
```

`Validate()` is called inside `LoadConfigFromEnv()`, which has no knowledge of which servers are
being started. Serve-specific validation must be done separately in `Controller.Start()`, where
`c.ServePortal` and `c.ServeSiteadmin` are known.

Add a `LoadConfigOptions` struct and pass it to `LoadConfigFromEnv`, which continues to call
`Validate` internally — validation is guaranteed at the call site and callers cannot forget it:

```go
// config.go
type LoadConfigOptions struct {
    ServePortal    bool
    ServeSiteadmin bool
}

func LoadConfigFromEnv(opts LoadConfigOptions) (*Config, error) {
    config := &Config{}
    err := envconfig.Process("", config)
    if err != nil {
        return nil, fmt.Errorf("cannot load server config: %w", err)
    }
    err = config.Validate(opts)
    if err != nil {
        return nil, fmt.Errorf("invalid server config: %w", err)
    }
    return config, nil
}

func (c *Config) Validate(opts LoadConfigOptions) error {
    ctx := &validation.Context{}

    // always required
    if !ok { // CONFIG_SOURCE_TYPE check (unchanged)
        ctx.Child("CONFIG_SOURCE_TYPE").EmitErrorMessage(...)
    }
    if c.GlobalDatabase.DatabaseURL == "" {
        ctx.Child("DATABASE_URL").EmitErrorMessage("missing database URL")
    }

    if opts.ServePortal {
        if c.Authgear.ClientID == "" {
            ctx.Child("AUTHGEAR_CLIENT_ID").EmitErrorMessage("missing authgear client ID")
        }
        if c.Authgear.Endpoint == "" {
            ctx.Child("AUTHGEAR_ENDPOINT").EmitErrorMessage("missing authgear endpoint")
        }
    }

    if opts.ServeSiteadmin {
        if c.SiteadminAuthgear.AppID == "" {
            ctx.Child("SITEADMIN_AUTHGEAR_APP_ID").EmitErrorMessage("missing siteadmin authgear app ID")
        }
        if c.SiteadminAuthgear.Endpoint == "" {
            ctx.Child("SITEADMIN_AUTHGEAR_ENDPOINT").EmitErrorMessage("missing siteadmin authgear endpoint")
        }
    }

    return ctx.Error("invalid server configuration")
}
```

```go
// server.go — Controller.Start()
cfg, err := LoadConfigFromEnv(LoadConfigOptions{
    ServePortal:    c.ServePortal,
    ServeSiteadmin: c.ServeSiteadmin,
})
if err != nil {
    panic(fmt.Errorf("failed to load server config: %w", err))
}
```

`AppID` is required by `AuthzMiddleware` (collaborator check) and `Endpoint` is required by
`SessionInfoMiddleware` (JWK fetch). `ClientID` is not needed by siteadmin since it never
initiates an OAuth flow.

- `cmd/portal/server/server.go` — shallow-copy `p` and override `AuthgearConfig` for siteadmin:

```go
if c.ServeSiteadmin {
    // Shallow-copy the RootProvider so that the siteadmin server can use a
    // different AuthgearConfig (different AppID, Endpoint, etc.) without
    // affecting the portal server.
    //
    // A shallow copy is sufficient because:
    //   - Only AuthgearConfig needs to differ between portal and siteadmin.
    //   - All other fields (Database, RedisPool, ConfigSourceController, …)
    //     are pointers to shared infrastructure that both servers should reuse.
    //   - Dereferencing `p` copies the struct value, so overriding
    //     AuthgearConfig on the copy does not touch p.AuthgearConfig.
    siteadminProvider := *p
    siteadminProvider.AuthgearConfig = &cfg.SiteadminAuthgear
    specs = append(specs, server.NewSpec(ctx, &server.Spec{
        Name:          "authgear-portal-siteadmin",
        ListenAddress: cfg.SiteadminListenAddr,
        Handler:       siteadmin.NewRouter(&siteadminProvider),
    }))
    specs = append(specs, server.NewSpec(ctx, &server.Spec{
        Name:          "authgear-portal-siteadmin-internal",
        ListenAddress: cfg.SiteadminInternalListenAddr,
        Handler:       pprofutil.NewServeMux(),
    }))
}
```

New env vars exposed:
- `SITEADMIN_AUTHGEAR_CLIENT_ID`
- `SITEADMIN_AUTHGEAR_ENDPOINT`
- `SITEADMIN_AUTHGEAR_ENDPOINT_INTERNAL`
- `SITEADMIN_AUTHGEAR_APP_ID`
- `SITEADMIN_AUTHGEAR_WEB_SDK_SESSION_TYPE`

**Commit Message:** `"Allow siteadmin to use a separate AuthgearConfig"`

---

### **Commit 2: Add AuthzMiddleware**

**Files Created:**
- `pkg/siteadmin/transport/middleware_authz.go`

**Files Modified:**
- `pkg/siteadmin/transport/deps.go` — add `wire.Struct(new(AuthzMiddleware), "*")`

**Commit Message:** `"Add AuthzMiddleware for siteadmin collaborator check"`

---

### **Commit 3: Wire session and authz middleware**

**Files Modified:**
- `pkg/siteadmin/deps.go` — add `session.DependencySet`, partial `CollaboratorService` struct, bind `AuthzCollaboratorService`
- `pkg/siteadmin/wire.go` — add `newSessionMiddleware`, `newAuthzMiddleware`
- `pkg/siteadmin/wire_gen.go` — regenerated via `wire gen`

**Build Step:**
```bash
wire gen ./pkg/siteadmin/...
go build ./pkg/siteadmin/...
```

**Commit Message:** `"Wire session and authz middleware into siteadmin DI"`

---

### **Commit 4: Add SessionInfoMiddleware to route chain**

**Files Modified:**
- `pkg/siteadmin/routes.go` — insert `newSessionMiddleware` into `apiChain`

**Note:** No enforcement yet — session info is populated in context but not acted on.

**Commit Message:** `"Add SessionInfoMiddleware to siteadmin route chain"`

---

### **Commit 5: Enforce authorization**

**Files Modified:**
- `pkg/siteadmin/routes.go` — insert `newAuthzMiddleware` after `newSessionMiddleware`

**Commit Message:** `"Enforce authentication and authorization on siteadmin API routes"`

---

## Dependency Graph

```
Commit 1 (SiteadminAuthgear config + server.go override)
    ↓
Commit 2 (AuthzMiddleware struct + transport/deps.go)
    ↓
Commit 3 (wire both middleware)
    ↓
Commit 4 (routes: add SessionInfoMiddleware — no enforcement)
    ↓
Commit 5 (routes: add AuthzMiddleware — enforcement live)
```

---

## Verification

### Unauthenticated request → 401

```bash
curl -i http://localhost:3005/api/v1/apps
```
→ `401 Unauthorized` with `{"reason":"Unauthenticated",...}`

### Authenticated but not a collaborator → 403

```bash
curl -i -H "Authorization: Bearer <token-for-non-collaborator>" \
  http://localhost:3005/api/v1/apps
```
→ `403 Forbidden` with `{"reason":"Forbidden",...}`

### Authenticated collaborator → 200

```bash
curl -i -H "Authorization: Bearer <token-for-collaborator>" \
  http://localhost:3005/api/v1/apps
```
→ `200 OK` with apps list
