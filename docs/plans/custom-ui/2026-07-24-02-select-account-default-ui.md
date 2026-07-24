# Implementation Plan: Reuse the Authentication Flow API in the Default (Built-in) Auth UI's Select-Account Screen

Spec: [docs/specs/custom-ui-select-account.md](../specs/custom-ui-select-account.md)
Depends on: [2026-07-24-01-select-account.md](2026-07-24-01-select-account.md) (Part 1 —
adds the `select_account` identification kind to the authentication-flow engine;
this plan cannot start before Part 1 lands).

This is "Part 3" of the 3-part breakdown (Part 2, OAuth authorize-endpoint
changes, was investigated and found to have **no gap** — `/oauth2/authorize`
already redirects into any UI, Custom or built-in, whenever `prompt != none`
regardless of session state, and `AuthenticationFlowV1CreateHandler`/
`AuthflowController` already resolve `Prompt`/`LoginHint`/`UserIDHint`/
`SuppressIDPSessionCookie` identically from the OAuth session for both the
JSON API and the webapp. Part 2 is skipped — no plan file for it).

---

## 0. Decision already made (do not re-litigate)

Keep `/authflow/v2/select_account` as its own screen, with its existing
early-exit routing (anonymous `login_hint` → promote_user, `user_id_hint` →
reauth/login, `oauthProviderAlias` → skip, no session/`prompt=login` → skip)
unchanged. **Only** the "Continue as X" button's implementation changes: from
directly minting a session/authentication-info entry
(`session.CreateNewAuthenticationInfoByThisSession()`) to creating/advancing a
`login` authentication flow and submitting `{"identification":
"select_account", "index": 0}` — the same mechanism a Custom UI uses via the
JSON API.

The alternative (fully folding account continuation into `/login`'s own
`identify` step, eliminating `/select_account` as a separate screen) was
considered and explicitly rejected as out of scope for this plan — it would
require re-homing every one of today's pre-flow early-exit branches into
flow-creation logic and is a materially larger, riskier change.

---

## 1. Goal / scope

1. Make the *default* (auto-generated) `login_flows`/`signup_login_flows`
   config include a `select_account` entry, so every project — not just those
   with hand-authored `authentication_flow` config — keeps today's "Continue
   as X" capability after this refactor.
2. Add the missing `select_account` case to `AuthflowV2Navigator.navigateStepIdentify`
   (currently panics on any unhandled `model.AuthenticationFlowIdentification`
   value — mandatory once a flow can produce this value in `Action.Identification`,
   e.g. when a `select_account` entry has `bot_protection` configured and needs
   its own page to show a challenge before submitting).
3. Refactor `AuthflowV2SelectAccountHandler`'s `"continue"` POST action
   (`pkg/auth/handler/webapp/authflowv2/select_account.go`) to create/advance a
   `login` flow via `*handlerwebapp.AuthflowController`, instead of calling
   `session.CreateNewAuthenticationInfoByThisSession()` directly.
