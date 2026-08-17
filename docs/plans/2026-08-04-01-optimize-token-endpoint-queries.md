# Optimize duplicated database queries in the token endpoint

## 1. Goal and scope

`POST /oauth2/token` spends essentially all of its wall-clock time on sequential Postgres
round-trips. Production traces show 46 DB queries per request, all inside a single
`DB Connection Use` span (one `*sql.Tx`, therefore strictly serial), with a per-query cost
of roughly 100 ms.

Baseline, from three production traces taken against build `cd39dd064f1a`:

| Trace | Endpoint | Wall | DB spans | DB time | % of wall |
|---|---|---|---|---|---|
| 1 | `POST /oauth2/token` | 5094 ms | 46 | 4877 ms | 96% |
| 2 | `POST /oauth2/token` | 6188 ms | 46 | 4406 ms | 71% |
| 3 | `POST /api/v1/authentication_flows/states/input` | 6840 ms | 42 | 5374 ms | 79% |

Query multiplicity in trace 1:

| Query | Times |
|---|---|
| `_auth_identity_login_id` join | 6 |
| `_auth_identity` refs where `type = $2` | 6 |
| `_auth_user` by id | 4 |
| `_auth_identity` refs | 4 |
| `_auth_authenticator` refs | 4 |
| `_auth_authenticator_password` join | 4 |
| `_auth_verified_claim` (all) | 3 |
| `_auth_verified_claim` (`name = ANY`) | 3 |
| `_auth_user_role` join | 3 |
| `_auth_user_group` join | 3 |
| `nextval('_auth_event_sequence')` | 2 |

The repeating 9-query block is `user.Queries.GetMany` (`pkg/lib/authn/user/queries.go:79`). It
runs three times in one token request:

```
 336ms  userinfo cache MISS -> UserInfoService.getUserInfoFromDatabase   16 queries  1877ms
2214ms  IDTokenIssuer.PrepareIDToken
        ListIdentitiesThatHaveStandardAttributes                          4 queries   453ms
        nextval(_auth_event_sequence)                                     1 query      52ms
        Resolver.Resolve -> Queries.GetMany                               9 queries  1033ms
3753ms  AccessTokenEncoding.PrepareUserAccessToken (JWT access token)
        userinfo cache HIT                                                            0.5ms
        ListIdentitiesThatHaveStandardAttributes                          4 queries   337ms  <- identical args
        nextval(_auth_event_sequence)                                     1 query      30ms
        Resolver.Resolve -> Queries.GetMany                               9 queries   915ms  <- identical args
```

About 2.8 s of the 5.1 s is spent building the payloads for two blocking webhook events,
`oidc.id_token.pre_create` (`pkg/lib/oauth/oidc/id_token.go:148,170`) and `oidc.jwt.pre_create`
(`pkg/lib/oauth/token_encoding.go:140,161`). Both resolve the same user with the same role and
list the same identities. `hook.Sink.ReceiveBlockingEvent` (`pkg/lib/hook/sink.go:57`) discards
the result unless a matching `blocking_handlers` entry exists; the audit, reindex, and userinfo
sinks are no-ops for blocking events (`pkg/lib/audit/sink.go:19`,
`pkg/lib/search/reindex/sink.go:18`, `pkg/lib/userinfo/sink.go:13`).

### In scope

1. Skip blocking-event payload resolution when no hook is registered for that event type.
2. Remove the duplicated user/identity reads in the token endpoint by carrying the value
   through explicitly. Not a cache.
3. Collapse the redundant verified-claims query inside `user.Queries.GetMany`.
4. Fix the duplicated payload resolve inside `event.Service.DispatchEventImmediately`.
5. For non-blocking payloads: resolve twice only when the payload declares a deleted user;
   otherwise resolve once, keeping the later (commit-time) resolve.

### Explicitly out of scope

- Collapsing identity-ref-then-detail and authenticator-ref-then-detail into single joins.
  `identity/service.Service.ListByUserIDs` fans out to seven per-type stores; a single
  `LEFT JOIN` formulation would require restructuring all of them. Tracked separately.
- The ~100 ms per-query floor (two round-trips per parameterised query under `lib/pq`; the
  per-connection statement cache at `pkg/lib/infra/db/prepared_statements_handle.go` is only
  wired into `userimport`) and the Postgres network placement. Both are larger, separate
  changes and are multipliers on, not alternatives to, this work.
- `MaxOpenConnection` tuning. The `DB Connection Wait` of 626 ms in trace 3 is a symptom of
  requests holding a connection for 3–5 s; reducing query count is the fix.
- Merging the per-user role and group listing queries (`Store.ListRolesByUserIDs` and
  `Store.ListGroupsByUserIDs`) into one `UNION ALL` query. Prototyped and reverted: unlike
  every other change here, the two queries are not redundant with each other (roles and groups
  are different data, both genuinely needed), so combining them doesn't remove any scan or
  join Postgres was already doing — it trades one round trip for a cross-arm sort the combined
  query has to do that neither original query did. That's a real cost with no matching
  reduction in database work, unlike the round-trip reductions elsewhere in this plan, which
  all remove queries that were either duplicated or entirely unobserved.

### Expected result

Per token request, for an app with no `oidc.*_pre_create` blocking hooks: 46 → about 19
queries. For an app that does register those hooks: 46 → about 31. `GetMany` itself goes from
9 to 8 queries for every caller, including the userinfo cache-miss path and the
authentication-flow endpoints.

