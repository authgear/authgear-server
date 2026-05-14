# Account Lockout – Admin API Implementation Plan

## 1. Goal / Scope

Add two capabilities to the Admin GraphQL API:

1. **`User.accountLockout` query field** — returns the current lockout state of a user.
2. **`resetAccountLockout` mutation** — clears all lockout state for a user.

Spec: `docs/specs/account-lockout.md` §Admin API.

---

## 2. Redis Data Model (existing — read-only reference)

Key derivation: `redisRecordKey(appID, spec)` → `fmt.Sprintf("app:%s:lockout:%s", appID, spec.Key())` where `spec.Key()` = `"AccountAuthentication:{userID}"`.

| Key | Type | TTL | Contents |
|---|---|---|---|
| `app:{appID}:lockout:AccountAuthentication:{userID}` (the "hash") | Redis Hash | `EXPIREAT now + history_duration` (reset on every attempt) | Field `total` = global count; field `{ip}` = per-IP attempt count |
| `app:{appID}:lockout:AccountAuthentication:{userID}:lock:global` | String | `EXPIREAT locked_until_epoch` | Unix epoch of `lockedUntil` (for `per_user` lockout) |
| `app:{appID}:lockout:AccountAuthentication:{userID}:lock:{ip}` | String | `EXPIREAT locked_until_epoch` | Unix epoch of `lockedUntil` for that IP (for `per_user_per_ip` lockout) |

### How IP addresses are discovered for `per_user_per_ip`

IP addresses are stored as **field names** in the attempt hash (`HSET record_key, contributor, contributor_total` where `contributor` = IP). `HGetAll` on the hash enumerates all IPs. The lock key for each IP is derived as `{record_key}:lock:{ip}`.

**Known gap:** the hash TTL is `now + history_duration`, while each lock key TTL is at most `locked_until_epoch` (`maximum_duration` in the future). If `history_duration < maximum_duration`, the hash can expire while lock keys are still alive. Both `GetStatus` and `ClearAll` have this gap since they use `HGetAll` to discover IPs. Affects only configurations where `history_duration < maximum_duration` — not the intended default.

---

## 3. New Types

### New file: `pkg/api/model/lockout.go`

`AccountLockoutType` is defined here as the single source of truth. `pkg/lib/config` already imports `pkg/api/model`, so it can alias its own type to this one without a cycle.

```go
package model

import "time"

type AccountLockoutType string

const (
    AccountLockoutTypePerUser      AccountLockoutType = "per_user"
    AccountLockoutTypePerUserPerIP AccountLockoutType = "per_user_per_ip"
)

type LockedIP struct {
    IPAddress   string    `json:"ip_address"`
    LockedUntil time.Time `json:"locked_until"`
}

// AccountLockoutStatus is the admin-facing lockout state of a user.
// LockoutType is derived from config. LockedIPs is sorted by LockedUntil descending.
type AccountLockoutStatus struct {
    LockoutType AccountLockoutType `json:"lockout_type"`
    IsLocked    bool               `json:"is_locked"`
    LockedUntil *time.Time         `json:"locked_until,omitempty"` // non-nil only for per_user
    LockedIPs   []LockedIP         `json:"locked_ips"`             // non-empty only for per_user_per_ip, sorted LockedUntil desc
}
```

### `pkg/lib/config/authentication_lockout.go`

Replace the `AuthenticationLockoutType` type definition and its constants with a type alias pointing to `model.AccountLockoutType`:

```go
// Before:
type AuthenticationLockoutType string

const (
    AuthenticationLockoutTypePerUser      AuthenticationLockoutType = "per_user"
    AuthenticationLockoutTypePerUserPerIP AuthenticationLockoutType = "per_user_per_ip"
)

// After:
type AuthenticationLockoutType = model.AccountLockoutType

const (
    AuthenticationLockoutTypePerUser      = model.AccountLockoutTypePerUser
    AuthenticationLockoutTypePerUserPerIP = model.AccountLockoutTypePerUserPerIP
)
```

