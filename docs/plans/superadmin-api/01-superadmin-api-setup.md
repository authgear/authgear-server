# Part 1: Superadmin API Setup - Package Creation & Integration

## Context

We want to add a **superadmin GraphQL API** to the portal server following the **portal API pattern** (not the authgear admin API pattern).

**Startup behavior:**
- Superadmin is **NOT** started by default with `./authgear-portal start`
- Superadmin is **ONLY** started when explicitly specified: `./authgear-portal start superadmin`
- When running `./authgear-portal start superadmin`, **ONLY superadmin server** starts (not the portal UI/API)
- This follows the same pattern as authgear: `./authgear start [main|resolver|admin]`

**Key design decisions:**
- **Reuse portal auth**: Uses `SessionInfoMiddleware` like the regular portal GraphQL API
  - Validates sessions via Authgear's OIDC/JWT endpoint (configured in `AuthgearConfig`)
  - Portal users already logged in can call the superadmin API
  - No custom auth config needed — inherited from portal's `AuthgearConfig`
- **OTel enabled**: Portal DOES use OpenTelemetry; superadmin follows the same pattern
- **Same middleware structure**: Mirrors portal's `graphqlChain` (otel → panic → body limit → sentry → session → security headers → content-type check)
- **Simpler than admin API**: No per-app context, no transaction wrapping, no custom auth layer, reuses portal's session validation
- **Optional startup**: Controlled via command-line flag (like authgear)

---

## Architecture Overview

**Command patterns:**
```
# Portal only (default behavior)
./authgear-portal start

# Superadmin only (new)
./authgear-portal start superadmin

# Superadmin with internal pprof
./authgear-portal start superadmin --internal
```

**Server startup by mode:**
```
Mode: portal (default)
  └── Controller.Start(ctx)
        ├── portal.NewRouter(p)           → :3003
        └── pprofutil.NewServeMux()       → :13003

Mode: superadmin (NEW)
  └── Controller.Start(ctx)
        ├── superadmin.NewRouter(p)       → :3005
        └── pprofutil.NewServeMux()       → :13005
```

**Provider hierarchy (reused from portal):**
```
deps.RootProvider  (process-scoped singleton)
  └─> deps.RequestProvider  (per-request, created inline by Middleware/Handler)
```