## 2. Change 1 — gate blocking-event resolution on hook registration

### 2.1 Sink interface

`pkg/lib/event/service.go` — extend the `Sink` interface so the capability is compile-enforced
rather than discovered by type assertion:

```go
type Sink interface {
	ReceiveBlockingEvent(ctx context.Context, e *event.Event) error
	ReceiveNonBlockingEvent(ctx context.Context, e *event.Event) error
	WillDeliverBlockingEvent(eventType event.Type) bool
}
```

Implementations:

- `pkg/lib/hook/sink.go` — already has `WillDeliverBlockingEvent` at line 182. No change.
- `pkg/lib/audit/sink.go` — add `func (s *Sink) WillDeliverBlockingEvent(eventType event.Type) bool { return false }`.
- `pkg/lib/search/reindex/sink.go` — same.
- `pkg/lib/userinfo/sink.go` — same.

Each addition sits directly above the existing `ReceiveBlockingEvent`, with a one-line comment
stating that the sink does not consume blocking events.

New method on `event.Service`:

```go
// WillDeliverBlockingEvent reports whether any sink will act on a blocking event
// of this type. When it is false, the event payload is never observed, so the
// caller may skip populating it.
func (s *Service) WillDeliverBlockingEvent(eventType event.Type) bool {
	for _, sink := range s.Sinks {
		if sink.WillDeliverBlockingEvent(eventType) {
			return true
		}
	}
	return false
}
```

### 2.2 Resolver

`pkg/lib/event/resolve.go` — add `ResolveWithUser` so a caller that already holds the
`*model.User` can populate the payload without a read. Refactor the existing reflection walk
into one private method:

```go
type Resolver interface {                  // pkg/lib/event/service.go
	Resolve(ctx context.Context, anything any) (err error)
	ResolveWithUser(ctx context.Context, anything any, u *model.User) (err error)
}

func (r *ResolverImpl) Resolve(ctx context.Context, anything any) error {
	return r.resolve(ctx, anything, nil)
}

// ResolveWithUser resolves anything, using u for any resolve:"user" field whose
// UserRef.ID matches u.ID instead of reading the user from the database. Other
// resolve tags, and a resolve:"user" field with a different ID, still read.
func (r *ResolverImpl) ResolveWithUser(ctx context.Context, anything any, u *model.User) error {
	return r.resolve(ctx, anything, u)
}

func (r *ResolverImpl) resolve(ctx context.Context, anything any, override *model.User) (err error) { ... }
```

Inside `resolve`, at the point that currently calls `r.Users.Get`:

```go
var u *model.User
if override != nil && jsonName == "user" && override.ID == userRef.ID {
	u = override
} else {
	u, err = r.Users.Get(ctx, userRef.ID, accesscontrol.RoleGreatest)
	if errors.Is(err, user.ErrUserNotFound) {
		continue
	}
	if err != nil {
		return
	}
}
struc.Field(j).Set(reflect.ValueOf(*u))
```

The `jsonName == "user"` guard matters: `UserAnonymousPromotedEventPayload`
(`pkg/api/event/nonblocking/user_anonymous_promoted.go:13`) also carries
`resolve:"anonymous_user"`, which must keep reading from the database.

### 2.3 Preparation options

New file `pkg/api/event/prepare_blocking.go`. The type lives in `pkg/api/event` because both
`pkg/lib/oauth` and `pkg/lib/authenticationflow` already import it and neither imports
`pkg/lib/event`:

```go
package event

// PrepareBlockingEventOptions carries optional pre-computed inputs for
// preparing a blocking event.
type PrepareBlockingEventOptions struct {
	// ResolvedUser, when non-nil, populates the payload's resolve:"user" field
	// instead of reading the user from the database. Callers set it when they
	// have already read the same user within the same request.
	ResolvedUser *model.User
}
```

### 2.4 PrepareBlockingEventWithTx

`pkg/lib/event/service.go:141` — new signature and gated resolve:

```go
func (s *Service) PrepareBlockingEventWithTx(
	ctx context.Context,
	payload event.BlockingPayload,
	opts event.PrepareBlockingEventOptions,
) (e *event.Event, err error) {
	eventContext := s.makeContext(ctx, payload)
	var seq int64
	seq, err = s.nextSeq(ctx)
	if err != nil {
		return
	}
	switch {
	case opts.ResolvedUser != nil:
		err = s.Resolver.ResolveWithUser(ctx, payload, opts.ResolvedUser)
	case s.WillDeliverBlockingEvent(payload.BlockingEventType()):
		err = s.Resolver.Resolve(ctx, payload)
	default:
		// No sink will observe the payload; skip the read entirely.
	}
	if err != nil {
		return
	}
	e = newBlockingEvent(seq, payload, eventContext)
	return
}
```

`nextSeq` stays unconditional. See §7 for why.

Call sites to update with `event.PrepareBlockingEventOptions{}`:

- `pkg/lib/authenticationflow/declarative/node_pre_initialize.go:32`
- `pkg/lib/authenticationflow/declarative/node_pre_authenticate.go:32`
- `pkg/lib/authenticationflow/declarative/node_post_identified.go:41`

Interface declarations to update:

- `pkg/lib/authenticationflow/dependencies.go:124`
- `pkg/lib/oauth/oidc/id_token.go:45` (`IDTokenIssuerEventService`) — also add
  `WillDeliverBlockingEvent(eventType event.Type) bool`
- `pkg/lib/oauth/token_encoding.go:39` (`EventService`) — same addition