All existing call sites that use `config.AuthenticationLockoutType` continue to compile unchanged. The type alias makes `config.AuthenticationLockoutType` and `model.AccountLockoutType` identical, so no cast is needed in the facade.

### `pkg/lib/lockout/models.go`

Add below the existing `MakeAttemptResult`. `LockoutStatus` is the raw storage-layer result and uses `model.LockedIP` — `pkg/lib/lockout` can import `pkg/api/model` with no cycle.

```go
import apimodel "github.com/authgear/authgear-server/pkg/api/model"

// LockoutStatus is the raw per-user status returned by Storage.GetStatus.
// For per_user: IsLocked and LockedUntil are populated; LockedIPs is nil.
// For per_user_per_ip: IsLocked and LockedIPs are populated; LockedUntil is nil.
type LockoutStatus struct {
    IsLocked    bool
    LockedUntil *time.Time
    LockedIPs   []apimodel.LockedIP
}
```

---

## 4. Storage Layer

### `pkg/lib/lockout/storage.go`

Add two methods to the `Storage` interface:

```go
GetStatus(ctx context.Context, spec LockoutSpec) (*LockoutStatus, error)
ClearAll(ctx context.Context, spec LockoutSpec) error
```

### `pkg/lib/lockout/record.go` — new Lua scripts and Go wrappers

Add two Lua scripts and two Go wrapper functions at the bottom of the file.

#### `getStatusLuaScript`

Declared as `goredis.NewScript(constants + ...)` to reuse the existing `GLOBAL_TOTAL_KEY` constant.

```lua
-- KEYS[1]: record_key
-- ARGV[1]: is_global ("1" for per_user, "0" for per_user_per_ip)
--
-- Returns for per_user (is_global=1):
--   {0}                    -- not locked
--   {1, locked_until_epoch} -- locked
--
-- Returns for per_user_per_ip (is_global=0):
--   First element: is_any_locked (0 or 1)
--   Followed by pairs: ip_string, locked_until_epoch
--   Example: {1, "1.2.3.4", 1234567890, "5.6.7.8", 1234567999}
redis.replicate_commands()
local record_key = KEYS[1]
local is_global = ARGV[1] == "1"
local now_raw = redis.call("TIME")
local now = tonumber(now_raw[1])

if is_global then
    local lock_key = record_key .. ":lock:global"
    local v = redis.pcall("GET", lock_key)
    if v and not v["err"] and type(v) == "string" then
        local epoch = tonumber(v)
        if epoch and epoch > now then
            return {1, epoch}
        end
    end
    return {0}
else
    local result = {}
    local is_any_locked = 0
    local hash_data = redis.pcall("HGETALL", record_key)
    if hash_data and not hash_data["err"] then
        for i = 1, #hash_data, 2 do
            local field = hash_data[i]
            if field ~= GLOBAL_TOTAL_KEY then
                local lock_key = record_key .. ":lock:" .. field
                local v = redis.pcall("GET", lock_key)
                if v and not v["err"] and type(v) == "string" then
                    local epoch = tonumber(v)
                    if epoch and epoch > now then
                        is_any_locked = 1
                        table.insert(result, field)
                        table.insert(result, epoch)
                    end
                end
            end
        end
    end
    table.insert(result, 1, is_any_locked)
    return result
end
```

#### `clearAllLuaScript`

Declared as `goredis.NewScript(constants + ...)` to reuse the existing `GLOBAL_TOTAL_KEY` constant.

```lua
-- KEYS[1]: record_key
-- ARGV[1]: is_global ("1" or "0")
-- Returns: 1
redis.replicate_commands()
local record_key = KEYS[1]
local is_global = ARGV[1] == "1"

if is_global then
    redis.call("DEL", record_key, record_key .. ":lock:global")
else
    local hash_data = redis.pcall("HGETALL", record_key)
    if hash_data and not hash_data["err"] then
        local keys_to_del = {record_key}
        for i = 1, #hash_data, 2 do
            local field = hash_data[i]
            if field ~= GLOBAL_TOTAL_KEY then
                table.insert(keys_to_del, record_key .. ":lock:" .. field)
            end
        end
        redis.call("DEL", unpack(keys_to_del))
    else
        redis.call("DEL", record_key)
    end
end
return 1
```

