# Part 1: Site Admin API Setup - Package Creation & Integration

## Context

We want to add a **Site Admin RESTful API** to the portal server following the **portal API pattern** (not the authgear admin API pattern).

**Startup behavior:**
- Site Admin is **NOT** started by default with `./authgear-portal start`
- Site Admin is **ONLY** started when explicitly specified: `./authgear-portal start siteadmin`
- When running `./authgear-portal start siteadmin`, **ONLY Site Admin server** starts (not the portal UI/API)
- This follows the same pattern as authgear: `./authgear start [main|resolver|admin]`

**Key design decisions:**
- **Authorization**: Will be handled later (not part of this setup)
- **OTel enabled**: Portal DOES use OpenTelemetry; Site Admin follows the same pattern
- **Same middleware structure**: Mirrors portal's middleware chain (otel → panic → body limit → sentry → security headers)
- **RESTful API**: Simple JSON-based REST endpoints instead of GraphQL
- **Simpler than admin API**: No per-app context, no transaction wrapping
- **Optional startup**: Controlled via command-line flag (like authgear)

---

## Architecture Overview

**Command patterns:**
```
# Portal only (default behavior)
./authgear-portal start

# Site Admin only (new)
./authgear-portal start siteadmin

# Site Admin with internal pprof
./authgear-portal start siteadmin --internal
```

**Server startup by mode:**
```
Mode: portal (default)
  └── Controller.Start(ctx)
        ├── portal.NewRouter(p)           → :3003
        └── pprofutil.NewServeMux()       → :13003

Mode: siteadmin (NEW)
  └── Controller.Start(ctx)
        ├── siteadmin.NewRouter(p)       → :3005
        └── pprofutil.NewServeMux()       → :13005
```

**Provider hierarchy (reused from portal):**
```
deps.RootProvider  (process-scoped singleton)
  └─> deps.RequestProvider  (per-request, created inline by Middleware/Handler)
```

**Middleware chain:**
```
Request
  1. OtelMiddleware          → OpenTelemetry tracing
  2. PanicMiddleware         → Panic recovery
  3. BodyLimitMiddleware     → Max request body size
  4. SentryMiddleware        → Error capture
  5. (Authorization)         → TODO: Will be handled later
  6. SecurityMiddleware      → Headers (XContentTypeOptions, XFrame, etc.)
  7. NoStore                 → Cache-Control: no-store
Handler (REST)
```

---

## Command Line Startup Changes

**Before (portal only):**
```bash
./authgear-portal start          # Starts portal
```

**After (with optional siteadmin):**
```bash
./authgear-portal start          # Default: starts portal (backward compatible)
./authgear-portal start portal   # Explicit: starts portal
./authgear-portal start siteadmin  # Explicit: starts siteadmin only (NEW)
```

**Implementation in `cmd/portal/cmd/cmdstart/start.go`** (mirrors `cmd/authgear/cmd/cmdstart/start.go`):
- Parse command args into known modes: `"portal"`, `"siteadmin"`
- Default (no args): set `PortalMode = true`
- `portal` arg: set `PortalMode = true`
- `siteadmin` arg: set `SiteadminMode = true` only

---

## Files to Create (7 files)

### 1. `pkg/siteadmin/routes.go`

```go
package siteadmin

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/siteadmin/transport"
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
		httproute.MiddlewareFunc(portal.PortalCSPMiddleware),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
	)

	rootChain := httproute.Chain(
		p.Middleware(newOtelMiddleware),
		p.Middleware(newPanicMiddleware),
		p.Middleware(newBodyLimitMiddleware),
		p.Middleware(newSentryMiddleware),
	)

	apiChain := httproute.Chain(
		rootChain,
		// TODO: Authorization will be handled later
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoStore),
	)

	route := httproute.Route{Middleware: apiChain}
	router.Add(transport.ConfigureProjectsListRoute(route), p.Handler(newProjectsListHandler))

	return router.HTTPHandler()
}
```

### 2. `pkg/siteadmin/transport/projects_list.go`

```go
package transport

import "github.com/authgear/authgear-server/pkg/util/httproute"

func ConfigureProjectsListRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").
		WithPathPattern("/api/v1/projects")
}
```

### 3. `pkg/siteadmin/transport/handler_projects_list.go`

```go
package transport

import (
	"encoding/json"
	"net/http"
)

type ProjectsListHandler struct {
	// Add service dependencies here as needed
}

type ProjectsListResponse struct {
	Projects []interface{} `json:"projects"`
}

func (h *ProjectsListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Scaffolding: return empty list for now
	response := ProjectsListResponse{
		Projects: []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
```