### 2.5 DispatchEventOnCommit

`pkg/lib/event/service.go:60` — remove the unconditional `Resolve` at line 75 and move it into
the type switch. This is the same edit as Change 5, so it lands in one commit:

```go
switch typedPayload := payload.(type) {
case event.BlockingPayload:
	// A blocking payload is resolved here or not at all; there is no later pass.
	if s.WillDeliverBlockingEvent(typedPayload.BlockingEventType()) {
		if err = s.Resolver.Resolve(ctx, payload); err != nil {
			return
		}
	}
	eventContext := s.makeContext(ctx, payload)
	// ... unchanged: nextSeq, newBlockingEvent, sink loop
case event.NonBlockingPayload:
	// A non-blocking payload is resolved again in WillCommitTx. Resolving it
	// here as well is only observable when the later resolve cannot find the
	// user: resolve.go swallows ErrUserNotFound and keeps whatever was
	// resolved earlier. That happens exactly when the user row is gone by
	// commit time, which the payload declares via DeletedUserIDs.
	if len(typedPayload.DeletedUserIDs()) > 0 {
		if err = s.Resolver.Resolve(ctx, payload); err != nil {
			return
		}
	}
	s.NonBlockingPayloads = append(s.NonBlockingPayloads, typedPayload)
default:
	panic(...)
}
```

### 2.6 Skipping the identity list at the two OIDC call sites

`ListIdentitiesThatHaveStandardAttributes` runs before `PrepareBlockingEventWithTx`, so the gate
must also be visible to the caller. Both call sites are restructured in Change 2, which folds
the gate and the carried value into one code path — see §3.3 and §3.4.

## 3. Change 2 — carry the user and identity values through the token endpoint

### 3.1 The window is provably write-free

Both preparations run inside `h.Database.WithTx` (`handler_token.go:304`). Both hook
deliveries — the only thing that can mutate the user, via `event.PerformEffects` from
`hook/sink.go:127` — run after that transaction commits, at `handler_token.go:314` and `:328`.
So no write to the user can occur between the two resolves. The value is carried, not cached:
it is never stored, never reused across requests, and never revalidated.

### 3.2 The carried value

New file `pkg/lib/oauth/user_blocking_event_context.go`:

```go
package oauth

type UserBlockingEventContextUserService interface {
	Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error)
}

type UserBlockingEventContextIdentityService interface {
	ListIdentitiesThatHaveStandardAttributes(ctx context.Context, userID string) ([]*identity.Info, error)
}

// UserBlockingEventContext holds the user-derived values needed to populate the
// oidc.id_token.pre_create and oidc.jwt.pre_create blocking event payloads.
//
// A single token request prepares both events for the same user, so the values
// are computed once and passed explicitly to both preparation calls. This is a
// value carried within one request, not a cache: it is not stored, not reused
// across requests, and not revalidated. It is only safe because both
// preparations happen inside one transaction, before any hook can run and
// mutate the user.
type UserBlockingEventContext struct {
	UserID     string
	UserModel  *model.User
	Identities []model.Identity
}

// GetUserModel and GetIdentities are nil-receiver safe so call sites do not
// branch on whether the context was computed.
func (c *UserBlockingEventContext) GetUserModel() *model.User
func (c *UserBlockingEventContext) GetIdentities() []model.Identity

type UserBlockingEventContextProvider struct {
	Users      UserBlockingEventContextUserService
	Identities UserBlockingEventContextIdentityService
}

// Get reads the identity list and the user model. It issues 4 identity queries
// plus the 7 queries of user.Queries.GetMany.
func (p *UserBlockingEventContextProvider) Get(ctx context.Context, userID string) (*UserBlockingEventContext, error)
```

`Get` calls `p.Identities.ListIdentitiesThatHaveStandardAttributes(ctx, userID)`, maps to
`[]model.Identity` via `Info.ToModel()`, then
`p.Users.Get(ctx, userID, accesscontrol.RoleGreatest)` — the same role
`ResolverImpl.resolve` uses, so the resulting model is byte-identical to what the resolver
would have produced.

`pkg/lib/oauth` gains no new package dependency: it already imports `pkg/api/model` and
`pkg/lib/authn/identity`. `accesscontrol` is a `pkg/util` leaf. Both service fields are
interfaces bound by wire, so there is no import of `pkg/lib/authn/user`.

Wire: add `wire.Struct(new(UserBlockingEventContextProvider), "*")` to
`pkg/lib/oauth/deps.go`.

### 3.3 IDTokenIssuer

`pkg/lib/oauth/oidc/id_token.go`:

- Delete the `IDTokenIssuerIdentityService` interface (line 41) and the `Identities` field on
  `IDTokenIssuer` (line 54). The provider now owns that read.
- Add `UserBlockingEventContexts *oauth.UserBlockingEventContextProvider` to `IDTokenIssuer`.
- Add `UserBlockingEventContext *oauth.UserBlockingEventContext` to `PrepareIDTokenOptions`
  (line 79).

`PrepareIDToken`, replacing lines 148–174:

```go
var eventUserCtx *oauth.UserBlockingEventContext
if ti.Events.WillDeliverBlockingEvent(blocking.OIDCIDTokenPreCreate) {
	eventUserCtx = opts.UserBlockingEventContext
	if eventUserCtx == nil || eventUserCtx.UserID != opts.AuthenticationInfo.UserID {
		eventUserCtx, err = ti.UserBlockingEventContexts.Get(ctx, opts.AuthenticationInfo.UserID)
		if err != nil {
			return nil, err
		}
	}
}

eventPayload := &blocking.OIDCIDTokenPreCreateBlockingEventPayload{
	UserRef:    model.UserRef{Meta: model.Meta{ID: opts.AuthenticationInfo.UserID}},
	Identities: eventUserCtx.GetIdentities(),
	IDToken:    blocking.OIDCIDToken{Payload: forMutation},
}

event, err := ti.Events.PrepareBlockingEventWithTx(ctx, eventPayload, apievent.PrepareBlockingEventOptions{
	ResolvedUser: eventUserCtx.GetUserModel(),
})
```

The `UserID` mismatch check is a guard against a caller threading the wrong user; it falls back
to reading rather than emitting a payload for the wrong subject.

### 3.4 AccessTokenEncoding

`pkg/lib/oauth/token_encoding.go`, the same shape:

- Delete `AccessTokenEncodingIdentityService` (line 43) and the `Identities` field (line 53).
- Add `UserBlockingEventContexts *UserBlockingEventContextProvider` to `AccessTokenEncoding`.
- Add `UserBlockingEventContext *UserBlockingEventContext` to `EncodeUserAccessTokenOptions`
  (line 56).
- In `PrepareUserAccessToken`, replace lines 139–161 with the gate/carry/fallback block above,
  using `blocking.OIDCJWTPreCreate` and `OIDCJWTPreCreateBlockingEventPayload`.

Note that the early return for `!options.ClientConfig.IssueJWTAccessToken` (line 93) already
skips all of this for opaque access tokens.

### 3.5 Threading

Add `UserBlockingEventContext *UserBlockingEventContext` to
`oauth.PrepareUserAccessGrantOptions` (`pkg/lib/oauth/grant_access_service.go:20`), and copy it
into `EncodeUserAccessTokenOptions` inside `AccessGrantService.PrepareUserAccessGrant` (line
67). `handler.PrepareUserAccessGrantByRefreshTokenOptions`
(`pkg/lib/oauth/handler/service_token.go:59`) embeds `PrepareUserAccessGrantOptions`, so it
needs no change.

New helper on `TokenHandler` in `handler_token.go`:

```go
// resolveUserBlockingEventContext reads the values shared by the
// oidc.id_token.pre_create and oidc.jwt.pre_create payloads, once, so the two
// preparations in this request do not each read the same rows. It returns nil
// when neither event will be delivered.
func (h *TokenHandler) resolveUserBlockingEventContext(ctx context.Context, userID string) (*oauth.UserBlockingEventContext, error) {
	if !h.Events.WillDeliverBlockingEvent(blocking.OIDCIDTokenPreCreate) &&
		!h.Events.WillDeliverBlockingEvent(blocking.OIDCJWTPreCreate) {
		return nil, nil
	}
	return h.UserBlockingEventContexts.Get(ctx, userID)
}
```

`TokenHandler` gains `UserBlockingEventContexts *oauth.UserBlockingEventContextProvider` and
its `Events` interface (`handler_token.go:77`) gains
`WillDeliverBlockingEvent(event.Type) bool`.

Call it once in each of the four functions that prepare both events, and set the result on both
option structs:

| Function | `PrepareIDTokenOptions` | `PrepareUserAccessGrantOptions` |
|---|---|---|
| `handleAnonymousRequest` | `handler_token.go:1151` | `:1133` |
| `handleBiometricAuthenticate` | `:1438` | `:1420` |
| `doIssueTokensForAuthorizationCode` | `:1966` | `:1940` |
| `issueTokensForRefreshToken` | `:2009` | `:2021` |

Sites that prepare only one of the two events are left alone; they hit the nil fallback in
§3.3/§3.4, which costs exactly what it costs today:

- `handlePreAuthenticatedURLToken` (`:985`) — ID token only.
- `handleIDToken` (`:1643`) — ID token only.
- `TokenService.PrepareUserAccessGrantByRefreshToken` reached from
  `service_preauthenticated_url.go:106` — access grant only.
- `handler_anonymous_user.go:189,251` and `pkg/admin/facade/oauth.go:112` — access grant only.

## 4. Change 3 — read verified claims once inside GetMany

`user.Queries.GetMany` (`pkg/lib/authn/user/queries.go:79`) issues 9 queries. One is
removable without touching the per-type identity/authenticator stores.

### 4.1 Verified claims: read once instead of twice

Today:

- `Verification.AreUsersVerified` → `GetVerificationStatuses` →
  `ClaimStore.ListByUserIDs(userIDs)` — all claims (`pkg/lib/feature/verification/service.go:100`).
- `StandardAttributes.DeriveStandardAttributesForUsers` →
  `ClaimStore.ListByUserIDsAndClaimNames(userIDs, [email, phone_number])`
  (`pkg/lib/feature/stdattrs/service_noevent.go:213`) — a strict subset of the above.

`pkg/lib/feature/verification/service.go`:

```go
// ListClaimsByUserIDs returns every verified claim of these users.
func (s *Service) ListClaimsByUserIDs(ctx context.Context, userIDs []string) ([]*Claim, error) {
	return s.ClaimStore.ListByUserIDs(ctx, userIDs)
}

// AreUsersVerifiedWithClaims is AreUsersVerified with the claims already read.
func (s *Service) AreUsersVerifiedWithClaims(
	ctx context.Context,
	identitiesByUserIDs map[string][]*identity.Info,
	claims []*Claim,
) (map[string]bool, error)
```