#### Go wrapper functions

// record.go must import apimodel "github.com/authgear/authgear-server/pkg/api/model"
// for the getStatus wrapper below.

func getStatus(
    ctx context.Context, conn redis.Redis_6_0_Cmdable,
    key string,
    isGlobal bool,
) (*LockoutStatus, error) {
    isGlobalStr := "0"
    if isGlobal {
        isGlobalStr = "1"
    }
    result, err := getStatusLuaScript.Run(ctx, conn, []string{key}, isGlobalStr).Slice()
    if err != nil {
        return nil, err
    }

    if isGlobal {
        // {0} or {1, epoch}
        isLocked := result[0].(int64) == 1
        status := &LockoutStatus{IsLocked: isLocked}
        if isLocked && len(result) > 1 {
            t := time.Unix(result[1].(int64), 0).UTC()
            status.LockedUntil = &t
        }
        return status, nil
    }

    // {is_any_locked, ip1, epoch1, ip2, epoch2, ...}
    isLocked := result[0].(int64) == 1
    var lockedIPs []apimodel.LockedIP
    for i := 1; i+1 < len(result); i += 2 {
        ip := result[i].(string)
        t := time.Unix(result[i+1].(int64), 0).UTC()
        lockedIPs = append(lockedIPs, apimodel.LockedIP{IPAddress: ip, LockedUntil: t})
    }
    return &LockoutStatus{IsLocked: isLocked, LockedIPs: lockedIPs}, nil
}

func clearAll(
    ctx context.Context, conn redis.Redis_6_0_Cmdable,
    key string,
    isGlobal bool,
) error {
    isGlobalStr := "0"
    if isGlobal {
        isGlobalStr = "1"
    }
    _, err := clearAllLuaScript.Run(ctx, conn, []string{key}, isGlobalStr).Bool()
    return err
}
```

### `pkg/lib/lockout/storage_redis.go`

Add `GetStatus` and `ClearAll` to `StorageRedis`:

```go
func (s StorageRedis) GetStatus(ctx context.Context, spec LockoutSpec) (status *LockoutStatus, err error) {
    err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
        status, err = getStatus(ctx, conn, redisRecordKey(s.AppID, spec), spec.IsGlobal)
        return err
    })
    return status, err
}

func (s StorageRedis) ClearAll(ctx context.Context, spec LockoutSpec) (err error) {
    err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
        return clearAll(ctx, conn, redisRecordKey(s.AppID, spec), spec.IsGlobal)
    })
    return err
}
```

---

## 5. Service Layer

### `pkg/lib/lockout/service.go`

Add two methods to `Service`:

```go
func (s *Service) GetStatus(ctx context.Context, spec LockoutSpec) (*LockoutStatus, error) {
    if !spec.Enabled {
        return &LockoutStatus{IsLocked: false}, nil
    }
    return s.Storage.GetStatus(ctx, spec)
}

func (s *Service) ClearAll(ctx context.Context, spec LockoutSpec) error {
    if !spec.Enabled {
        return nil
    }
    return s.Storage.ClearAll(ctx, spec)
}
```

---

## 6. Admin Facade

### New file: `pkg/admin/facade/lockout.go`

```go
package facade

import (
    "context"
    "sort"

    apimodel "github.com/authgear/authgear-server/pkg/api/model"
    "github.com/authgear/authgear-server/pkg/lib/config"
    lockoutpkg "github.com/authgear/authgear-server/pkg/lib/lockout"
)

type LockoutProvider interface {
    GetStatus(ctx context.Context, spec lockoutpkg.LockoutSpec) (*lockoutpkg.LockoutStatus, error)
    ClearAll(ctx context.Context, spec lockoutpkg.LockoutSpec) error
}

type LockoutFacade struct {
    LockoutConfig *config.AuthenticationLockoutConfig
    Lockout       LockoutProvider
}

