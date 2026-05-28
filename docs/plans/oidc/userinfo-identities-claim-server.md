# Implementation Plan: `https://authgear.com/claims/user/identities`

## 1. Goal / Scope

Add `https://authgear.com/claims/user/identities` to the OIDC userinfo endpoint.
Each element exposes `type` (string), for login ID identities `login_id_key` (string), and for OAuth identities `provider_alias` (string).

This follows the same pattern as `https://authgear.com/claims/user/authenticators`:
- Returned from the userinfo endpoint only (not embedded in the ID token).
- Gated by `isClaimAllowed` in `GetUserInfo`, so it is only included when the client requests the claim in scope (first-party clients get it via `full-access`).
- Cached inside the existing `UserInfo` Redis cache; no new cache key.

Spec: `docs/specs/user-profile/design.md` (Special Claims) and `docs/specs/sdk-settings-actions.md` (Full UserInfo Design).

---

## 2. Model Changes

### `pkg/api/model/claims.go`

Add one constant after `ClaimAuthenticators`:

```go
ClaimIdentities ClaimName = "https://authgear.com/claims/user/identities"
```

### `pkg/api/model/userinfo.go`

Add a new struct after `UserInfoAuthenticator`:

```go
type UserInfoIdentity struct {
    CreatedAt     time.Time          `json:"created_at"`
    UpdatedAt     time.Time          `json:"updated_at"`
    Type          model.IdentityType `json:"type"`
    LoginIDKey    string             `json:"login_id_key,omitempty"`
    ProviderAlias string             `json:"provider_alias,omitempty"`
}
```

Mirrors `UserInfoAuthenticator` which also carries `created_at`/`updated_at`.

---

## 3. UserInfo Service

### `pkg/lib/userinfo/userinfo.go`

**New interface** (place alongside the existing `UserInfoAuthenticatorService` interface):

```go
type UserInfoIdentityService interface {
    ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
}
```

This is satisfied by `*identity/service.Service`, which already has:
```go
// pkg/lib/authn/identity/service/service.go
func (s *Service) ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
```

**Add field to `UserInfoService` struct:**

```go
IdentityService UserInfoIdentityService
```

**Add field to `UserInfo` struct:**

```go
Identities []model.UserInfoIdentity `json:"identities"`
```

**Extend `getUserInfoFromDatabase()`** — add after the authenticators block (before the `recoveryCodes` block):

```go
identityInfos, err := s.IdentityService.ListByUser(ctx, userID)
if err != nil {
    return nil, err
}

userinfoIdentities := []model.UserInfoIdentity{}
for _, info := range identityInfos {
    uiIdentity := model.UserInfoIdentity{
        CreatedAt: info.CreatedAt,
        UpdatedAt: info.UpdatedAt,
        Type:      info.Type,
    }
    if info.Type == model.IdentityTypeLoginID && info.LoginID != nil {
        uiIdentity.LoginIDKey = info.LoginID.LoginIDKey
    }
    if info.Type == model.IdentityTypeOAuth && info.OAuth != nil {
        uiIdentity.ProviderAlias = info.OAuth.ProviderAlias
    }
    userinfoIdentities = append(userinfoIdentities, uiIdentity)
}
```

`info.CreatedAt` and `info.UpdatedAt` are on `identity.Info` (line 9–10 of `pkg/lib/authn/identity/info.go`).

`info.LoginID.LoginIDKey` is the `LoginIDKey string` field on `pkg/lib/authn/identity/loginid_identity.go` (the configured key name, e.g. `"email"`, `"phone"`, `"username"`).

`info.OAuth.ProviderAlias` is the `ProviderAlias string` field defined at `pkg/lib/authn/identity/oauth_identity.go`.

**Update the return value:**

```go
return &UserInfo{
    User:                    u,
    AccountAccountStaleFrom: u.AccountStatusStaleFrom,
    EffectiveRoleKeys:       roleKeys,
    Authenticators:          userinfoAuthens,
    Identities:              userinfoIdentities,
    RecoveryCodeEnabled:     len(recoveryCodes) > 0,
}, nil
```

---

## 4. OIDC Token Issuer

### `pkg/lib/oauth/oidc/id_token.go`

In `GetUserInfo()`, add after the `ClaimAuthenticators` block:

```go
if isClaimAllowed(string(model.ClaimIdentities)) {
    out[string(model.ClaimIdentities)] = userInfo.Identities
}
```

No changes to `PopulateUserClaimsInIDToken` — the identities claim is userinfo-only, matching the authenticators claim.

---

## 5. Wire / DI

`pkg/lib/userinfo/deps.go` uses `wire.Struct(new(UserInfoService), "*")`, which injects all exported fields by type automatically. No change is needed to `deps.go` itself — wire will pick up `IdentityService UserInfoIdentityService` as long as the concrete type is bound.

The concrete type `*identity/service.Service` is already present in all wire scopes that construct `UserInfoService` (auth, admin, resolver, redisqueue). The existing wire graphs already provide `*identity/service.Service` for other dependencies in those same scopes.

After adding the field, run:

```
make generate
```

This regenerates all `wire_gen.go` files (`pkg/auth/wire_gen.go`, `pkg/admin/wire_gen.go`, `pkg/resolver/wire_gen.go`, `pkg/redisqueue/wire_gen.go`). Each `userInfoService := &userinfo.UserInfoService{...}` block gains `IdentityService: identityService`.