Refactor: extract the body of `GetVerificationStatuses` after its `ListByUserIDs` call into
`func (s *Service) getVerificationStatusesWithClaims(idensByUserID map[string][]*identity.Info, claims []*Claim) map[string][]ClaimStatus`.
`GetVerificationStatuses` and `AreUsersVerified` keep their current signatures and behaviour by
reading claims then delegating. `AreUsersVerifiedWithClaims` holds the per-user aggregation
currently in `AreUsersVerified` (line 197).

`pkg/lib/feature/stdattrs/service_noevent.go`:

```go
// DeriveStandardAttributesForUsersWithClaims is DeriveStandardAttributesForUsers
// with the claims already read. claims may contain claims of any name; only
// email and phone_number are consulted.
func (s *ServiceNoEvent) DeriveStandardAttributesForUsersWithClaims(
	ctx context.Context,
	role accesscontrol.Role,
	userIDs []string,
	updatedAts []time.Time,
	attrsList []map[string]any,
	claims []*verification.Claim,
) (map[string]map[string]any, error)
```

Move the existing body from line 219 onward into it. `DeriveStandardAttributesForUsers` becomes
the `ListByUserIDsAndClaimNames` read plus a delegation, so its behaviour and query are
unchanged for all other callers. The name filter is already applied per claim inside the loop
(`claim.Name != stdattrs.Email` / `!= stdattrs.PhoneNumber`), so passing the unfiltered set is
correct without further changes.

`pkg/lib/authn/user/queries.go` — widen the two interfaces and read once:

```go
type VerificationService interface {
	IsUserVerified(ctx context.Context, identities []*identity.Info) (bool, error)
	ListClaimsByUserIDs(ctx context.Context, userIDs []string) ([]*verification.Claim, error)
	AreUsersVerifiedWithClaims(ctx context.Context, identitiesByUserIDs map[string][]*identity.Info, claims []*verification.Claim) (map[string]bool, error)
}

type StandardAttributesService interface {
	DeriveStandardAttributes(...)  // unchanged
	DeriveStandardAttributesForUsersWithClaims(ctx context.Context, role accesscontrol.Role, userIDs []string, updatedAts []time.Time, attrsList []map[string]any, claims []*verification.Claim) (map[string]map[string]any, error)
}
```

`AreUsersVerified` is dropped from the `VerificationService` interface; `GetPageForExport`
(line 184) is updated the same way as `GetMany`.

No import cycle: `pkg/lib/feature/verification` does not import `pkg/lib/authn/user`, and it
cannot start to, because `pkg/lib/authn/user` already imports `pkg/lib/authn/identity` which
`verification` also imports.

### 4.2 Resulting query count

`GetMany` goes from 9 to 8 queries: `_auth_user`, identity refs, identity detail, authenticator
refs, authenticator detail, `_auth_verified_claim`, roles, groups. Roles and groups stay as two
separate queries — see "Explicitly out of scope" for why merging them was rejected.

## 5. Change 4 — DispatchEventImmediately resolves twice

`pkg/lib/event/service.go:105`. `DispatchEventImmediately` resolves at line 110 and then calls
`resolveNonBlockingEvent`, which resolves the same payload again at line 277. Nothing runs
between them, so the first is pure duplication — 7 queries per call after Change 3.

Delete lines 107–113 (the three-line comment, the `Resolve` call, and its error check). Replace
with a comment pointing at `resolveNonBlockingEvent` as the single resolve for this path.

## 6. Change 5 — resolve non-blocking payloads once, except for user deletion

Covered by the `DispatchEventOnCommit` restructure in §2.5. The reasoning, which the code
comment records:

- `resolve.go` only ever resolves users. Every `resolve:` tag in the repository is
  `resolve:"user"` or `resolve:"anonymous_user"`, and both are `model.UserRef` →
  `Users.Get`. No tag targets an identity, authenticator, session, role, or group. So
  identity or authenticator deletion changes the *contents* of the resolved user model, never
  its resolvability.
- For contents, the commit-time resolve already wins today, because it overwrites. Example:
  `nonblocking/identity_loginid_removed.go:15` currently emits the user *without* the removed
  login ID, because resolve #2 replaces resolve #1. Dropping the eager resolve preserves that.
- Dropping it diverges only when `Users.Get` returns `ErrUserNotFound` at `WillCommitTx` time,
  since `resolve.go:35` `continue`s and keeps the earlier value. That needs the user row to be
  gone.
- `Store.Get` and `Store.GetByIDs` (`pkg/lib/authn/user/store.go:282`) filter on `id` only —
  no `is_anonymized` or `delete_at` predicate — so anonymisation and scheduled deletion still
  resolve normally.
- The payloads that carry a deleted user do not use `resolve:` at all.
  `UserDeletedEventPayload` (`nonblocking/user_deleted.go:13`), `UserAnonymizedEventPayload`,
  and `AdminAPIMutationDeleteUserExecutedEventPayload` embed `UserModel model.User` directly,
  with the comment *"We cannot use UserRef here because the user will be deleted BEFORE
  retrieval."*
- The one case where two distinct user IDs are in play, anonymous promotion into an existing
  account (`interaction/nodes/do_ensure_session.go:153`), deletes only the anonymous
  *identity* (`ensure_remove_anonymous_identity.go:60`), not the user.

So `len(typedPayload.DeletedUserIDs()) > 0` is currently never true for a `resolve:`-tagged
payload, and the guard is insurance rather than the mechanism that makes this correct. Because
that invariant is held by convention across ~50 payload types and a violation would fail
silently (a zero-valued `user` in a webhook), §8.2 adds a test that enforces it.