func (f *LockoutFacade) GetAccountLockoutStatus(ctx context.Context, userID string) (*apimodel.AccountLockoutStatus, error) {
    // NewAccountAuthenticationSpecForCheck returns a disabled spec when IsEnabled() is false;
    // Service.GetStatus returns {IsLocked: false} for a disabled spec without hitting Redis.
    spec := lockoutpkg.NewAccountAuthenticationSpecForCheck(f.LockoutConfig, userID)
    status, err := f.Lockout.GetStatus(ctx, spec)
    if err != nil {
        return nil, err
    }
    sort.Slice(status.LockedIPs, func(i, j int) bool {
        return status.LockedIPs[i].LockedUntil.After(status.LockedIPs[j].LockedUntil)
    })
    return &apimodel.AccountLockoutStatus{
        LockoutType: f.LockoutConfig.LockoutType,
        IsLocked:    status.IsLocked,
        LockedUntil: status.LockedUntil,
        LockedIPs:   status.LockedIPs,
    }, nil
}

func (f *LockoutFacade) ResetAccountLockout(ctx context.Context, userID string) error {
    // NewAccountAuthenticationSpecForCheck returns a disabled spec when IsEnabled() is false;
    // Service.ClearAll is a no-op for a disabled spec.
    spec := lockoutpkg.NewAccountAuthenticationSpecForCheck(f.LockoutConfig, userID)
    return f.Lockout.ClearAll(ctx, spec)
}
```

### `pkg/admin/facade/deps.go`

Add `wire.Struct(new(LockoutFacade), "*")` to `DependencySet`.

---

## 7. Admin GraphQL Layer

### `pkg/admin/graphql/context.go`

Add import `apimodel "github.com/authgear/authgear-server/pkg/api/model"`.

Add interface and field to `Context`:

```go
type AccountLockoutFacade interface {
    GetAccountLockoutStatus(ctx context.Context, userID string) (*apimodel.AccountLockoutStatus, error)
    ResetAccountLockout(ctx context.Context, userID string) error
}

// In Context struct:
AccountLockoutFacade AccountLockoutFacade
```

### New file: `pkg/admin/graphql/user_lockout.go`

Each field has its own typed `Resolve` func — no `map[string]interface{}`. Pattern follows `fraud_protection_overview.go`.

```go
package graphql

import (
    "github.com/graphql-go/graphql"

    apimodel "github.com/authgear/authgear-server/pkg/api/model"
)

var lockedIPType = graphql.NewObject(graphql.ObjectConfig{
    Name:        "LockedIP",
    Description: "A locked IP address and when its lock expires",
    Fields: graphql.Fields{
        "ipAddress": &graphql.Field{
            Type:        graphql.NewNonNull(graphql.String),
            Description: "The locked IP address",
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                return p.Source.(apimodel.LockedIP).IPAddress, nil
            },
        },
        "lockedUntil": &graphql.Field{
            Type:        graphql.NewNonNull(graphql.DateTime),
            Description: "The time the lock for this IP expires",
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                return p.Source.(apimodel.LockedIP).LockedUntil, nil
            },
        },
    },
})

var accountLockoutType = graphql.NewObject(graphql.ObjectConfig{
    Name:        "AccountLockout",
    Description: "The account lockout state of a user",
    Fields: graphql.Fields{
        "lockoutType": &graphql.Field{
            Type:        graphql.NewNonNull(graphql.String),
            Description: "The configured lockout type: \"per_user\" or \"per_user_per_ip\"",
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                return string(p.Source.(*apimodel.AccountLockoutStatus).LockoutType), nil
            },
        },
        "isLocked": &graphql.Field{
            Type:        graphql.NewNonNull(graphql.Boolean),
            Description: "Whether the user is currently locked",
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                return p.Source.(*apimodel.AccountLockoutStatus).IsLocked, nil
            },
        },
        "lockedUntil": &graphql.Field{
            Type:        graphql.DateTime,
            Description: "When the global lock expires. Non-nil only for per_user lockout type",
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                return p.Source.(*apimodel.AccountLockoutStatus).LockedUntil, nil
            },
        },
        "lockedIPs": &graphql.Field{
            Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(lockedIPType))),
            Description: "Locked IPs ordered by lockedUntil descending. Non-empty only for per_user_per_ip lockout type",
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                ips := p.Source.(*apimodel.AccountLockoutStatus).LockedIPs
                out := make([]interface{}, len(ips))
                for i, ip := range ips {
                    out[i] = ip
                }
                return out, nil
            },
        },
    },
})