4. Preserve exact current behavior for projects whose OAuth client resolves to
   a **customized** login flow that does not (yet) declare its own
   `select_account` entry — see [§6](#6-backward-compatibility-risk-customized-login-flows) for the fallback
   this requires.

Everything else in `select_account.go` (GET rendering, `GetData`, all
early-exit branches, the `"login"` POST action) is unchanged.

---

## 2. Config model and schema

No new config structs or schema changes beyond what Part 1 already adds
(`select_account` as a valid `identification` enum value in
`AuthenticationFlowLoginFlowIdentify`/`AuthenticationFlowSignupLoginFlowIdentify`).
This part only changes what the **generators** emit into that already-valid
shape.

### 2.1 `generate_config_login_flow.go`

File: `pkg/lib/authenticationflow/declarative/generate_config_login_flow.go`

`generateLoginFlowStepIdentify` currently builds `step.OneOf` purely by
iterating `cfg.Authentication.Identities` (`model.IdentityType*`) — and
`select_account` is not an identity type, so it's never a member of that loop.
Add an unconditional prepend, before the identity-type loop:

```go
func generateLoginFlowStepIdentify(cfg *config.AppConfig) *config.AuthenticationFlowLoginFlowStep {
	step := &config.AuthenticationFlowLoginFlowStep{
		Name: nameStepIdentify(config.AuthenticationFlowTypeLogin),
		Type: config.AuthenticationFlowLoginFlowStepTypeIdentify,
	}

	// select_account is a session-continuation mechanism, independent of
	// which identity types the project has enabled — always offer it first.
	step.OneOf = append(step.OneOf, &config.AuthenticationFlowLoginFlowOneOf{
		Identification: model.AuthenticationFlowIdentificationSelectAccount,
	})

	for _, identityType := range cfg.Authentication.Identities {
		...
	}

	return step
}
```

No nested `steps` — matches this plan's decision to keep the default UI's
existing "continue silently, no step-up" behavior (UC3-style step-up is an
opt-in config feature per Part 1, not part of the default).

### 2.2 `generate_config_signup_login_flow.go`

File: `pkg/lib/authenticationflow/declarative/generate_config_signup_login_flow.go`

The existing `newSignupLoginFlowOneOf(i model.AuthenticationFlowIdentification)`
helper sets both `SignupFlow: nameGeneratedFlow` and `LoginFlow:
nameGeneratedFlow` — reused by every other identification kind's generated
entry. `select_account` must **not** get a `SignupFlow` (per Part 1's spec:
"declares `login_flow` only, never `signup_flow`" — and Part 1's schema change
now hard-*forbids* it via `"not": {"required": ["signup_flow"]}`). Add a
dedicated construction (not reusing `newSignupLoginFlowOneOf`), prepended
before the existing identity-type loop in `generateSignupLoginFlowStepIdentify`:

```go
step.OneOf = append(step.OneOf, &config.AuthenticationFlowSignupLoginFlowOneOf{
	Identification: model.AuthenticationFlowIdentificationSelectAccount,
	LoginFlow:      nameGeneratedFlow,
	// SignupFlow intentionally left empty — select_account never signs up.
})
```

This generated `signup_login` flow's `select_account` entry points
`LoginFlow: nameGeneratedFlow` ("default") — the same generated `login_flows`
entry from §2.1, which (per §2.1) now also declares its own `select_account`
entry. This satisfies Part 1's spec requirement that "the referenced
`login_flow` must itself declare a matching `select_account` entry" for every
generated pair, automatically.

### 2.3 Generator test fixtures

`pkg/lib/authenticationflow/declarative/generate_config_login_flow_test.go`
and `generate_config_signup_login_flow_test.go` (exact names to confirm on
read — check for existing golden-output tests asserting the full generated
`AuthenticationFlowLoginFlow`/`AuthenticationFlowSignupLoginFlow` JSON/struct
for various `cfg.Authentication.Identities` combinations) will need their
expected output updated to include the new leading `select_account` entry.
Find these via `grep -rl generateLoginFlowStepIdentify` in `_test.go` files
before writing new assertions — update expected fixtures rather than
special-casing the new entry out.

---

## 3. Runtime flow

### 3.1 `AuthflowV2Navigator.navigateStepIdentify` — new case

File: `pkg/auth/handler/webapp/authflowv2/routes.go` (~line 319-389).

Confirmed: this function's `switch identification` **panics** on any
unhandled `model.AuthenticationFlowIdentification`
(`panic(fmt.Errorf("unexpected identification: %v", identification))`). Add:

```go
case model.AuthenticationFlowIdentificationSelectAccount:
	// Resume on the select-account page itself — there is no dedicated
	// per-identification screen for it, unlike oauth/passkey.
	n.NavigateSelectAccount(result)
	return
```

placed alongside the existing `case model.AuthenticationFlowIdentificationIDToken:`
/`Email`/`Phone`/`Username`/`Passkey` group (these currently share one
"redirect to the expected path with `x_step` set" branch — `select_account`
should NOT join that group, since its "expected path" is
`AuthflowV2RouteSelectAccount`, a fixed constant, not derived from the current
route the way the shared branch computes it). `NavigateSelectAccount` already
exists (used today by `AuthflowController.Restart`) — confirm its exact
signature via `grep -n "func.*NavigateSelectAccount"` before wiring, and reuse
it as-is rather than hand-rolling a new redirect.

This case is reached only when the `identify` step's action is *not yet*
resolved and needs a page to show — i.e., when a `select_account` entry has
`bot_protection` configured and the challenge hasn't been completed yet. In
the common case (no bot protection), `select_account` resolves synchronously
within the same `AdvanceWithInput` call (§3.3) and this navigator path is
never hit for it — but it must exist to avoid a panic in the
bot-protection-configured case, and to keep the switch exhaustive per this
file's existing convention.

### 3.2 `AuthflowV2SelectAccountHandler` — dependency wiring

File: `pkg/auth/handler/webapp/authflowv2/select_account.go`

Rename the existing `ControllerFactory handlerwebapp.ControllerFactory` field
to `NonAuthflowControllerFactory` (matching the established naming convention
already used by `AuthflowV2ResetPasswordHandler`,
`pkg/auth/handler/webapp/authflowv2/reset_password.go`, which mixes both
controller styles in exactly this way). Add:

```go
Controller *handlerwebapp.AuthflowController
```

All other fields (`BaseViewModel`, `Renderer`, `Users`, `UserFacade`,
`Identities`, `AuthenticationInfoService`, `UIInfoResolver`, `Cookies`,
`OAuthConfig`, `UIConfig`, `OAuthClientResolver`, `SignedUpCookie`,
`AuthenticationConfig`) are unchanged.

This requires a `wire` regeneration (`make generate`, or the repo's
`update-important-modules`/DI-specific skill if one exists — check
`cmd/authgear/**/wire_gen.go` for `AuthflowV2SelectAccountHandler`
construction) in the **same commit** as the field change, per this repo's
"never hand-edit generated files" rule.

### 3.3 `AuthflowV2SelectAccountHandler.ServeHTTP` — new `"continue"` implementation

Current code (`ServeHTTP`, `continueWithCurrentAccount`) is preserved
verbatim for every branch **except** the body of `continueWithCurrentAccount`.

New flow for the `"continue"` POST action:

```go
ctrl.PostAction("continue", func(ctx context.Context) error {
	var handlers handlerwebapp.AuthflowControllerHandlers
	handlers.PostAction("continue", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, map[string]any{
			"identification": "select_account",
			"index":          0,
		}, nil)
		if err != nil {
			if errors.Is(err, authflow.ErrIncompatibleInput) {
				// The resolved login flow does not declare a select_account
				// one_of entry (a customized, non-generated flow that
				// predates this feature). Fall back to today's behavior.
				return h.continueWithCurrentAccountLegacy(ctx)
			}
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})

	opts := webapp.SessionOptions{
		OAuthSessionID: oauthSessionID,
		SAMLSessionID:  samlSessionID,
	}
	h.Controller.HandleStartOfFlow(ctx, w, r, opts, authflow.FlowTypeLogin, &handlers, nil)
	return nil
})
```

Key mechanics, verified against `authflow_controller.go` and
`reset_password.go` (the one existing precedent for mixing both controller
styles in a single handler):

- `HandleStartOfFlow(..., input: nil)` is called on **every** request to this
  action (not just the first) — it internally calls `GetScreen` first
  (`s.Authflow != nil` check via `authflow.ErrFlowNotFound`), and only
  `createScreen`s a brand-new flow when none exists yet for this webapp
  session; otherwise it reuses the existing `s.Authflow`. This exactly mirrors
  how `/login`'s own handler calls `HandleStartOfFlow` unconditionally on
  every request.
- After resolving/creating the screen, `handleWithScreen` dispatches based on
  `r.FormValue("x_action")` — the **same** form value the outer, legacy
  `ctrl.PostAction("continue", ...)` dispatch already used to reach this code
  path. Registering the inner `handlers.PostAction("continue", ...)` with the
  identical action name is what makes this nested dispatch resolve correctly
  to the intended handler on the same incoming request — verify this via a
  request trace / manual test during implementation, since nesting two
  independent action-dispatch mechanisms on the same request is unusual
  (`reset_password.go` does not nest two controllers' dispatch within a single
  request the way this does — it picks one controller's entry point per
  request based on the presence of a `?code=` query param, not by nesting).
  If manual verification shows the nested dispatch does *not* behave as
  expected, the fallback is to have the outer legacy `PostAction("continue",
  ...)` call a lower-level helper that does the create-flow +
  `FeedInput("select_account")` sequence directly against
  `AuthenticationFlowV1WorkflowService`/`authflow.Service` (the same interface
  `AuthenticationFlowV1CreateHandler` uses), bypassing `AuthflowController`'s
  request-dispatch layer entirely and only using it (or equivalent
  bookkeeping) to persist `s.Authflow` for any follow-up step. Treat this as
  an implementation-time decision point, not a fixed design — the plan's
  intent (feed `select_account` through the real engine) is fixed; the exact
  call shape to reach it is to be confirmed against actual behavior, not
  assumed correct from static reading alone.
- `opts.OAuthSessionID`/`SAMLSessionID` reuse the same `oauthSessionID`/
  `samlSessionID` locals `ServeHTTP` already computes in its
  `ctrl.BeforeHandle` callback — no new resolution logic needed.
- On success, `result.WriteResponse(w, r)` follows the exact convention every
  other `AuthflowController`-backed screen uses (see `reset_password.go`'s
  `"" ` action, `login.go`'s `login_id` action) — it internally handles both
  "flow finished → redirect to `finish_redirect_uri`" and "flow needs another
  step → redirect to the next screen's URL via the navigator" (which is
  exactly why §3.1's navigator case must exist first, or this redirect could
  panic for a bot-protection-gated `select_account` entry).

### 3.4 `continueWithCurrentAccountLegacy` — fallback for non-generated flows

New unexported method, containing **exactly** today's
`continueWithCurrentAccount` body verbatim (session → `authenticationinfo.Entry`
→ `AuthenticationInfoService.Save` → `UIInfoResolver.SetAuthenticationInfoInQuery`
→ redirect). Renamed, not rewritten — this is a straight extraction, not new
logic. Called only when `AdvanceWithInput` fails with
`authflow.ErrIncompatibleInput` (§3.3, §6).

---

## 4. Event / delivery flow

Once `select_account` resolves through the flow engine, the exact same
`AuthenticationPostIdentifiedBlockingEventPayload` blocking event fires as it
does for any other `select_account` completion (Custom UI or otherwise) — see
Part 1 §4. This is a **behavior change** worth calling out explicitly: today,
`continueWithCurrentAccount` does **not** fire this blocking event at all (it
bypasses the flow engine entirely) — after this refactor, every built-in
"Continue as X" click will fire `AuthenticationPostIdentifiedBlockingEventPayload`
with `Identification.Identification == "select_account"`, same as a fresh
login. Any project with a blocking webhook on this event that assumed it only
fires for "real" logins (not session-continuation clicks) needs to be aware —
flag this in the PR description; not a code change, a behavior-change
disclosure.

---

## 5. Compatibility and deployment behavior

- **No storage/migration involved** — same reasoning as Part 1 §5 (ephemeral,
  Redis-backed flow state; no persisted schema).
- **Generator output changes for every project on deploy.** Unlike Part 1
  (additive, opt-in only), this part's generator change (§2) means **every**
  project using the *default* (non-customized) `login`/`signup_login` flow
  config gets a new `select_account` entry the moment this ships — this is
  the intended behavior (preserving today's UX), but it does mean the
  generated flow's `identify` step `one_of` array gains a new leading member
  for 100% of default-config deployments simultaneously with rollout, not
  gradually. No client-visible API contract changes as a result (the
  built-in Auth UI's HTML pages don't expose the raw `one_of` array), but any
  admin API / portal preview feature that visualizes the generated flow's
  branches (`pkg/auth/handler/webapp/viewmodels/preview_authflow_branch.go`,
  referenced in Part 1 §2.1) will now show this extra branch too — verify
  during implementation whether that preview code needs a
  `select_account`-aware rendering case, or whether it's acceptable for it to
  simply not model this branch (it already only handles email/phone/username
  hardcoded, not oauth/passkey/ldap either — likely fine to leave as-is, but
  confirm rather than assume).

---

## 6. Backward-compatibility risk: customized login flows

Any OAuth client that resolves (via `UIConfig.AuthenticationFlow` groups /
`AuthenticationFlowAllowlist`) to a **customized**, hand-authored
`login_flows` entry — rather than the generator's default one — will **not**
automatically gain a `select_account` `one_of` entry from this plan (customized
flows are exactly what a customer wrote; this plan does not rewrite customer
config). For such a client, feeding `{"identification": "select_account"}`
into that flow fails with `authflow.ErrIncompatibleInput` (no matching
`one_of` branch — same generic error any other unrecognized identification
kind would produce, per Part 1 §2.5's established precedent that this
mismatch is never given a dedicated, friendlier error).

**Decision**: §3.3/§3.4's fallback (`continueWithCurrentAccountLegacy`) makes
this a **non-breaking** change for such projects — they keep exactly today's
`session.CreateNewAuthenticationInfoByThisSession()`-based continuation,
silently, with no event firing (matching today exactly, unlike the new path —
see §4). This is deliberate: a customer who customized their login flow made
an explicit choice about what `identify` looks like, and this plan does not
require them to add `select_account` to keep "Continue as X" working. If a
customer wants their customized flow to ALSO fire the blocking event / go
through the "real" engine path for account continuation, they add
`select_account` to their own flow config themselves (Part 1's feature is
exactly for this) — no forced migration.

---

## 7. File-level change plan

| File | Change |
|---|---|
| `pkg/lib/authenticationflow/declarative/generate_config_login_flow.go` | Prepend unconditional `select_account` `one_of` entry in `generateLoginFlowStepIdentify` |
| `pkg/lib/authenticationflow/declarative/generate_config_signup_login_flow.go` | Prepend `select_account` `one_of` entry (LoginFlow only) in `generateSignupLoginFlowStepIdentify` |
| `pkg/lib/authenticationflow/declarative/generate_config_login_flow_test.go` (confirm exact name) | Update golden/expected output fixtures |
| `pkg/lib/authenticationflow/declarative/generate_config_signup_login_flow_test.go` (confirm exact name) | Update golden/expected output fixtures |
| `pkg/auth/handler/webapp/authflowv2/routes.go` | Add `select_account` case to `AuthflowV2Navigator.navigateStepIdentify` |
| `pkg/auth/handler/webapp/authflowv2/select_account.go` | Rename `ControllerFactory`→`NonAuthflowControllerFactory`; add `Controller *handlerwebapp.AuthflowController`; rewrite `"continue"` action per §3.3; extract `continueWithCurrentAccountLegacy` per §3.4 |
| `cmd/authgear/**/wire_gen.go` (exact path TBD — grep for `AuthflowV2SelectAccountHandler` construction) | Regenerate via `make generate`, do not hand-edit |
| `pkg/auth/handler/webapp/viewmodels/preview_authflow_branch.go` | Verify whether a `select_account`-aware case is needed (§5) — likely no change, confirm during implementation |

No changes to: Part 1's engine-level files (already complete dependency),
CORS, portal frontend, GraphQL schemas, admin API, e2e test infrastructure
(Part 1 already adds the `session_cookie` e2e capability this part's e2e
tests, §8.2, will reuse).

---

## 8. Test plan

### 8.1 Unit tests

- Extend the generator golden tests (§2.3) to assert the new leading
  `select_account` entry for both `login_flows` and `signup_login_flows`
  generation, across the existing identity-type combinations already tested.
- `AuthflowV2Navigator` test (find existing test file for `routes.go` if any;
  else confirm none exists and decide whether to add one) — assert
  `navigateStepIdentify` no longer panics for `select_account` and routes to
  `AuthflowV2RouteSelectAccount`.
- `AuthflowV2SelectAccountHandler` currently appears to have no dedicated
  `_test.go` (confirm on read) — if none exists, this is not a regression
  introduced by this plan, but strongly consider adding a table test for the
  fallback branch (§3.4): simulate `AdvanceWithInput` returning
  `authflow.ErrIncompatibleInput` and assert `continueWithCurrentAccountLegacy`
  runs instead of propagating the error.

### 8.2 E2E tests

Requires Part 1's e2e `session_cookie` infra (already planned there) plus the
webapp test harness's ability to follow redirects and inspect rendered HTML —
confirm via `write-e2e-test` skill and existing `e2e/tests/webapp/` examples
whether webapp-flow e2e coverage is already established practice in this repo
before writing browser-level assertions; if not, scope this to:

1. **Default flow, no customization**: injected session → `GET
   /authflow/v2/select_account` → renders "Continue as X" → `POST
   x_action=continue` → redirects to `finish_redirect_uri` /
   `/oauth2/consent`. Assert (via a DB/audit query step, matching this repo's
   e2e conventions) that the `select_account`-tagged authentication event was
   recorded — proving the new path (not the legacy fallback) was taken for a
   default/generated flow.
2. **Customized flow without `select_account`**: same setup, but
   `authentication_flow.login_flows` overridden with a hand-authored flow
   lacking any `select_account` entry → `POST x_action=continue` still
   succeeds (fallback engaged) → redirect behavior identical to test 1's
   final redirect, proving no regression.

### 8.3 Manual/local verification before marking complete

- `make test` scoped to `./pkg/lib/authenticationflow/...` and any
  `./pkg/auth/handler/webapp/...` packages touched.
- `make generate` after the DI field rename (§3.2), then re-run the full
  build to confirm no stale wire references remain.
- Local manual click-through of `/authflow/v2/select_account`'s "Continue as
  X" button against a locally running server with an existing session, per
  the repo's "test the golden path in a browser" convention for
  frontend/UI-adjacent changes — this handler renders real end-user-facing
  HTML, so an automated test suite passing is not sufficient sign-off on its
  own.

---

## 9. Fixed behavioral decisions

- `/authflow/v2/select_account` remains a distinct screen; its early-exit
  routing is unchanged (§0).
- Every project's *default* login/signup_login flow gains a leading
  `select_account` entry (§2); customized flows do not, and are not migrated
  automatically (§6).
- The new path fires the `AuthenticationPostIdentifiedBlockingEventPayload`
  blocking event; the legacy fallback path does not — this asymmetry is
  intentional and matches each path's own prior behavior exactly (§4, §6).
- No step-up (nested `authenticate`) is added to the generated
  `select_account` entry — the default UI's behavior stays "continue
  silently," matching today exactly. Step-up remains an opt-in customization
  (Part 1's UC3), not a default-UI feature.
- `continueWithCurrentAccountLegacy` is a verbatim extraction of existing
  logic, not a rewrite — it must produce byte-for-byte identical behavior to
  today's `continueWithCurrentAccount` for the cases it now exclusively
  handles.

---

## 10. Implementation order

1. `generate_config_login_flow.go` + `generate_config_signup_login_flow.go` +
   updated golden tests (§2). Independently testable/mergeable — the
   generated config now includes `select_account`, but nothing in the webapp
   yet knows how to render/advance it differently from any other
   unrecognized-but-now-technically-valid identify option (harmless: the
   built-in UI doesn't call the JSON API directly, so this alone changes
   nothing user-visible yet).
2. `routes.go`'s navigator case (§3.1) — small, isolated, prevents a future
   panic once step 3 can produce `select_account` as a pending action.
3. `select_account.go`'s field rename + `Controller` addition + wire
   regeneration (§3.2) — compiles, still behaviorally identical (new field
   unused until step 4).
4. `select_account.go`'s `"continue"` rewrite + `continueWithCurrentAccountLegacy`
   extraction (§3.3, §3.4) — the actual behavior change, gated by the
   fallback for safety.
5. E2E tests (§8.2) once steps 1-4 are complete.
6. `review-pr` skill pass before considering the change complete (mandatory
   per repo convention) — pay particular attention to the nested
   action-dispatch mechanism flagged in §3.3 as needing empirical
   verification, not just code review.

---

## 11. Atomic commit plan

1. **`Include select_account in the default login and signup_login flows`**
   - Files: `generate_config_login_flow.go`, `generate_config_signup_login_flow.go`,
     their `_test.go` golden-fixture updates.
   - Depends on: Part 1 fully merged (needs `model.AuthenticationFlowIdentificationSelectAccount`
     and the config schema enum to exist).

2. **`Add select_account case to the authflow v2 navigator`**
   - Files: `pkg/auth/handler/webapp/authflowv2/routes.go`.
   - Independently safe: adds a case that's unreachable until commit 4.

3. **`Wire AuthflowController into the select-account handler`**
   - Files: `select_account.go` (field rename + addition only, no behavior
     change), regenerated `wire_gen.go`.
   - Both `NonAuthflowControllerFactory` and `Controller` fields present but
     `Controller` unused — compiles, behaviorally inert.

4. **`Reuse the authentication flow engine for select-account continuation`**
   - Files: `select_account.go` (`"continue"` action rewrite,
     `continueWithCurrentAccountLegacy` extraction).
   - This is the actual behavior change. Include the fallback (§3.4) in the
     same commit — do not ship the new path without it, since that would
     regress customized-flow projects (§6).

5. **`chore: Update .vettedpositions`**
   - Run `make update-vettedpositions` after all line-number-affecting
     commits land.

Each commit is independently compilable; behavior only changes at commit 4,
and only for the flows that already declare (by commit 1, or by customer
choice) a matching `select_account` entry — bisect-safe by construction.