## 7. Fixed behavioural decisions

1. **`nextSeq` stays unconditional in `PrepareBlockingEventWithTx`.** Skipping it would save one
   round-trip per blocking event, but `newEvent` derives `e.ID` from the sequence
   (`pkg/lib/event/event.go:15`, `e.ID = fmt.Sprintf("%016x", seq)`), so a skipped read yields
   `ID: "0000000000000000"` on a struct still handed to `DispatchEventWithoutTx`. That is a
   footgun for any future sink that starts consuming blocking events. About 100 ms of the
   ~2800 ms target is left on the table deliberately.
2. **A skipped resolve leaves the payload's `UserModel` zero-valued.** This is unobservable:
   `hook.Sink.ReceiveBlockingEvent` re-checks `WillDeliverBlockingEvent` before delivering
   (`hook/sink.go:57`), the other three sinks return `nil` for blocking events, and
   `MakeIDTokenFromPreparationResult` / `MakeUserAccessTokenFromPreparationResult` read only
   `payload.IDToken.Payload` / `payload.JWT.Payload`.
3. **`WillDeliverBlockingEvent` is added to the `Sink` interface, not detected by type
   assertion.** An optional interface would silently skip payload population if a new sink
   forgot to implement it.
4. **The carried value is scoped to one transaction and passed explicitly.** No context value,
   no memoisation keyed by user ID, no invalidation hooks. A `UserID` mismatch falls back to
   reading.
5. **`DeriveStandardAttributesForUsers` and `AreUsersVerified` keep their signatures and their
   current queries.** Only `Queries.GetMany`, `Queries.GetPageForExport`, and the new
   `*WithClaims` variants change behaviour, so no other caller is affected.

## 8. Test plan

### 8.1 Generated files

`make generate` in the same commit as each interface change. Affected mocks:

- `pkg/lib/event/service_mock_test.go` (`Sink`, `Resolver`, `Store`)
- `pkg/lib/oauth/token_encoding_mock_test.go` (`EventService`, removed
  `AccessTokenEncodingIdentityService`)
- `pkg/lib/oauth/oidc/id_token_mock_test.go` (`IDTokenIssuerEventService`, removed
  `IDTokenIssuerIdentityService`)
- `pkg/lib/feature/verification/service_mock_test.go` (`ClaimStore` unchanged; regenerate for
  safety)

`make generate` also refreshes `wire_gen.go` for the new `UserBlockingEventContextProvider` and
the removed `Identities` fields. Both must land in the same commit as the code, or the build
breaks at that commit.

### 8.2 Unit tests

All packages below use Convey (`. "github.com/smartystreets/goconvey/convey"`), confirmed in
`pkg/lib/event/service_test.go:19` and `pkg/lib/authn/user/model_test.go:6`. Match that style.

`pkg/lib/event/service_test.go` — extend:

- `PrepareBlockingEventWithTx` with no sink returning `true` for the type: asserts
  `Resolver.Resolve` is never called, the event is still returned, and `nextSeq` still runs.
- `PrepareBlockingEventWithTx` with the hook sink returning `true`: asserts `Resolve` is
  called exactly once.
- `PrepareBlockingEventWithTx` with `opts.ResolvedUser` set: asserts `ResolveWithUser` is
  called with that user and `Resolve` is not called, regardless of the sink verdict.
- `DispatchEventOnCommit` with a blocking payload and no registered hook: asserts no resolve.
- `DispatchEventOnCommit` with a non-blocking payload whose `DeletedUserIDs()` is empty,
  followed by `WillCommitTx`: asserts `Resolve` is called exactly once, in `WillCommitTx`.
- `DispatchEventOnCommit` with a non-blocking payload whose `DeletedUserIDs()` is non-empty,
  followed by `WillCommitTx`: asserts `Resolve` is called twice.
- `DispatchEventImmediately`: asserts `Resolve` is called exactly once.

New `pkg/lib/event/resolve_test.go`:

- `Resolve` populates the `json:"user"` field from `Users.Get`.
- `Resolve` swallows `ErrUserNotFound` and leaves a previously populated field intact — this
  pins the behaviour that Change 5's guard depends on.
- `ResolveWithUser` with a matching ID does not call `Users.Get`.
- `ResolveWithUser` with a non-matching ID falls back to `Users.Get`.
- `ResolveWithUser` on `UserAnonymousPromotedEventPayload` uses the override for
  `resolve:"user"` and still calls `Users.Get` for `resolve:"anonymous_user"`.

New `pkg/api/event/resolve_tag_invariant_test.go` — enforces the §6 invariant. Iterate a
registry of every payload type in `pkg/api/event/blocking` and `pkg/api/event/nonblocking`;
for each with a `resolve:` tag on any field, assert that a zero value's `DeletedUserIDs()` is
empty for non-blocking payloads. Failure message states that a payload cannot both be resolved
and declare its user deleted, and points at `event.Service.DispatchEventOnCommit`. The registry
is a slice literal in the test file; add a companion assertion that its length matches the
number of `*.go` files in each directory so a new payload cannot be silently omitted.

`pkg/lib/feature/verification/service_test.go` — add: `AreUsersVerifiedWithClaims` returns the
same result as `AreUsersVerified` for the existing table of cases, with claims supplied
directly and `ClaimStore` expected zero times.