### 4. `pkg/siteadmin/transport/deps.go`

```go
package transport

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(ProjectsListHandler), "*"),
)
```

### 5. `pkg/siteadmin/deps.go`

```go
package siteadmin

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/siteadmin/transport"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	transport.DependencySet,
)
```

### 6. `pkg/siteadmin/wire.go` (build tag: `wireinject`)

```go
//go:build wireinject
// +build wireinject

package siteadmin

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/healthz"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/siteadmin/transport"
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

func newHealthzHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		healthz.DependencySet,
		wire.Bind(new(http.Handler), new(*healthz.Handler)),
	))
}

func newProjectsListHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		DependencySet,
		wire.Bind(new(http.Handler), new(*transport.ProjectsListHandler)),
	))
}
```

### 7. `pkg/siteadmin/wire_gen.go`

Generated by `wire gen ./pkg/siteadmin/...` — do not edit manually.

---

## Files to Modify

### `cmd/portal/server/config.go`

Add two new listen address fields:

```go
PortalSiteadminListenAddr         string `envconfig:"PORTAL_SITEADMIN_LISTEN_ADDR" default:"0.0.0.0:3005"`
PortalSiteadminInternalListenAddr string `envconfig:"PORTAL_SITEADMIN_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13005"`
```

### `cmd/portal/server/server.go`

Add import:
```go
"github.com/authgear/authgear-server/pkg/siteadmin"
```