---

## 6. Cache / Deployment Compatibility

- **Cache key**: unchanged — `app:{appID}:userinfo:{userID}:{role}`.
- **Cached shape change**: the `UserInfo` JSON gains a new `"identities"` field. Old cached entries lack it; Go's `json.Unmarshal` leaves `Identities` as `nil`.
- **Serving stale cache**: during the deploy window, a cached entry without `"identities"` will cause the claim to be absent from the response. This is transient — `duration.Short` (5 minutes) is the cache TTL, so all stale entries expire quickly.
- **Identity change invalidation**: `pkg/lib/userinfo/sink.go` calls `PurgeUserInfo` for every userID returned by `RequireReindexUserIDs()` and `DeletedUserIDs()` on each non-blocking event. Identity add/remove events already trigger reindex, so caches are purged on identity changes post-deploy.
- **No cache key version bump required.**

---

## 7. Mock Regeneration

`pkg/lib/userinfo/userinfo.go` has a `//go:generate` directive that produces `userinfo_mock_test.go`. After adding `UserInfoIdentityService`, run:

```
make generate
```

This regenerates the mock for the new interface so tests can mock `IdentityService`.

---

## 8. Test Plan

### `pkg/lib/oauth/oidc/id_token_test.go`

Style: Convey BDD (existing file imports `goconvey`).

**Extend `TestGetUserInfo`** (the `TestIDTokenIssuer_GetUserInfo` Convey block):

1. Add `Identities: []model.UserInfoIdentity{...}` to the mock `userinfo.UserInfo` return value in `mockUserInfoService.EXPECT()`.
2. Add `string(model.ClaimIdentities)` to the `scopes` slice passed to `oauth.ClientClientLike`.
3. Assert the output JSON contains `"https://authgear.com/claims/user/identities"` with the expected array.

Concrete cases to cover:

| Scenario | `Identities` in mock | Expected JSON key |
|---|---|---|
| OAuth identity | `[{CreatedAt: t, UpdatedAt: t, Type: "oauth", ProviderAlias: "google"}]` | `[{"created_at":"...","updated_at":"...","type":"oauth","provider_alias":"google"}]` |
| Login ID identity | `[{CreatedAt: t, UpdatedAt: t, Type: "login_id", LoginIDKey: "email"}]` | `[{"created_at":"...","updated_at":"...","type":"login_id","login_id_key":"email"}]` |
| Mixed | both of the above | both elements |
| Empty | `[]` | `[]` |

**Extend `TestGetUserInfo`** (the `TestGetUserInfo` Convey block, which uses map assertions):

Add:
```go
So(userInfo[string(model.ClaimIdentities)], ShouldResemble, []model.UserInfoIdentity{
    {Type: model.IdentityTypeOAuth, ProviderAlias: "google"},
})
```

---

## 9. File-Level Change Summary

| File | Change |
|---|---|
| `pkg/api/model/claims.go` | Add `ClaimIdentities` constant |
| `pkg/api/model/userinfo.go` | Add `UserInfoIdentity` struct |
| `pkg/lib/userinfo/userinfo.go` | Add interface, field on service, field on UserInfo, populate in `getUserInfoFromDatabase` |
| `pkg/lib/oauth/oidc/id_token.go` | Add `ClaimIdentities` to `GetUserInfo` output |
| `pkg/auth/wire_gen.go` | Regenerated — `IdentityService` field added to `UserInfoService` construction |
| `pkg/admin/wire_gen.go` | Regenerated |
| `pkg/resolver/wire_gen.go` | Regenerated |
| `pkg/redisqueue/wire_gen.go` | Regenerated |
| `pkg/lib/userinfo/userinfo_mock_test.go` | Regenerated (new mock for `UserInfoIdentityService`) |
| `pkg/lib/oauth/oidc/id_token_test.go` | Extend existing Convey tests |

---

## 10. Atomic Commit Plan

### Commit 1 — Model types
**Files:** `pkg/api/model/claims.go`, `pkg/api/model/userinfo.go`

Add `ClaimIdentities` constant and `UserInfoIdentity` struct. No behavior change.

### Commit 2 — UserInfo service: populate identities
**Files:** `pkg/lib/userinfo/userinfo.go`

Add `UserInfoIdentityService` interface, `IdentityService` field on `UserInfoService`, `Identities` field on `UserInfo`, and the population logic in `getUserInfoFromDatabase`.

### Commit 3 — Expose claim in OIDC userinfo
**Files:** `pkg/lib/oauth/oidc/id_token.go`

Add `ClaimIdentities` output block in `GetUserInfo`.

### Commit 4 — Wire regeneration
**Files:** `pkg/auth/wire_gen.go`, `pkg/admin/wire_gen.go`, `pkg/resolver/wire_gen.go`, `pkg/redisqueue/wire_gen.go`, `pkg/lib/userinfo/userinfo_mock_test.go`

Run `make generate`. Commit all regenerated files together. Build must pass (`go build ./...`).

### Commit 5 — Tests
**Files:** `pkg/lib/oauth/oidc/id_token_test.go`

Extend Convey tests to cover the identities claim in all cases listed in the test plan. Run `go test ./pkg/lib/oauth/oidc/...` to verify.