func init() {
    nodeUser.AddFieldConfig("accountLockout", &graphql.Field{
        Type:        graphql.NewNonNull(accountLockoutType),
        Description: "The account lockout state of this user",
        Resolve: func(p graphql.ResolveParams) (interface{}, error) {
            source := p.Source.(*apimodel.User)
            ctx := p.Context
            gqlCtx := GQLContext(ctx)
            return gqlCtx.AccountLockoutFacade.GetAccountLockoutStatus(ctx, source.ID)
        },
    })
}
```

**Note on `lockedIPs`:** `graphql-go` requires `[]interface{}` for list fields; the resolve func converts `[]apimodel.LockedIP` to `[]interface{}`. Each element is passed as `apimodel.LockedIP` (value type) to `lockedIPType`'s field resolvers, which assert `p.Source.(apimodel.LockedIP)`.

### New file: `pkg/admin/graphql/lockout_mutation.go`

```go
package graphql

import (
    "github.com/graphql-go/graphql"

    relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

    "github.com/authgear/authgear-server/pkg/api/apierrors"
    "github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var resetAccountLockoutInput = graphql.NewInputObject(graphql.InputObjectConfig{
    Name: "ResetAccountLockoutInput",
    Fields: graphql.InputObjectConfigFieldMap{
        "userID": &graphql.InputObjectFieldConfig{
            Type:        graphql.NewNonNull(graphql.ID),
            Description: "Target user ID.",
        },
    },
})

var resetAccountLockoutPayload = graphql.NewObject(graphql.ObjectConfig{
    Name: "ResetAccountLockoutPayload",
    Fields: graphql.Fields{
        "user": &graphql.Field{
            Type: graphql.NewNonNull(nodeUser),
        },
    },
})