New `pkg/lib/feature/stdattrs/service_noevent_test.go` — the file does not exist and the
existing code carries a `TODO: Write some tests` at line 199. Add: given claims containing
`email`, `phone_number`, and an unrelated claim name,
`DeriveStandardAttributesForUsersWithClaims` sets `email_verified` and
`phone_number_verified` correctly and ignores the unrelated claim. This pins the §4.1 change
from filtered to unfiltered input.

`pkg/lib/oauth/oidc/id_token_test.go` and `pkg/lib/oauth/token_encoding_test.go` — extend the
existing tests:

- `WillDeliverBlockingEvent` false: `UserBlockingEventContexts` is never called and
  `PrepareBlockingEventWithTx` receives `PrepareBlockingEventOptions{}` with a nil
  `ResolvedUser`.
- `WillDeliverBlockingEvent` true with `opts.UserBlockingEventContext` supplied: the provider
  is not called, and the payload's `Identities` plus
  `PrepareBlockingEventOptions.ResolvedUser` come from the supplied value.
- `WillDeliverBlockingEvent` true with a supplied context whose `UserID` does not match: the
  provider is called and its result is used.
- `WillDeliverBlockingEvent` true with no supplied context: the provider is called once.

### 8.3 E2E tests

The default e2e client already sets `issue_jwt_access_token: true`
(`e2e/var/authgear.yaml:120`), so every existing test using `oauth_exchange_code` already
exercises both `oidc.id_token.pre_create` and `oidc.jwt.pre_create` preparation. Those ten
tests are the regression net for Changes 1 and 2 and must keep passing unchanged:

```
e2e/tests/webapp/login/email_password.test.yaml
e2e/tests/webapp/login/email_password_2fa.test.yaml
e2e/tests/webapp/signup/email_password.test.yaml
e2e/tests/webapp/signup/email_password_2fa.test.yaml
e2e/tests/webapp/signup/email_passwordless.test.yaml
e2e/tests/webapp/signup/email_password_bot_protection.test.yaml
e2e/tests/webapp/signup_login/new_user_email_password.test.yaml
e2e/tests/webapp/signup_login/existing_user_email_password.test.yaml
e2e/tests/webapp/select_account/login_hint_mismatch.test.yaml
e2e/tests/oauth_idp/include_identity_attributes_in_id_token.test.yaml
```

New YAML tests, following the `authgear.yaml.override` + `extra_files_directory: ./var` +
`authgeardeno:///deno/*.ts` pattern of `e2e/tests/hook/authentication.pre_initialize/`.

**`e2e/tests/hook/oidc.id_token.pre_create/carried_user_is_resolved.test.yaml`**

Registers a blocking handler for `oidc.id_token.pre_create` pointing at
`var/deno/echouserintoidtoken.ts`, which reads `e.payload.user` and mutates the ID token:

```ts
export default async function (e: any): Promise<any> {
  const u = e.payload.user;
  return {
    is_allowed: true,
    mutations: {
      id_token: {
        payload: {
          ...e.payload.id_token.payload,
          x_echo_sub: u?.id ?? "",
          x_echo_email: u?.standard_attributes?.email ?? "",
          x_echo_is_verified: String(u?.is_verified),
          x_echo_identity_count: String((e.payload.identities ?? []).length),
        },
      },
    },
  };
}
```

Steps: `oauth_setup`, a login flow for an imported user with a verified email, then
`oauth_exchange_code` asserting the decoded `id_token` contains `x_echo_sub` as a non-empty
string, `x_echo_email` equal to the imported user's email, `x_echo_is_verified` `"true"`, and
`x_echo_identity_count` `"1"`. This is the test that would fail if the carried
`UserBlockingEventContext` were empty, mismatched, or resolved with the wrong role.

**`e2e/tests/hook/oidc.jwt.pre_create/carried_user_is_resolved.test.yaml`**

`oauth_exchange_code` exposes `access_token` only as a string
(`e2e/pkg/e2eclient/client.go:326`), so assert through hook control flow instead. A deno hook
`var/deno/requireuserinjwt.ts` returns `is_allowed: false` when
`e.payload.user?.id` is empty and `is_allowed: true` otherwise. The test drives a login through
to `oauth_exchange_code` and expects success, which only happens if the payload was populated.
A companion negative case is not added: there is no way to make the payload legitimately empty
while a hook is registered, and asserting the failure branch would only test the deno hook.

**`e2e/tests/hook/oidc.id_token.pre_create/no_hook_registered.test.yaml`**

No `blocking_handlers` at all. Drives the same login and `oauth_exchange_code` and asserts a
normal `id_token` with `sub` as a string and no `x_echo_*` claims. This is the test that
catches Change 1 breaking the happy path when the resolve is skipped.

**`e2e/tests/hook/oidc.id_token.pre_create/refresh_token_grant.test.yaml`**

`issueTokensForRefreshToken` is the path the production traces captured, and it is the only
one of the four threading sites not already covered by an existing `oauth_exchange_code` test.
Uses `generate_refresh_token`, then an `http_request` to `/oauth2/token` with
`grant_type=refresh_token`, with the `echouserintoidtoken.ts` hook registered, asserting the
response contains an `id_token`. Assert the mutated claim if the runner can decode the
`id_token` from a raw `http_request` response; if not, assert `http_status: 200` and a
non-empty `id_token` field, and note in a YAML comment that claim-level assertion for this
grant is covered by the unit tests in §8.2.