**Middleware chain (matches portal's graphqlChain):**
```
Request
  1. OtelMiddleware          → OpenTelemetry tracing
  2. PanicMiddleware         → Panic recovery
  3. BodyLimitMiddleware     → Max request body size
  4. SentryMiddleware        → Error capture
  5. SessionInfoMiddleware   → Validate session via Authgear
  6. SecurityMiddleware      → Headers (XContentTypeOptions, XFrame, etc.)
  7. NoStore                 → Cache-Control: no-store
  8. CheckContentType        → application/json or application/graphql
Handler (GraphQL)
```

---

## Command Line Startup Changes

**Before (portal only):**
```bash
./authgear-portal start          # Starts portal
```

**After (with optional superadmin):**
```bash
./authgear-portal start          # Default: starts portal (backward compatible)
./authgear-portal start portal   # Explicit: starts portal
./authgear-portal start superadmin  # Explicit: starts superadmin only (NEW)
```

**Implementation in `cmd/portal/cmd/cmdstart/start.go`** (mirrors `cmd/authgear/cmd/cmdstart/start.go`):
- Parse command args into known modes: `"portal"`, `"superadmin"`
- Default (no args): set `PortalMode = true`
- `portal` arg: set `PortalMode = true`
- `superadmin` arg: set `SuperadminMode = true` only

---

## Files to Create (13 files)

### 1. `pkg/portal/superadmin/routes.go`

```go
package superadmin

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/superadmin/transport"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider) http.Handler {
	router := httproute.NewRouter()
	router.Health(p.Handler(newHealthzHandler))

	securityMiddleware := httproute.Chain(
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.XFrameOptionsDeny),
		httproute.MiddlewareFunc(httputil.XRobotsTag),
		httproute.MiddlewareFunc(SuperadminCSPMiddleware),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
	)

	rootChain := httproute.Chain(
		p.Middleware(newOtelMiddleware),
		p.Middleware(newPanicMiddleware),
		p.Middleware(newBodyLimitMiddleware),
		p.Middleware(newSentryMiddleware),
	)

	graphqlChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionInfoMiddleware),
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoStore),
		httproute.MiddlewareFunc(httputil.CheckContentType([]string{
			graphqlutil.ContentTypeJSON,
			graphqlutil.ContentTypeGraphQL,
		})),
	)

	route := httproute.Route{Middleware: graphqlChain}
	router.AddRoutes(p.Handler(newGraphQLHandler), transport.ConfigureGraphQLRoute(route)...)

	return router.HTTPHandler()
}
```

### 2. `pkg/portal/superadmin/csp_middleware.go`

```go
package superadmin

import "net/http"

func SuperadminCSPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'none'; "+
			"script-src 'self' https: 'strict-dynamic'; "+
			"object-src 'none'; "+
			"base-uri 'none'; "+
			"frame-ancestors 'none';")
		next.ServeHTTP(w, r)
	})
}
```

### 3. `pkg/portal/superadmin/transport/graphql.go`

```go
package transport

import "github.com/authgear/authgear-server/pkg/util/httproute"

func ConfigureGraphQLRoute(route httproute.Route) []httproute.Route {
	route = route.WithMethods("GET", "POST")
	return []httproute.Route{
		route.WithPathPattern("/api/graphql"),
	}
}
```

### 4. `pkg/portal/superadmin/transport/handler_graphql.go`

```go
package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal/superadmin/graphql"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type GraphQLHandler struct {
	GraphQLContext *graphql.Context
}

func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		graphiql := &graphqlutil.GraphiQL{Title: "GraphiQL: Superadmin API - Authgear"}
		graphiql.ServeHTTP(w, r)
		return
	}
	q := r.URL.Query()
	q.Del("query")
	r.URL.RawQuery = q.Encode()

	graphqlHandler := &graphqlutil.Handler{Schema: graphql.Schema}
	ctx := graphql.WithContext(r.Context(), h.GraphQLContext)
	graphqlHandler.ContextHandler(ctx, w, r)
}
```

### 5. `pkg/portal/superadmin/transport/deps.go`

```go
package transport

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(GraphQLHandler), "*"),
)
```

### 6. `pkg/portal/superadmin/graphql/schema.go`

```go
package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var Schema *graphql.Schema

func init() {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:      query,
		Mutation:   mutation,
		Extensions: []graphql.Extension{&graphqlutil.APIErrorExtension{}},
	})
	if err != nil {
		panic(err)
	}
	Schema = &schema
}
```

### 7. `pkg/portal/superadmin/graphql/query.go`

```go
package graphql

import "github.com/graphql-go/graphql"

var query = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"__typename": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "Query", nil
			},
		},
	},
})
```

### 8. `pkg/portal/superadmin/graphql/mutation.go`

```go
package graphql

import "github.com/graphql-go/graphql"

var mutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"__typename": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "Mutation", nil
			},
		},
	},
})
```

### 9. `pkg/portal/superadmin/graphql/context.go`

```go
package graphql

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Context struct {
	Request *http.Request
	// Add service interfaces here as operations are added
}

func WithContext(ctx context.Context, gqlCtx *Context) context.Context {
	return graphqlutil.WithContext(ctx, gqlCtx)
}

func GQLContext(ctx context.Context) *Context {
	return graphqlutil.GQLContext(ctx).(*Context)
}
```

### 10. `pkg/portal/superadmin/graphql/deps.go`

```go
package graphql

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Context), "*"),
)
```

### 11. `pkg/portal/superadmin/deps.go`

```go
package superadmin

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/superadmin/graphql"
	"github.com/authgear/authgear-server/pkg/portal/superadmin/transport"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	graphql.DependencySet,
	transport.DependencySet,
)
```

### 12. `pkg/portal/superadmin/wire.go` (build tag: `wireinject`)

```go
//go:build wireinject
// +build wireinject

package superadmin

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/healthz"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func newPanicMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.PanicMiddleware)),
	))
}

func newBodyLimitMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		middleware.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*middleware.BodyLimitMiddleware)),
	))
}

func newOtelMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		deps.DependencySet,
		otelauthgear.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*otelauthgear.HTTPInstrumentationMiddleware)),
	))
}

func newSentryMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		deps.DependencySet,
		wire.Struct(new(middleware.SentryMiddleware), "*"),
		wire.Bind(new(httproute.Middleware), new(*middleware.SentryMiddleware)),
	))
}

func newSessionInfoMiddleware(p *deps.RequestProvider) httproute.Middleware {
	panic(wire.Build(
		DependencySet,
		session.DependencySet,
		wire.Bind(new(httproute.Middleware), new(*session.SessionInfoMiddleware)),
	))
}

func newHealthzHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		healthz.DependencySet,
		wire.Bind(new(http.Handler), new(*healthz.Handler)),
	))
}

func newGraphQLHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.GraphQLHandler)),
	))
}
```

### 13. `pkg/portal/superadmin/wire_gen.go`

Generated by `wire gen ./pkg/portal/superadmin/...` — do not edit manually.

---

## Files to Modify

### `cmd/portal/server/config.go`

Add two new listen address fields:

```go
PortalSuperadminListenAddr         string `envconfig:"PORTAL_SUPERADMIN_LISTEN_ADDR" default:"0.0.0.0:3005"`
PortalSuperadminInternalListenAddr string `envconfig:"PORTAL_SUPERADMIN_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13005"`
```

### `cmd/portal/server/server.go`

Add import:
```go
"github.com/authgear/authgear-server/pkg/portal/superadmin"
```

**Add Controller struct with mode selection** (similar to authgear's cmd/authgear/server/server.go):

```go
type Controller struct {
	PortalMode     bool
	SuperadminMode bool
}

func (c *Controller) Start(ctx context.Context) {
	logger := logger.GetLogger(ctx)

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		err = fmt.Errorf("failed to load server config: %w", err)
		panic(err)
	}

	ctx, p, err := deps.NewRootProvider(ctx, ...)
	if err != nil {
		err = fmt.Errorf("failed to setup server: %w", err)
		panic(err)
	}

	// ... existing config source setup ...

	var specs []signalutil.Daemon

	if c.PortalMode {
		// existing portal specs
		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-portal",
			ListenAddress: cfg.PortalListenAddr,
			Handler:       portal.NewRouter(p),
		}))
		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-portal-internal",
			ListenAddress: cfg.PortalInternalListenAddr,
			Handler:       pprofutil.NewServeMux(),
		}))
	}

	if c.SuperadminMode {
		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-portal-superadmin",
			ListenAddress: cfg.PortalSuperadminListenAddr,
			Handler:       superadmin.NewRouter(p),
		}))
		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-portal-superadmin-internal",
			ListenAddress: cfg.PortalSuperadminInternalListenAddr,
			Handler:       pprofutil.NewServeMux(),
		}))
	}

	signalutil.Start(ctx, specs...)
}
```

**Update start command handler** (in `cmd/portal/cmd/cmdstart/start.go`):
- Parse command args: `"portal"` or `"superadmin"`
- Set corresponding Controller flags
- If no args provided, default to `PortalMode = true`
- If `superadmin` arg provided, set `SuperadminMode = true` only

---

## Key Reused Utilities (do not re-implement)

| Utility | Import |
|---|---|
| `SessionInfoMiddleware` | `pkg/portal/session/` |
| `httproute.NewRouter`, `Chain`, `Route` | `pkg/util/httproute/` |
| `httputil.NoStore`, `XContentTypeOptionsNosniff`, etc. | `pkg/util/httputil/` |
| `graphqlutil.ContentTypeJSON/GraphQL`, `graphqlutil.Handler`, `graphqlutil.GraphiQL` | `pkg/util/graphqlutil/` |
| `middleware.PanicMiddleware`, `BodyLimitMiddleware`, `SentryMiddleware` | `pkg/lib/infra/middleware/` |
| `otelauthgear.HTTPInstrumentationMiddleware` | `pkg/lib/otelauthgear/` |
| `healthz.Handler`, `healthz.DependencySet` | `pkg/lib/healthz/` |
| `deps.RootProvider`, `deps.RequestProvider` | `pkg/portal/deps/` |

---

## Implementation Roadmap: 9 Atomic Commits

Each commit is independently reviewable and builds on previous ones with zero mixing of concerns.

### **Commit 1: Add superadmin config fields**
**Files Modified:** `cmd/portal/server/config.go`
```go
PortalSuperadminListenAddr         string `envconfig:"PORTAL_SUPERADMIN_LISTEN_ADDR" default:"0.0.0.0:3005"`
PortalSuperadminInternalListenAddr string `envconfig:"PORTAL_SUPERADMIN_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13005"`
```
**Commit Message:** "Add superadmin API listen address config"

---

### **Commit 2: Create superadmin GraphQL schema and types**
**Files Created:**
- `pkg/portal/superadmin/graphql/schema.go`
- `pkg/portal/superadmin/graphql/query.go`
- `pkg/portal/superadmin/graphql/mutation.go`
- `pkg/portal/superadmin/graphql/context.go`
- `pkg/portal/superadmin/graphql/deps.go`

**Scope:** GraphQL domain layer only
**Commit Message:** "Create superadmin GraphQL schema and context"

---

### **Commit 3: Create superadmin transport handlers**
**Files Created:**
- `pkg/portal/superadmin/transport/graphql.go`
- `pkg/portal/superadmin/transport/handler_graphql.go`
- `pkg/portal/superadmin/transport/deps.go`

**Scope:** HTTP transport layer only
**Commit Message:** "Create superadmin GraphQL HTTP handlers"

---

### **Commit 4: Create superadmin CSP middleware**
**Files Created:**
- `pkg/portal/superadmin/csp_middleware.go`

**Scope:** Security middleware only
**Commit Message:** "Add Content-Security-Policy middleware to superadmin"

---

### **Commit 5: Create superadmin dependency sets**
**Files Created:**
- `pkg/portal/superadmin/deps.go`

**Scope:** Dependency injection composition (combines graphql + transport + portal deps)
**Commit Message:** "Create superadmin dependency sets"

---

### **Commit 6: Create superadmin wire injectors**
**Files Created:**
- `pkg/portal/superadmin/wire.go` (build tag: `wireinject`)

**Scope:** Wire injector functions (middleware and handler factories)
**Commit Message:** "Add superadmin wire injectors for DI"

---

### **Commit 7: Create superadmin router**
**Files Created:**
- `pkg/portal/superadmin/routes.go`

**Scope:** HTTP routing and middleware chain assembly
**Dependencies:** Requires wire.go to be in place (wire functions are called here)
**Commit Message:** "Create superadmin router with middleware chain"

---

### **Commit 8: Generate superadmin wire code**
**Build Step:**
```bash
wire gen ./pkg/portal/superadmin/...
```

**Files Generated:**
- `pkg/portal/superadmin/wire_gen.go` (auto-generated, do not edit)

**Scope:** Wire code generation (no manual edits)
**Prerequisites:** Must run after Commit 6 (wire.go must exist)
**Post-step:** `go build ./cmd/portal/...` should pass
**Commit Message:** "Generate superadmin wire dependency injection code"

---

### **Commit 9: Integrate superadmin router into portal server**
**Files Modified:**
- `cmd/portal/server/config.go` (in `Controller` struct - likely already modified in Commit 1, add Mode flags if not present)
- `cmd/portal/server/server.go`
- `cmd/portal/cmd/cmdstart/start.go` (new file or modify existing)

**Changes to `server.go`:**
- Add import: `"github.com/authgear/authgear-server/pkg/portal/superadmin"`
- Add `Controller` struct with `PortalMode` and `SuperadminMode` flags
- Update `Start()` method to use conditional server specs based on flags

**Changes to `cmdstart/start.go`:**
- Parse command args: `"portal"` or `"superadmin"`
- Set corresponding Controller flags
- Default (no args): set `PortalMode = true`
- `superadmin` arg: set `SuperadminMode = true` only

**Scope:** Server startup integration only
**Dependencies:** All superadmin package commits must be complete
**Commit Message:** "Wire superadmin API into portal server startup"

---

## Dependency Graph

```
Commit 1 (config)
    ↓
Commit 2 (graphql)  ────────┐
    ↓                        │
Commit 3 (transport) ────────┤
    ↓                        │
Commit 4 (csp)              │
    ↓                        ↓
Commit 5 (deps) ────────────
    ↓
Commit 6 (wire)
    ↓
Commit 7 (routes)
    ↓
Commit 8 (wire gen)
    ↓
Commit 9 (server integration)
```

**Key Properties:**
- ✅ Each commit is **independently reviewable**
- ✅ Each commit adds **one logical component**
- ✅ No mixing of concerns (config → schema → transport → middleware → wiring → routing → generation → integration)
- ✅ Build succeeds after Commit 8
- ✅ Server fully functional after Commit 9

---

## Verification

### Build & Setup

1. Run wire generation:
   ```bash
   wire gen ./pkg/portal/superadmin/...
   ```

2. Build:
   ```bash
   go build ./cmd/portal/...
   ```

### Test Portal Mode (Default)

3. Start portal (default behavior):
   ```bash
   ./authgear-portal start
   ```
   → Portal starts on `:3003`, internal on `:13003`
   → Superadmin does NOT start

4. Verify portal is running:
   ```bash
   curl http://localhost:3003/healthz
   ```
   → Returns `200 OK`

5. Verify superadmin is NOT running:
   ```bash
   curl http://localhost:3005/healthz
   ```
   → Connection refused (superadmin not started)

### Test Superadmin Mode (Explicit)

6. Start superadmin only:
   ```bash
   ./authgear-portal start superadmin
   ```
   → Superadmin starts on `:3005`, internal on `:13005`
   → Portal UI/API does NOT start

7. Test superadmin health check:
   ```bash
   curl http://localhost:3005/healthz
   ```
   → Returns `200 OK`

8. Verify portal is NOT running:
   ```bash
   curl http://localhost:3003/healthz
   ```
   → Connection refused (portal not started)

9. Test GraphiQL UI: Open `http://localhost:3005/api/graphql` in browser

10. Test GraphQL POST:
    ```bash
    curl -X POST http://localhost:3005/api/graphql \
      -H 'Content-Type: application/json' \
      -d '{"query":"{ __typename }"}'
    ```
    → Returns GraphQL response

11. Test superadmin pprof internal server:
    ```bash
    curl http://localhost:13005/debug/pprof/
    ```
    → Responds
