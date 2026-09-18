# Resource Indicator Part 1 — Client categories and the four-key access policy

Spec: [docs/specs/api-resource.md — Access Policy](../../specs/api-resource.md#access-policy).

This part is pure data model and read paths. Nothing changes behaviour on its own: after it lands, the only key any caller can set is still `allow_dynamic_third_party_client_access`, and the only category anything asks for is still dynamic-third-party. Part 2 opens the other three keys to the Admin API; Part 3 makes the OAuth endpoints ask for the caller's real category.

## 1. Goal / Scope

`model.AccessPolicy` (`pkg/api/model/resource.go:7-9`) has exactly one field today, and the two JSONB read paths — `Store.GetResourceByURIForThirdPartyAccess` (`pkg/lib/resourcescope/store_resource.go:176`) and `Store.ListScopesForThirdPartyAccess` (`pkg/lib/resourcescope/store_scope.go:322`) — hardcode its JSON key through a single constant, `accessPolicyAllowDynamicThirdPartyClientAccessKey` (`pkg/lib/resourcescope/resource.go:16`).

In scope:

- Three new fields on `model.AccessPolicy`, one per remaining client category.
- One method, `AccessPolicy.AllowsClient`, mapping a client's existing `IsDynamicClient()`/`IsThirdParty()` pair onto those four fields.
- Moving the policy check out of SQL and into Go, which is what lets the check be expressed that way.

Out of scope: a migration (there is none to write — see §2), the Admin API (Part 2), the OAuth endpoints (Part 3).

## 2. No migration

`_auth_resource.access_policy` and `_auth_resource_scope.access_policy` are `jsonb NOT NULL DEFAULT '{}'` (`cmd/authgear/cmd/cmddatabase/migrations/authgear/20260817120002-add_resource_access_policy.sql`). An absent key already reads as `false` everywhere — the JSONB lookups are `(access_policy->>'<key>')::boolean IS TRUE`, which is `false` for a missing key, a JSON `null` and a JSON `false` alike. Adding three keys to the Go struct therefore needs no schema change and no backfill, and every pre-existing row keeps the behaviour it has today: the three new categories are denied, because the keys are not there.

This is the flat-boolean shape the spec's Extensibility note is describing, and it is the reason the spec could be written as "a key that is not set means that category is not allowed" rather than as a migration story.

## 3. `model.AccessPolicy` — three new fields

`pkg/api/model/resource.go`:

```go
type AccessPolicy struct {
	AllowStaticFirstPartyClientAccess  bool `json:"allow_static_first_party_client_access,omitempty"`
	AllowStaticThirdPartyClientAccess  bool `json:"allow_static_third_party_client_access,omitempty"`
	AllowDynamicFirstPartyClientAccess bool `json:"allow_dynamic_first_party_client_access,omitempty"`
	AllowDynamicThirdPartyClientAccess bool `json:"allow_dynamic_third_party_client_access,omitempty"`
}
```

Keep `omitempty` on all four. `false` and absent mean the same thing, so omitting a `false` key keeps stored objects small and keeps a round-trip through `CreateResource` from writing four keys where the caller set one. The existing struct-level doc comment ("Missing keys default to false, so the zero value is 'no access' for every grant this policy governs") is still accurate and covers all four; do not restate it per field.

The struct is shared: `resourcescope.Resource.AccessPolicy` and `resourcescope.Scope.AccessPolicy` are already `model.AccessPolicy` (`resource.go:58`, `scope.go:53`), and `ToModel()` copies it through. Adding fields here reaches the store, the Admin API and the portal codegen without further plumbing.

## 4. `AccessPolicy.AllowsClient`

No new client-classification type. The two predicates that describe a client already exist — `OAuthClientConfig.IsDynamicClient()` (`pkg/lib/config/oauth.go:334`) and `IsThirdParty()` (`:411`) — and the spec's four categories are exactly their 2×2. One method on the policy maps that pair onto the four fields, in `pkg/api/model/resource.go` beside the struct:

```go
// AllowsClient reports whether p opens this Resource or Scope to a client,
// given config.OAuthClientConfig's IsDynamicClient() and IsThirdParty().
// The pair is the spec's four client categories (docs/specs/api-resource.md
// § Client categories); this is the only place it is mapped onto fields.
//
// An m2m client is neither dynamic nor third-party and so would read
// AllowStaticFirstPartyClientAccess here, which would be wrong -- m2m is in
// no category at all. Nothing calls this for one: client_credentials is its
// only grant, access_policy does not govern that grant, and
// /oauth2/authorize rejects m2m before any policy is read
// (handler_authz.go:252, :1078).
func (p AccessPolicy) AllowsClient(isDynamicClient, isThirdParty bool) bool {
	switch {
	case isDynamicClient && isThirdParty:
		return p.AllowDynamicThirdPartyClientAccess
	case isDynamicClient:
		return p.AllowDynamicFirstPartyClientAccess
	case isThirdParty:
		return p.AllowStaticThirdPartyClientAccess
	default:
		return p.AllowStaticFirstPartyClientAccess
	}
}
```

The switch order matters: `isDynamicClient && isThirdParty` is tested first so a dynamic third-party client cannot fall into the static-third-party arm.

Call it as `policy.AllowsClient(client.IsDynamicClient(), client.IsThirdParty())`. Two bare booleans in a signature are a transposition hazard in general; here the argument expressions name themselves at every call site, and there is only one call shape in the codebase. That is the trade for not adding a type — worth stating so the next reader does not "fix" it by inventing one.

## 5. The JSON keys stop being duplicated

`pkg/lib/resourcescope/resource.go:11-16` holds `accessPolicyAllowDynamicThirdPartyClientAccessKey`, a hand-copied duplicate of a struct tag, existing solely so the raw SQL in §6 could name the key. §6 removes that SQL, so **delete the constant** and add nothing in its place. After this part the four JSON names appear exactly once each, in the `json:"..."` tags, and `AllowsClient` is the only thing that chooses between them.

That is the main reason the check moves to Go: with four keys, the alternative is four hand-copied strings that nothing type-checks against the struct, and a silently-denied category is what a typo in one of them looks like in production.

## 6. Store read paths drop the policy predicate

Both read paths currently filter on the policy in SQL. Neither can any more — the filter is now a Go method — so both lose the predicate and the filtering moves to their caller (Part 3 §2).

- **`GetResourceByURIForThirdPartyAccess`** (`store_resource.go:176`) — delete it. `GetResourceByURI` (`:~150`) already returns the Resource with its `AccessPolicy` populated (`scanResource` reads the column, `:311`, `:333`), which is everything the caller needs.
- **`ListScopesForThirdPartyAccess`** (`store_scope.go:322`) — rename to `ListScopesByResourceID` and drop the `access_policy` term from its `Where`, leaving `s.resource_id = ?`. It is the package's only unpaginated list-all-scopes-of-a-resource method, which is exactly what the caller now wants.

Both lose their `fmt.Sprintf` into a `Where` clause. That was never an injection — the key came from a constant, not a request — but it is one less place where a hand-built SQL string has to be re-argued in review, and `uri`/`resourceID` were already parameterised.

**The trade-off, stated plainly:** `ListScopesByResourceID` now returns every scope of the Resource and the caller discards the ones the policy excludes, where before the database returned only the permitted rows. Scopes per Resource is a hand-maintained permission list — tens, not thousands — and the query is already indexed on `(app_id, resource_id, scope)`, so this is a few extra rows on a resource-bound token issuance. If a project ever did have a Resource with thousands of scopes, the fix is a predicate back in SQL, and it would need the key map back with it.

`GetResourceByURI` has other callers (the Admin API path). Deleting the `...ForThirdPartyAccess` variant leaves them alone.

### 6.1 Call sites in this part

Part 3 rewrites both `handler_authz.go` call sites properly. To keep this part compiling and behaviour-identical, have `validateResource` and `resourceScopeDisplayNames` call `GetResourceByURI`/`ListScopesByResourceID` and apply `AllowsClient(true, true)` — the literal pair for a dynamic third-party client, which is the only category either reaches today — then update `AuthorizationHandlerResourceScopeService` (`handler_authz.go:152-155`) to the new method set. Leave a `TODO` naming Part 3 rather than letting a hardcoded `true, true` look intentional.

Note the one behavioural subtlety this introduces and Part 3 removes: `GetResourceByURI` finds a Resource the policy excludes, so the "not allowed" case is now an explicit branch in the caller rather than a `sql.ErrNoRows`. It must return the same `invalid_target` the old not-found path produced — see Part 3 §2, which makes that structural.

The `wire.Bind` to `*resourcescope.Store` (`pkg/lib/deps/deps_common.go:419`) is unaffected; `*resourcescope.Store` still satisfies the interface, so no `wire_gen.go` regeneration. Regenerate the gomock for the interface (`make generate`) since its method set changed.

## 7. Test plan

- `pkg/api/model/resource_test.go` (new or extend) — `AllowsClient` as a full 4×4: for each of the four argument pairs, a policy with exactly one field `true` returns `true` only for its own pair. Sixteen assertions, and they are the whole correctness argument for the mapping — a transposed arm (static-third-party reading the dynamic field, say) passes any test that only checks the diagonal.
- The same file — `json.Marshal` of an all-`true` policy has exactly the four expected keys. This is what pins the tag names now that nothing else duplicates them, and it is what Part 2 §3.4's JSONB merge depends on.
- `pkg/lib/resourcescope` store tests (extend whichever harness the package already uses for `GetResourceByURIForThirdPartyAccess`, then repoint it) — `ListScopesByResourceID` returns every scope of the Resource regardless of policy, and `GetResourceByURI` returns a Resource whose `AccessPolicy` round-tripped all four fields out of the JSONB column. The policy filtering itself is no longer the store's job and is tested at the handler in Part 3.

No `pkg/lib/config` test: this part adds nothing there.

## 8. Commit plan

1. **`Add the remaining three access policy keys`** — §3 and §4 (`AllowsClient`), plus their tests. Additive; nothing reads the new fields yet.
2. **`Check the access policy in Go instead of in SQL`** — §5, §6, §6.1, regenerated mocks, plus store tests. Behaviour-identical: the only call sites still check the dynamic-third-party pair.

Neither commit needs a `docs/BREAKING-CHANGES.md` entry — no stored data changes, no request is newly accepted or refused, and the only signatures that move are internal to `pkg/lib`.