**`e2e/tests/hook/user.pre_create/deleted_user_resolve.test.yaml`** — not added. No existing
flow produces a `resolve:`-tagged payload whose user is deleted in the same transaction (§6),
so there is nothing to drive from the outside. The invariant test in §8.2 covers it instead.

### 8.4 Commands

Per touched package, and as the CI-equivalent gate:

```
go test ./pkg/lib/event/... ./pkg/lib/oauth/... ./pkg/lib/authn/user/... \
        ./pkg/lib/feature/verification/... ./pkg/lib/feature/stdattrs/... \
        ./pkg/api/event/...
make lint
make test
make -C e2e run
```

Run the `review-pr` skill on the final diff and resolve every finding before marking complete.

## 9. Atomic commit plan

Each commit builds and passes `make test` on its own.

**Commit 1 — `Add WillDeliverBlockingEvent to event sinks`**

- `pkg/lib/event/service.go`: extend `Sink`, add `Service.WillDeliverBlockingEvent`.
- `pkg/lib/audit/sink.go`, `pkg/lib/search/reindex/sink.go`, `pkg/lib/userinfo/sink.go`: add
  `WillDeliverBlockingEvent` returning `false`.
- `pkg/lib/event/service_mock_test.go`: regenerated in this commit.
- No behaviour change yet; the new method has no callers.

**Commit 2 — `Add Resolver.ResolveWithUser`**

- `pkg/lib/event/resolve.go`: extract `resolve(ctx, anything, override)`, add `ResolveWithUser`.
- `pkg/lib/event/service.go`: extend the `Resolver` interface.
- `pkg/lib/event/resolve_test.go`: new, per §8.2.
- Mocks regenerated in this commit.
- No behaviour change; `Resolve` delegates with a nil override.

**Commit 3 — `Skip blocking event payload resolve when no hook is registered`**

- `pkg/api/event/prepare_blocking.go`: new, `PrepareBlockingEventOptions`.
- `pkg/lib/event/service.go`: new `PrepareBlockingEventWithTx` signature and gate; move the
  eager `Resolve` in `DispatchEventOnCommit` into the blocking branch and gate it.
- `pkg/lib/authenticationflow/dependencies.go` and the three
  `declarative/node_*.go` call sites: pass `event.PrepareBlockingEventOptions{}`.
- `pkg/lib/oauth/oidc/id_token.go`, `pkg/lib/oauth/token_encoding.go`: interface and call-site
  signature updates only, no gating at the call sites yet.
- `pkg/lib/event/service_test.go`: the blocking-path cases from §8.2.
- Mocks regenerated in this commit.

**Commit 4 — `Resolve non-blocking event payloads once unless the user is deleted`**

- `pkg/lib/event/service.go`: the `DeletedUserIDs` guard in the non-blocking branch of
  `DispatchEventOnCommit`; delete the duplicate `Resolve` in `DispatchEventImmediately`.
- `pkg/lib/event/service_test.go`: the non-blocking cases from §8.2.
- `pkg/api/event/resolve_tag_invariant_test.go`: new.
- Covers Changes 4 and 5. Kept separate from Commit 3 so a bisect can attribute any webhook
  payload regression to either the blocking or the non-blocking path.

**Commit 5 — `Carry the resolved user through the token endpoint blocking events`**

- `pkg/lib/oauth/user_blocking_event_context.go`: new.
- `pkg/lib/oauth/deps.go`: wire the provider.
- `pkg/lib/oauth/oidc/id_token.go`: drop `IDTokenIssuerIdentityService` and the `Identities`
  field, add the provider and the `PrepareIDTokenOptions` field, apply the gate/carry/fallback
  block.
- `pkg/lib/oauth/token_encoding.go`: the same for `AccessTokenEncoding` and
  `EncodeUserAccessTokenOptions`.
- `pkg/lib/oauth/grant_access_service.go`: add the field to `PrepareUserAccessGrantOptions` and
  copy it through.
- `pkg/lib/oauth/handler/handler_token.go`: `resolveUserBlockingEventContext`, the `Events`
  interface addition, the `UserBlockingEventContexts` field, and the four call-site pairs.
- Mocks and `wire_gen.go` regenerated in this commit — the removed `Identities` fields break
  the build otherwise.
- `pkg/lib/oauth/oidc/id_token_test.go`, `pkg/lib/oauth/token_encoding_test.go`: per §8.2.

**Commit 6 — `Read verified claims once in user.Queries.GetMany`**

- `pkg/lib/feature/verification/service.go`: `ListClaimsByUserIDs`,
  `AreUsersVerifiedWithClaims`, `getVerificationStatusesWithClaims`.
- `pkg/lib/feature/stdattrs/service_noevent.go`:
  `DeriveStandardAttributesForUsersWithClaims`, existing method becomes a wrapper.
- `pkg/lib/authn/user/queries.go`: interface changes, single claim read in `GetMany` and
  `GetPageForExport`.
- `pkg/lib/feature/verification/service_test.go`,
  `pkg/lib/feature/stdattrs/service_noevent_test.go`: per §8.2.
- Mocks regenerated in this commit.

**Commit 7 — `Add e2e tests for OIDC pre-create blocking event payloads`**

- The four new YAML tests and their `var/deno/*.ts` hooks from §8.3.
- No production code. Kept last so the e2e suite runs against the complete change.

**Commit 8 — `chore: Update .vettedpositions`**

- Run `make update-vettedpositions`. Only if the line-number moves in `pkg/lib/event`,
  `pkg/lib/oauth`, and `pkg/lib/authn/user` shift goanalysis output.