**Add Controller struct with mode selection** (similar to authgear's cmd/authgear/server/server.go):

```go
type Controller struct {
	PortalMode     bool
	SiteadminMode bool
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

	if c.SiteadminMode {
		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-portal-siteadmin",
			ListenAddress: cfg.PortalSiteadminListenAddr,
			Handler:       siteadmin.NewRouter(p),
		}))
		specs = append(specs, server.NewSpec(ctx, &server.Spec{
			Name:          "authgear-portal-siteadmin-internal",
			ListenAddress: cfg.PortalSiteadminInternalListenAddr,
			Handler:       pprofutil.NewServeMux(),
		}))
	}

	signalutil.Start(ctx, specs...)
}
```

**Update start command handler** (in `cmd/portal/cmd/cmdstart/start.go`):
- Parse command args: `"portal"` or `"siteadmin"`
- Set corresponding Controller flags
- If no args provided, default to `PortalMode = true`
- If `siteadmin` arg provided, set `SiteadminMode = true` only

---

## Key Reused Utilities (do not re-implement)

| Utility | Import |
|---|---|
| `PortalCSPMiddleware` | `pkg/portal/` |
| `httproute.NewRouter`, `Chain`, `Route` | `pkg/util/httproute/` |
| `httputil.NoStore`, `XContentTypeOptionsNosniff`, etc. | `pkg/util/httputil/` |
| `middleware.PanicMiddleware`, `BodyLimitMiddleware`, `SentryMiddleware` | `pkg/lib/infra/middleware/` |
| `otelauthgear.HTTPInstrumentationMiddleware` | `pkg/lib/otelauthgear/` |
| `healthz.Handler`, `healthz.DependencySet` | `pkg/lib/healthz/` |
| `deps.RootProvider`, `deps.RequestProvider` | `pkg/portal/deps/` |

---

## Implementation Roadmap: 7 Atomic Commits

Each commit is independently reviewable and builds on previous ones with zero mixing of concerns.

### **Commit 1: Add Site Admin config fields**
**Files Modified:** `cmd/portal/server/config.go`
```go
PortalSiteadminListenAddr         string `envconfig:"PORTAL_SITEADMIN_LISTEN_ADDR" default:"0.0.0.0:3005"`
PortalSiteadminInternalListenAddr string `envconfig:"PORTAL_SITEADMIN_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13005"`
```
**Commit Message:** "Add Site Admin API listen address config"

---

### **Commit 2: Create Site Admin transport handlers**
**Files Created:**
- `pkg/siteadmin/transport/projects_list.go`
- `pkg/siteadmin/transport/handler_projects_list.go`
- `pkg/siteadmin/transport/deps.go`

**Scope:** HTTP transport layer with scaffolding REST handler
**Commit Message:** "Create Site Admin REST transport handlers"

---

### **Commit 3: Create Site Admin dependency sets**
**Files Created:**
- `pkg/siteadmin/deps.go`

**Scope:** Dependency injection composition (combines transport + portal deps)
**Commit Message:** "Create Site Admin dependency sets"

---

### **Commit 4: Create Site Admin wire injectors**
**Files Created:**
- `pkg/siteadmin/wire.go` (build tag: `wireinject`)

**Scope:** Wire injector functions (middleware and handler factories)
**Commit Message:** "Add Site Admin wire injectors for DI"

---

### **Commit 5: Create Site Admin router and generate wire code**
**Files Created:**
- `pkg/siteadmin/routes.go`

**Build Step:**
```bash
wire gen ./pkg/siteadmin/...
```

**Files Generated:**
- `pkg/siteadmin/wire_gen.go` (auto-generated, do not edit)

**Scope:** HTTP routing, middleware chain assembly, and wire code generation
**Post-step:** `go build ./cmd/portal/...` should pass
**Commit Message:** "Create Site Admin router with middleware chain"

---

### **Commit 6: Integrate Site Admin router into portal server**
**Files Modified:**
- `cmd/portal/server/server.go`
- `cmd/portal/cmd/cmdstart/start.go` (new file or modify existing)

**Changes to `server.go`:**
- Add import: `"github.com/authgear/authgear-server/pkg/siteadmin"`
- Add `Controller` struct with `PortalMode` and `SiteadminMode` flags
- Update `Start()` method to use conditional server specs based on flags

**Changes to `cmdstart/start.go`:**
- Parse command args: `"portal"` or `"siteadmin"`
- Set corresponding Controller flags
- Default (no args): set `PortalMode = true`
- `siteadmin` arg: set `SiteadminMode = true` only

**Scope:** Server startup integration only
**Dependencies:** All siteadmin package commits must be complete
**Commit Message:** "Wire Site Admin API into portal server startup"

---

### **Commit 7: Update local development setup**
**Files Modified:**
- `nginx.conf` - Add proxy configuration for Site Admin API on port 8100 (no auth proxy)
- `docker-compose.yaml` - Add port 8100 mapping to proxy service
- `Makefile` - Update portal targets to also start Site Admin server

**Changes to `nginx.conf`:**
- Add server block listening on port 8100 to proxy Site Admin requests
- No auth_request needed (authorization will be handled later)

**Changes to `docker-compose.yaml`:**
- Add port mapping `8100:8100` to proxy service

**Changes to `Makefile`:**
- Update `start-portal` target to run both portal and siteadmin servers

**Scope:** Local development configuration only
**Dependencies:** Server integration must be complete (Commit 6)
**Commit Message:** "Update local development setup for Site Admin API"

---

## Dependency Graph

```
Commit 1 (config)
    ↓
Commit 2 (transport)
    ↓
Commit 3 (deps)
    ↓
Commit 4 (wire)
    ↓
Commit 5 (routes + wire gen)
    ↓
Commit 6 (server integration)
    ↓
Commit 7 (local dev setup)
```

**Key Properties:**
- ✅ Each commit is **independently reviewable**
- ✅ Each commit adds **one logical component**
- ✅ No mixing of concerns (config → transport → deps → wiring → routing → integration → dev setup)
- ✅ Build succeeds after Commit 5
- ✅ Server fully functional after Commit 6
- ✅ Local development ready after Commit 7

---

## Verification

### Build & Setup

1. Run wire generation:
   ```bash
   wire gen ./pkg/siteadmin/...
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
   → Site Admin does NOT start

4. Verify portal is running:
   ```bash
   curl http://localhost:3003/healthz
   ```
   → Returns `200 OK`

5. Verify Site Admin is NOT running:
   ```bash
   curl http://localhost:3005/healthz
   ```
   → Connection refused (Site Admin not started)

### Test Site Admin Mode (through nginx)

6. Start portal and Site Admin:
   ```bash
   make start-portal
   ```
   → Portal starts on `:3003`, Site Admin starts on `:3005`
   → nginx proxies Site Admin on port `:8100`

7. Test Site Admin health check (direct):
   ```bash
   curl http://localhost:3005/healthz
   ```
   → Returns `200 OK`

8. Test REST API scaffolding endpoint through nginx (port 8100):
   ```bash
   curl http://localhost:8100/api/v1/projects
   ```
   → Returns `{"projects":[]}`

9. Test Site Admin pprof internal server (direct, not through nginx):
   ```bash
   curl http://localhost:13005/debug/pprof/
   ```
   → Responds

### Test Site Admin Standalone Mode (Optional)

10. Start Site Admin only (without portal):
    ```bash
    ./authgear-portal start siteadmin
    ```
    → Site Admin starts on `:3005`, internal on `:13005`
    → Portal does NOT start

11. Verify portal is NOT running:
    ```bash
    curl http://localhost:3003/healthz
    ```
    → Connection refused (portal not started)