var _ = registerMutationField(
    "resetAccountLockout",
    &graphql.Field{
        Description: "Reset the account lockout state of a user",
        Type:        graphql.NewNonNull(resetAccountLockoutPayload),
        Args: graphql.FieldConfigArgument{
            "input": &graphql.ArgumentConfig{
                Type: graphql.NewNonNull(resetAccountLockoutInput),
            },
        },
        Resolve: func(p graphql.ResolveParams) (interface{}, error) {
            input := p.Args["input"].(map[string]interface{})
            userNodeID := input["userID"].(string)

            resolvedNodeID := relay.FromGlobalID(userNodeID)
            if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
                return nil, apierrors.NewInvalid("invalid user ID")
            }
            userID := resolvedNodeID.ID

            ctx := p.Context
            gqlCtx := GQLContext(ctx)

            err := gqlCtx.AccountLockoutFacade.ResetAccountLockout(ctx, userID)
            if err != nil {
                return nil, err
            }

            return graphqlutil.NewLazyValue(map[string]interface{}{
                "user": gqlCtx.Users.Load(ctx, userID),
            }).Value, nil
        },
    },
)
```

---

## 8. Wiring

### `pkg/admin/deps.go`

Add import `lockoutpkg "github.com/authgear/authgear-server/pkg/lib/lockout"`.

Add to `DependencySet`:

```go
wire.Struct(new(facade.LockoutFacade), "*"),
wire.Bind(new(facade.LockoutProvider), new(*lockoutpkg.Service)),
wire.Bind(new(graphql.AccountLockoutFacade), new(*facade.LockoutFacade)),
```

`*lockout.Service` — already in graph via `deps.CommonDependencySet` → `lockout.DependencySet`.
`*config.AuthenticationLockoutConfig` — already in graph via `wire.FieldsOf(new(*config.AuthenticationConfig), "Lockout")` in `deps_config.go`.

### `pkg/admin/wire_gen.go`

Regenerate with `wire`. Do not edit by hand.

---

## 9. Compatibility and Deployment

- **No migration.** New code only reads and deletes existing Redis keys using the same key format.
- **Zero-downtime.** The new field and mutation are additive.
- **Disabled config.** `GetAccountLockoutStatus` returns `{isLocked: false, lockedIPs: []}`. `ResetAccountLockout` is a no-op.
- **HGETALL gap.** Documented in §2. Only affects `history_duration < maximum_duration` configurations.

---

## 10. Test Plan

### Unit tests — `pkg/lib/lockout/`

Add `storage_redis_admin_test.go` (requires real Redis, following `record_test.go` pattern with Convey BDD style):

| Test | Setup | Expected |
|---|---|---|
| `TestGetStatus_PerUser_NotLocked` | Empty Redis | `IsLocked=false`, `LockedUntil=nil` |
| `TestGetStatus_PerUser_Locked` | `makeAttempts` until locked | `IsLocked=true`, `LockedUntil` non-nil and in the future |
| `TestGetStatus_PerUser_Expired` | Manually write past epoch to lock key | `IsLocked=false` |
| `TestGetStatus_PerUserPerIP_NotLocked` | Empty Redis | `IsLocked=false`, `LockedIPs=nil` |
| `TestGetStatus_PerUserPerIP_SomeLocked` | `makeAttempts` for two IPs, only one past threshold | `IsLocked=true`, `LockedIPs` has exactly the locked IP |
| `TestClearAll_PerUser` | `makeAttempts` until locked; `ClearAll`; `GetStatus` | `IsLocked=false` |
| `TestClearAll_PerUserPerIP` | `makeAttempts` for two IPs until locked; `ClearAll`; `GetStatus` | `IsLocked=false`, `LockedIPs=nil` |
| `TestGetStatus_DisabledSpec` | Disabled spec | Returns `{IsLocked: false}`, no Redis call |
| `TestClearAll_DisabledSpec` | Disabled spec | Returns nil, no Redis call |

### E2E tests — `e2e/tests/admin_api/account_lockout_test.yaml`

| Case | Description |
|---|---|
| Query `accountLockout` — not locked | Create user; query via `node(id: $userID)` relay query; verify `isLocked=false`, `lockoutType` set, `lockedIPs=[]` |
| `resetAccountLockout` — clears lockout state | Create user; call `resetAccountLockout` mutation; verify user is returned with `accountLockout.isLocked=false` |

Note: E2E tests use YAML format (following project convention in `e2e/tests/`). Advanced cases (locked per_user, per_user_per_ip, expired locks) are covered by unit tests in `pkg/lib/lockout/`.

---

## 11. Atomic Commit Plan

| # | Commit message | Files | Notes |
|---|---|---|---|
| 1 | `Add lockout GetStatus and ClearAll to storage and service` | `pkg/api/model/lockout.go` (new), `pkg/lib/config/authentication_lockout.go`, `pkg/lib/lockout/models.go`, `pkg/lib/lockout/storage.go`, `pkg/lib/lockout/record.go`, `pkg/lib/lockout/storage_redis.go`, `pkg/lib/lockout/service.go` | `AuthenticationLockoutType` aliased to `model.AccountLockoutType`. Two new Lua scripts. Unit tests in same commit. |
| 2 | `[Admin API] Add LockoutFacade` | `pkg/admin/facade/lockout.go`, `pkg/admin/facade/deps.go` | Depends on commit 1. |
| 3 | `[Admin API] Add accountLockout field to User node and resetAccountLockout mutation` | `pkg/admin/graphql/context.go`, `pkg/admin/graphql/user_lockout.go`, `pkg/admin/graphql/lockout_mutation.go`, `pkg/admin/deps.go`, `pkg/admin/wire_gen.go` | Wire regeneration required in same commit. |
| 4 | `[Admin API] E2E tests for accountLockout query and resetAccountLockout mutation` | `e2e/tests/admin_api/account_lockout_test.yaml` | YAML-based E2E tests (project convention). Tests unlocked query and reset mutation. |
