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

Keep `/authflow/v2/select_account` as its own screen. **Revised from the
original decision below**: the screen-routing pre-checks that decide which
*different* screen to send the end-user to — unrelated to whether
`select_account` itself is offered — stay exactly as they are today:
anonymous `login_hint` → `/flows/promote_user`, `user_id_hint` present →
reauth/login/continue-without-interaction routing, `oauthProviderAlias` →
`/login`, request not reached from an authz endpoint → `gotoSignupOrLogin()`,
and a mismatched `login_hint` → treated as "no session" (today's exact
behavior, see [§3.1](#31-the-get-handlers-pre-checks-unchanged-in-effect)).

What **does** change: once those pre-checks pass, whether to actually render
"Continue as X" (and what account it refers to) is no longer computed by the
webapp handler independently re-deriving eligibility from
`session.GetSession(r.Context())` — it is now read directly off a real `login`
authentication flow's `identify`-step response, the exact same source of
truth a Custom UI relies on (see the spec's [Session and account
resolution](../specs/custom-ui-select-account.md#session-and-account-resolution)
and [Completing identification with the existing
session](../specs/custom-ui-select-account.md#completing-identification-with-the-existing-session)).
This closes the gap where the built-in UI maintained a second,
independently-evolving implementation of "is there an account to continue
as" alongside Part 1's `NewIdentificationOptionsSelectAccount` — the built-in
UI now dogfoods the same integration path a Custom UI is expected to use,
for both rendering *and* completing the flow (the POST `"continue"` action's
own refactor, unchanged from the original plan below).

A `login`/`signup_login` flow is required to detect eligibility (per the
spec: "the Custom UI's first flow-creation call... must be `type: login` or
`type: signup_login`... that's the only way to learn whether a session is
eligible") — so this plan **creates the `login` flow at GET time**, not only
at POST time as originally drafted. The flow persists on the webapp session
(`s.Authflow`) exactly as any other authflow-backed screen's flow does, so
the POST `"continue"` action reuses the *same* flow instance rather than
creating a second one — see [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling).

**Corrective note on scope**: an earlier draft of this adjustment considered
having the webapp force-suppress the IDP session (via a new
`AuthflowController`/`webapp.SessionOptions` plumbing addition) for the
`login_hint`-mismatch case, so the *engine* would also agree the session is
ineligible. This is unnecessary and has been dropped: a mismatched
`login_hint` already causes today's code to fall through to the exact same
`gotoSignupOrLogin()` destination as "no session at all" (`session = nil` at
line 165, folded into the `session == nil` check at line 292) — under this
plan, that pre-check simply means the flow is never created/consulted for
that request at all, exactly like the `userIDHint`/`oauthProviderAlias`/
anonymous-`login_hint` pre-checks already do for their own cases. No changes
to `pkg/auth/handler/webapp/authflow_controller.go` or
`pkg/auth/webapp/session.go` are needed for this plan.

**Second corrective note, decided after initial implementation**: an earlier
revision of this section (and of [§6](#6-backward-compatibility-risk-customized-login-flows))
argued that a *customized* login flow lacking a `select_account` entry
should still render "Continue as X" via `GetData` and complete via
`continueWithCurrentAccountLegacy`, purely for backward compatibility with
pre-refactor behavior. This is now **explicitly rejected**: if the resolved
flow does not declare `select_account`, that project does not support
account continuation — full stop, exactly what a Custom UI calling the JSON
API would see (no `select_account` option in the `identify` response at
all). There is no special-cased legacy rendering/completion path for this
case any more; both the GET and POST branches redirect to signup/login, the
same as "no session at all". `continueWithCurrentAccountLegacy` still
exists, but its only remaining caller is the pre-existing, unrelated
`userIDHint` continue-without-interaction branch — see
[§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)/[§3.5](#35-continuewithcurrentaccountlegacy--narrowed-to-the-pre-existing-useridhint-case)/[§6](#6-backward-compatibility-risk-customized-login-flows).
A project that wants to keep "Continue as X" after this ships must add
`select_account` to its own customized `login_flows`/`signup_login_flows`
config — this is now a real, disclosed migration requirement, not an
automatic carry-over.

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
3. Rewrite `AuthflowV2SelectAccountHandler`'s `ctrl.Get(...)` render branch
   (`pkg/auth/handler/webapp/authflowv2/select_account.go`) to create/resume a
   `login` flow via `*handlerwebapp.AuthflowController` and read its
   `identify`-step response to decide whether `select_account` is offered,
   instead of independently deriving eligibility from
   `session.GetSession(r.Context())`.
4. Refactor the same handler's `"continue"` POST action to advance that same
   `login` flow (submitting `{"identification": "select_account", "index":
   0}`) instead of calling `session.CreateNewAuthenticationInfoByThisSession()`
   directly — unchanged from the original plan.
5. Preserve exact current behavior for projects whose OAuth client resolves to
   a **customized** login flow that does not (yet) declare its own
   `select_account` entry — for **both** the GET render and the POST
   continuation — see [§6](#6-backward-compatibility-risk-customized-login-flows) for the fallback this requires.

Everything else in `select_account.go` (the screen-routing pre-checks,
`GetData`'s own computation, the `"login"` POST action) is unchanged — see
[§3.1](#31-the-get-handlers-pre-checks-unchanged-in-effect) for exactly what
"unchanged" means here.

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

### 3.1 The GET handler's pre-checks (unchanged in effect)

File: `pkg/auth/handler/webapp/authflowv2/select_account.go`, `ServeHTTP`'s
`ctrl.BeforeHandle` callback (lines 111-169) and `ctrl.Get(...)` body (lines
251-304). These only ever run for `GET` requests (the `NonAuthflowControllerFactory`-based
`ctrl.Get`/`ctrl.PostAction` dispatch is entirely separate from
`AuthflowController`'s own dispatch, introduced in [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)
below) — confirmed by reading `ServeHTTP`: none of these checks run for the
`"continue"`/`"login"` POST actions today, and this plan does not change
that.

Every one of the following stays **exactly as it is today, unmodified**,
because each decides which *different* screen to redirect to for a reason
that has nothing to do with whether `select_account` itself is offered — a
Custom UI never needs to make these same decisions, since it has no
equivalent "reauth screen" or "promote-anonymous-user screen" of its own:

- Anonymous `login_hint` (`loginHint.Type == oauth.LoginHintTypeAnonymous`) →
  `/flows/promote_user` (lines 253-256).
- `userIDHint != ""` → reauth (if `loginPrompt && canUseIntentReauthenticate`),
  continue-without-interaction (if session's user matches the hint and
  `!loginPrompt`), or `/login` otherwise (lines 258-278). Note the
  "continue-without-interaction" branch here still calls
  `continueWithCurrentAccount` directly, **not** the new flow-based path —
  this is intentional and unchanged: per spec, `user_id_hint` targets a
  *specific already-known* user, so there is nothing to "select", and this
  branch predates and is orthogonal to the account-chooser UI this plan
  changes.
- `oauthProviderAlias != ""` → `/login` (lines 286-289).
- Not reached from an authz endpoint (`oauthSessionID == "" && samlSessionID == ""`)
  → `gotoSignupOrLogin()` (line 291, the `!fromAuthzEndpoint` half of the
  existing `||` condition).
- A mismatched `login_hint` (lines 149-167: `loginHint != nil &&
  loginHint.Type == oauth.LoginHintTypeLoginID` and the session's user ID is
  not among `GetUserIDsByLoginIDLoginHint`'s results) sets the local
  `session` variable to `nil`. Unchanged in mechanism, but now has a
  different downstream consequence than today: previously this fed into the
  `session == nil` half of line 292's condition, which called
  `gotoSignupOrLogin()`; under this plan it means [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)'s
  render branch is never reached and the request falls through the same
  `gotoSignupOrLogin()` path — same observable destination, reached because
  the pre-check now short-circuits *before* any flow is created, not because
  the flow itself is told to treat the session as absent. **No new plumbing
  is needed for this** — see [§0](#0-decision-already-made-do-not-re-litigate)'s
  corrective note for why an earlier draft's `SuppressIDPSessionCookie`-forcing
  idea was dropped.

What is **removed** (folded into [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)'s
flow-based check instead of being re-implemented independently): the
`loginPrompt` half of line 292's condition, and the plain `session == nil`
half of it *for the purpose of deciding whether to render the page* — both
of these are exactly what `NewIdentificationOptionsSelectAccount` (Part 1)
already computes when constructing the `login` flow's `identify` step, so
there is no need for the webapp to compute them a second time before even
creating the flow. They are **not** discarded outright, though — a narrowed
form of this same check is retained as the *fallback-decision signal* for
customized flows lacking a `select_account` entry (see
[§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling),
[§6](#6-backward-compatibility-risk-customized-login-flows)).

### 3.2 `AuthflowV2Navigator.navigateStepIdentify` — new case

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
within the same `AdvanceWithInput` call ([§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling))
and this navigator path is never hit for it — but it must exist to avoid a
panic in the bot-protection-configured case, and to keep the switch
exhaustive per this file's existing convention.

### 3.3 `AuthflowV2SelectAccountHandler` — dependency wiring

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

### 3.4 `AuthflowV2SelectAccountHandler.ServeHTTP` — rewritten GET and `"continue"` handling

**Revised three times from the original draft below** — implementation
surfaced two real bugs and two wrong assumptions; see the revision notes
after each code block for what changed and why.

Current code (`ServeHTTP`) is preserved verbatim for every branch covered by
[§3.1](#31-the-get-handlers-pre-checks-unchanged-in-effect) — the
pre-checks, `gotoSignupOrLogin`/`gotoLogin`/`gotoReauth`, and the `"login"`
POST action. **Only** the `ctrl.Get(...)` body's final branch (today: call
`GetData` and render, reached once every pre-check has passed — lines
297-303) and the `"continue"` POST action's body change.

**GET render branch** — replaces lines 297-303 (as actually implemented):

```go
// session == nil / loginPrompt reflect webapp-specific reasons to distrust
// the session (suppressed, out-of-scope, or a mismatched login_hint — the
// latter deliberately NOT checked by NewIdentificationOptionsSelectAccount,
// Part 1's spec) that the resolved login flow does not know about. They
// gate here, unconditionally, before any flow is even created.
if session == nil || loginPrompt {
	gotoSignupOrLogin()
	return nil
}

// session == nil / loginPrompt: gate unconditionally before a flow is even
// created (revision note 1). The resolved login flow is created here so
// the POST "continue" action below can advance the same flow instance.
var getHandlers handlerwebapp.AuthflowControllerHandlers
getHandlers.Get(func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
	// If the flow doesn't declare select_account, this project doesn't
	// support account continuation at all (revision note 4) — same as
	// "no session", not a legacy-rendering fallback.
	if _, _, ok := selectAccountOptionFromScreen(screen); !ok {
		gotoSignupOrLogin()
		return nil
	}

	data, err := h.GetData(ctx, r, w, session.GetAuthenticationInfo().UserID)
	if err != nil {
		return err
	}
	h.Renderer.RenderHTML(w, r, TemplateWebSelectAccountHTML, data)
	return nil
})

opts := webapp.SessionOptions{
	OAuthSessionID: oauthSessionID,
	SAMLSessionID:  samlSessionID,
}
h.Controller.HandleStartOfFlow(ctx, w, r.WithContext(ctx), opts, authflow.FlowTypeLogin, &getHandlers, nil)
return nil
```

**Revision note 1 — `session == nil || loginPrompt` must gate *before* the
flow is even created, not inside a flow-based check.** An intermediate
draft put this check *inside* the `!ok` branch of a flow-based check —
reachable only when the flow didn't offer `select_account`. This crashed
live: a mismatched `login_hint` sets `session` to `nil`, but
`NewIdentificationOptionsSelectAccount` doesn't check `login_hint` (Part 1's
deliberate scope decision) and can still say `ok == true` based on the raw
session cookie — so `session.GetAuthenticationInfo()` ran on a `nil` value,
a nil pointer dereference. Confirmed live via
`e2e/tests/webapp/select_account/login_hint_mismatch.test.yaml` (added in
this work) failing with a 500 before the fix. Fixed by moving this check to
run first, unconditionally, before `selectAccountOptionFromScreen` is even
consulted (see revision note 4 for why the flow-based check itself is still
needed, just in the right place).

**Revision note 2 — `r.WithContext(ctx)`, not bare `r`, must be passed into
`HandleStartOfFlow`.** `AuthflowController.makeHTTPHandler`
(`authflow_controller.go:1064-1094`) dispatches `Get`/`PostAction` callbacks
using `r.Context()` — the plain `*http.Request`'s own context — not the
`ctx` parameter `HandleStartOfFlow` itself received. Since `HandleStartOfFlow`
is called from *inside* `ctrl.ServeWithDBTx`'s transaction (the tx lives on
`ctx`, not on `r`'s original context, because nothing before this point ever
called `r = r.WithContext(...)`), any DB read inside a `Get`/`PostAction`
callback — `GetData`'s `Identities.ListByUser`/`Users.Get` calls — panicked
with `programming_error: tx is not initialized`. This affected the
**already-existing** `e2e/tests/saml/user_authenticated_event.test.yaml`
regression test, not just new coverage — confirmed by running it against a
freshly built binary with no other changes. Fixed by passing
`r.WithContext(ctx)` into both `HandleStartOfFlow` calls (GET and POST) so
the request object `makeHTTPHandler` ultimately reads `.Context()` from
carries the transaction.

New unexported helper, in `select_account.go` (used by **both** the GET and
POST branches — final version, returning the option's real position in the
flow's `one_of` list, needed by the POST branch below):

```go
// selectAccountOptionFromScreen reports whether the login flow's current
// identify-step response offers a select_account option, and returns it
// (with its position in the options array, needed to submit a matching
// "index" — a hand-authored flow may declare select_account anywhere in
// its one_of list, not necessarily first, unlike the generated default
// flow — see revision note 4) if so. Mirrors the read pattern already used
// by every other AuthflowController-backed screen (e.g. reset_password.go's
// declarative.NewPasswordData type assertion) — returns ok == false for any
// other action type, not just a missing select_account entry, since a
// customized flow's identify step may not even be the current action.
func selectAccountOptionFromScreen(screen *webapp.AuthflowScreenWithFlowResponse) (option declarative.IdentificationOption, index int, ok bool) {
	data, ok := screen.StateTokenFlowResponse.Action.Data.(declarative.IdentificationData)
	if !ok {
		return declarative.IdentificationOption{}, 0, false
	}
	for i, o := range data.Options {
		if o.Identification == model.AuthenticationFlowIdentificationSelectAccount {
			return o, i, true
		}
	}
	return declarative.IdentificationOption{}, 0, false
}
```

`GetData` is **unchanged** — it still resolves
`session.GetAuthenticationInfo().UserID` (the same `session` this handler's
own `ctrl.BeforeHandle` already resolves, per
[§3.1](#31-the-get-handlers-pre-checks-unchanged-in-effect)) and calls
`Identities.ListByUser`/`Users.Get` exactly as today. Part 1's spec
deliberately never exposes a raw user ID in `select_account`'s option data
("No user identifier is included — the server resolves identity internally
from the input's `index`") — there is no way to derive `GetData`'s inputs
from the flow response even in principle. `selectAccountOption.DisplayName`
remains unused by this template's rendering, confirmed by reading
`resources/authgear/templates/en/web/authflowv2/select_account.html` and its
`v2.page.select-account.default.description` translation string (only
`$.UserProfile`'s structured attributes are consumed, never
`IdentityDisplayName` — that field is only used by a *different* screen's
translation keys, `v2.page.account-linking.default.by-*`).

**POST `"continue"` action** — replaces today's
`ctrl.PostAction("continue", func(ctx context.Context) error { return
continueWithCurrentAccount(ctx) })` (final version, as actually shipped):

```go
ctrl.PostAction("continue", func(ctx context.Context) error {
	var postHandlers handlerwebapp.AuthflowControllerHandlers
	postHandlers.PostAction("continue", func(ctx context.Context, s *webapp.Session, screen *webapp.AuthflowScreenWithFlowResponse) error {
		// If the flow doesn't declare select_account, this project doesn't
		// support account continuation at all (revision note 4) — redirect
		// the same way the GET branch does, rather than falling back to
		// legacy completion. Feeding {"identification":"select_account"} to
		// such a flow would be rejected by the input's JSON schema (built
		// only from the options this flow actually declares) BEFORE the
		// flow engine's ReactTo ever runs, surfacing a
		// *validation.AggregatedError, NOT authflow.ErrIncompatibleInput —
		// check upfront instead of relying on error matching (revision
		// note 3).
		//
		// index is this option's actual position in the flow's one_of
		// list, not assumed to be 0: a hand-authored flow may declare
		// select_account anywhere, unlike the generated default flow, which
		// always prepends it first (revision note 4).
		_, index, ok := selectAccountOptionFromScreen(screen)
		if !ok {
			gotoSignupOrLogin()
			return nil
		}

		result, err := h.Controller.AdvanceWithInput(ctx, r, s, screen, map[string]any{
			"identification": "select_account",
			"index":          index,
		}, nil)
		if err != nil {
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})

	opts := webapp.SessionOptions{
		OAuthSessionID: oauthSessionID,
		SAMLSessionID:  samlSessionID,
	}
	h.Controller.HandleStartOfFlow(ctx, w, r.WithContext(ctx), opts, authflow.FlowTypeLogin, &postHandlers, nil)
	return nil
})
```

**Revision note 3 — `errors.Is(err, authflow.ErrIncompatibleInput)` alone
cannot detect "customized flow lacking `select_account`".** An intermediate
draft relied solely on this check after calling `AdvanceWithInput`. It never
fires for this case: `IntentLoginFlowStepIdentify`'s `Options` (and the
per-request JSON schema `InputSchemaStepIdentify.SchemaBuilder()` builds
from them, `input_step_identify.go`) are constructed *only* from the flow's
own declared `one_of` branches. Submitting `{"identification":
"select_account", "index": 0}` against a flow lacking that branch fails
**JSON-schema validation** inside `MakeInput`, before `ReactTo` ever runs —
surfacing a `*validation.AggregatedError`, a different error type entirely.
Confirmed by tracing `doAccept` (`pkg/lib/authenticationflow/accept.go`):
a validation error on the first `Accept()` iteration propagates unchanged
through `Service.FeedInput` → `AuthflowController.feedInput` →
`AdvanceWithInputs`. The un-caught error then reached
`AuthflowController.makeHTTPHandler` → `c.renderError` →
`MakeAuthflowErrorResult`'s generic `default` branch, which redirects back
to the *current* request URL — observable as the POST redirecting back to
`/authflow/v2/select_account` itself instead of completing the SAML/OAuth
continuation. Confirmed via the pre-existing
`e2e/tests/saml/user_authenticated_event.test.yaml` regression test (its
second case's `login_flows` config originally had only `identification:
username`, no `select_account` — exactly this scenario) failing before the
fix. Fixed at the time by checking `selectAccountOptionFromScreen(screen)`
*before* calling `AdvanceWithInput`; superseded by revision note 4 below,
which removes the `errors.Is` fallback call entirely rather than keeping it
as a secondary guard.

**Revision note 4 — no fallback when the flow lacks `select_account`; use
the option's real index, not a hardcoded `0`.** Both this section and
[§6](#6-backward-compatibility-risk-customized-login-flows) originally kept
`continueWithCurrentAccountLegacy` as a fallback for customized flows
lacking `select_account`, reasoning that this preserved pre-refactor
backward compatibility. This is **explicitly rejected** (see
[§0](#0-decision-already-made-do-not-re-litigate)'s second corrective
note): if the flow doesn't declare `select_account`, that project simply
doesn't support account continuation, exactly like a Custom UI would see —
no special-cased rendering/completion path. Both GET and POST now redirect
to signup/login in that case.

This correction also surfaced a **second, independent bug**: the POST
handler previously hardcoded `"index": 0` when submitting to the flow. That
only worked because the *generated* default flow always prepends
`select_account` first (§2.1/§2.2). A hand-authored flow that declares
`select_account` anywhere else in its `one_of` list (not first) would have
this submission rejected by the same per-option `index` `Const` validation
described in revision note 3 — even though `select_account` genuinely is
present. `selectAccountOptionFromScreen` was extended to also return the
option's real index (its loop position within `Action.Data.Options`), and
the POST handler now submits that instead of a hardcoded `0`. Confirmed via
a new e2e test,
`e2e/tests/webapp/select_account/customized_flow_without_select_account.test.yaml`,
which also confirmed the "no fallback" behavior fails against the previous
implementation (redirects to `/authflow/v2/select_account` and renders
"Continue as X" instead of bouncing to `/login`) before this fix.

Confirmed mechanics for the nested-dispatch shape itself (an earlier draft
flagged this as needing empirical verification — confirmed sound, not the
source of either bug above): `makeHTTPHandler` (`authflow_controller.go:1064-1094`)
branches on `r.Method` first; for a `POST`, it reads `r.FormValue("x_action")`
(`authflow_controller.go:1072`) and looks it up in `handlers.PostHandlers`.
Both the outer, legacy `ctrl.PostAction("continue", ...)` dispatch and this
inner `postHandlers.PostAction("continue", ...)` registration read the exact
same, already-parsed `r.FormValue("x_action")` from the same request — there
is no shared mutable state between the two dispatch layers, so the inner
call reliably resolves to the intended handler on the same request that the
outer dispatch already routed here.

Because `HandleStartOfFlow` reuses `s.Authflow` when already present, the
flow created by the **GET** render branch is the *same* flow instance this
POST action advances — one flow per webapp session for this screen, created
once, exactly mirroring how `/login` itself works.

`opts.OAuthSessionID`/`SAMLSessionID` reuse the same `oauthSessionID`/
`samlSessionID` locals `ServeHTTP` already computes in its
`ctrl.BeforeHandle` callback — no new resolution logic needed, for either
branch.

On success, `result.WriteResponse(w, r)` follows the exact convention every
other `AuthflowController`-backed screen uses (see `reset_password.go`'s
`""` action, `login.go`'s `login_id` action) — it internally handles both
"flow finished → redirect to `finish_redirect_uri`" and "flow needs another
step → redirect to the next screen's URL via the navigator" (which is
exactly why [§3.2](#32-authflowv2navigatornavigatestepidentify--new-case)'s
navigator case must exist first, or this redirect could panic for a
bot-protection-gated `select_account` entry).

### 3.5 `continueWithCurrentAccountLegacy` — narrowed to the pre-existing `userIDHint` case

New unexported method, containing **exactly** today's
`continueWithCurrentAccount` body verbatim (session → `authenticationinfo.Entry`
→ `AuthenticationInfoService.Save` → `UIInfoResolver.SetAuthenticationInfoInQuery`
→ redirect). Renamed, not rewritten — this is a straight extraction, not new
logic.

**Revised from the original draft below**: this method is **not** a
fallback for "flow lacks `select_account`" any more (see
[§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)'s
revision note 4 and [§0](#0-decision-already-made-do-not-re-litigate)'s
second corrective note). Its **only** remaining caller is the pre-existing
`userIDHint` continue-without-interaction branch in `ctrl.Get(...)`
(`user_id_hint` targets a specific already-known user, so there is nothing
to "select" — this predates and is orthogonal to the account-chooser UI
this plan changes, per [§3.1](#31-the-get-handlers-pre-checks-unchanged-in-effect)).
Neither the GET render branch nor the POST `"continue"` action calls it any
more — both redirect to signup/login when the resolved flow doesn't declare
`select_account`.

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

Creating the `login` flow at **GET** time ([§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling))
does **not** itself fire this event or any other blocking event — a freshly
created flow only resolves as far as its first pending action (the
`identify` step, with no identification method chosen yet); nothing reacts
to `select_account` until the POST `"continue"` action actually submits
`{"identification": "select_account", "index": <n>}`. The event still fires at
exactly one point — flow completion — same as described above, just now
possibly on a flow instance that was created one HTTP request earlier than
where it's completed.

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

**Revised from the original decision below** — see
[§0](#0-decision-already-made-do-not-re-litigate)'s second corrective note
and [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)'s
revision note 4. This is now an intentional, disclosed **behavior change**
for such projects, not a preserved-compatibility path.

Any OAuth client that resolves (via `UIConfig.AuthenticationFlow` groups /
`AuthenticationFlowAllowlist`) to a **customized**, hand-authored
`login_flows` entry — rather than the generator's default one — will **not**
automatically gain a `select_account` `one_of` entry from this plan (customized
flows are exactly what a customer wrote; this plan does not rewrite customer
config). For such a client, feeding `{"identification": "select_account"}`
into that flow fails JSON-schema validation (no matching `one_of` branch —
see [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)'s
revision note 3 for the exact mechanism).

**Decision**: this is treated as "the project does not support
`select_account`", symmetrically at both the GET and POST layer — the same
outcome a Custom UI calling the JSON API would see for the same config (no
`select_account` option in the `identify` response at all):

- **GET (rendering)**: [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)
  redirects to signup/login instead of rendering "Continue as X" when the
  resolved flow doesn't declare `select_account` — the screen behaves
  exactly as if there were no session at all.
- **POST (completion)**: same redirect, no `continueWithCurrentAccountLegacy`
  fallback (that method's only remaining caller is the unrelated,
  pre-existing `userIDHint` branch — [§3.5](#35-continuewithcurrentaccountlegacy--narrowed-to-the-pre-existing-useridhint-case)).

This is deliberate: treating the flow config as the single, authoritative
source of truth for "is `select_account` available" — for **both** the
Custom UI integration path and the built-in UI — is simpler and more
correct than special-casing the built-in UI to preserve a pre-refactor
behavior the flow config was never consulted for in the first place. The
practical consequence: **any project with a customized `login_flows`/
`signup_login_flows` config that doesn't already declare `select_account`
loses the built-in "Continue as X" capability the moment this ships**, and
must add `select_account` to their own config to get it back (Part 1's
feature is exactly for this). This is a real, disclosed migration
requirement — flag it prominently in the PR description, not just this
plan doc.

---

## 7. File-level change plan

| File | Change |
|---|---|
| `pkg/lib/authenticationflow/declarative/generate_config_login_flow.go` | Prepend unconditional `select_account` `one_of` entry in `generateLoginFlowStepIdentify` |
| `pkg/lib/authenticationflow/declarative/generate_config_signup_login_flow.go` | Prepend `select_account` `one_of` entry (LoginFlow only) in `generateSignupLoginFlowStepIdentify` |
| `pkg/lib/authenticationflow/declarative/generate_config_login_flow_test.go` (confirm exact name) | Update golden/expected output fixtures |
| `pkg/lib/authenticationflow/declarative/generate_config_signup_login_flow_test.go` (confirm exact name) | Update golden/expected output fixtures |
| `pkg/auth/handler/webapp/authflowv2/routes.go` | Add `select_account` case to `AuthflowV2Navigator.navigateStepIdentify` |
| `pkg/auth/handler/webapp/authflowv2/select_account.go` | Rename `ControllerFactory`→`NonAuthflowControllerFactory`; add `Controller *handlerwebapp.AuthflowController`; rewrite the `ctrl.Get(...)` render branch and the `"continue"` action per §3.4 (new `selectAccountOptionFromScreen` helper); extract `continueWithCurrentAccountLegacy` per §3.5 |
| `cmd/authgear/**/wire_gen.go` (exact path TBD — grep for `AuthflowV2SelectAccountHandler` construction) | Regenerate via `make generate`, do not hand-edit |
| `pkg/auth/handler/webapp/viewmodels/preview_authflow_branch.go` | Verify whether a `select_account`-aware case is needed (§5) — likely no change, confirm during implementation |

No changes to: Part 1's engine-level files (already complete dependency),
`pkg/auth/handler/webapp/authflow_controller.go`, `pkg/auth/webapp/session.go`
(an earlier draft of this plan considered changes here for
`SuppressIDPSessionCookie` plumbing — dropped, see [§0](#0-decision-already-made-do-not-re-litigate)'s
corrective note), CORS, portal frontend, GraphQL schemas, admin API, e2e test
infrastructure (Part 1 already adds the `session_cookie` e2e capability this
part's e2e tests, §8.2, will reuse).

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
- New table test for `selectAccountOptionFromScreen` (§3.4): given a
  `*webapp.AuthflowScreenWithFlowResponse` whose `Action.Data` is a
  `declarative.IdentificationData` containing a `select_account` option →
  returns it with `ok == true`; containing no `select_account` option →
  `ok == false`; some other `Action.Data` type entirely (e.g.
  `declarative.NewPasswordData`) → `ok == false`, no panic from a failed type
  assertion.
- `AuthflowV2SelectAccountHandler` had no dedicated `_test.go` before this
  plan; `select_account_test.go` now exists, covering
  `selectAccountOptionFromScreen` (the three cases above). Full
  handler-level table tests for the POST fallback branch (simulating
  `AdvanceWithInput` returning `authflow.ErrIncompatibleInput`, or
  `selectAccountOptionFromScreen` returning `ok == false`) were not added —
  `AuthflowV2SelectAccountHandler` has many dependencies to mock and the
  e2e tests (§8.2) already exercise both fallback trigger points against a
  real server; revisit if a future change needs finer-grained coverage.

### 8.2 E2E tests

**Revised twice from the original draft below** — webapp-flow e2e coverage
(via `action: http_request` GET/POST against real webapp routes, following
redirects, asserting rendered HTML) is established practice in this repo:
`e2e/tests/webapp/login/email_password.test.yaml` is the reference pattern
for OAuth-driven webapp logins.

New files, under `e2e/tests/webapp/select_account/` (not
`oauth_setup`-based, since `SetupOAuth()` hardcodes `x_sso_enabled=false`,
which suppresses the IDP session cookie — incompatible with a test that
needs a session to persist across two authorizations; all build the
`/oauth2/authorize` request manually via `http_request_query` instead):

1. **`continue_with_current_account.test.yaml`**: a config with no
   `authentication_flow` override (so the default/generated flow, which now
   includes `select_account`, is used) with `login_id`/password auth. Logs
   in once via OAuth to establish a session, re-authorizes a second time
   while that session is still valid, confirms `/authflow/v2/select_account`
   renders "Continue as X", `POST x_action=continue`, and completes the
   OAuth exchange. Asserts via `audit_query` that `user.authenticated` fired
   twice with `continue_from_session` set on the second — proving both that
   the new engine path (not the legacy fallback) was taken, and that the
   session is reused rather than rotated (the Part 1 fix covered in that
   plan's §3.6.1).
2. **`login_hint_mismatch.test.yaml`**: establishes a session, then
   re-authorizes with a `login_hint` resolving to a *different* user.
   Confirms this redirects straight to `/login` (today's exact destination)
   rather than rendering `select_account` or crashing — this is the
   regression test for [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)'s
   revision note 1 (the nil-pointer bug), confirmed to fail with a 500
   before that fix.
3. **`customized_flow_without_select_account.test.yaml`**: a hand-authored
   `login_flows` config (`email`/`primary_password` only, no
   `select_account`). Establishes a session, then confirms a second
   authorization redirects straight to `/login` — never rendering "Continue
   as X" — and that a direct `POST x_action=continue` (simulating a stale
   client that cached the page from before the flow was reconfigured) also
   redirects rather than completing via any fallback. Asserts only one
   `user.authenticated` event total. This is the regression test for
   [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)'s
   revision note 4 — confirmed to fail (rendered "Continue as X", completed
   via the legacy path) against the implementation before that fix.

**Pre-existing test updated**: `e2e/tests/saml/user_authenticated_event.test.yaml`'s
second case originally used a hand-authored `login_flows` with only
`identification: username` (no `select_account`) to demonstrate "continuing
with current account fires `user.authenticated` with `continue_from_session`
set". Before revision note 4, this incidentally exercised the "customized
flow lacking `select_account`, legacy fallback" scenario — and was, in fact,
how both implementation bugs in an earlier revision of this section were
caught. After revision note 4, a flow lacking `select_account` no longer
completes at all, so this test's config now explicitly adds a
`select_account` `one_of` entry to keep demonstrating its original,
still-relevant intent (successful continuation via the new engine path) —
see [§3.4](#34-authflowv2selectaccounthandlerservehttp--rewritten-get-and-continue-handling)'s
revision note 4 and the "customized flow" test above for what replaced the
scenario it used to (accidentally) cover.

Original draft's 3-test plan (kept for the record): a "default flow, no
customization" case (→ test 1 above), a "customized flow without
select_account" case (→ test 3 above, added after revision note 4 — an
earlier revision of this plan thought the pre-existing SAML test already
covered this, before "no fallback" was decided), and a "login_hint mismatch"
case (→ test 2 above).

### 8.3 Manual/local verification before marking complete

- `make test` scoped to `./pkg/lib/authenticationflow/...` and any
  `./pkg/auth/handler/webapp/...` packages touched.
- `make generate` after the DI field rename (§3.3), then re-run the full
  build to confirm no stale wire references remain.
- Local manual click-through of `/authflow/v2/select_account`'s "Continue as
  X" button against a locally running server with an existing session, per
  the repo's "test the golden path in a browser" convention for
  frontend/UI-adjacent changes — this handler renders real end-user-facing
  HTML, so an automated test suite passing is not sufficient sign-off on its
  own.

---

## 9. Fixed behavioral decisions

- `/authflow/v2/select_account` remains a distinct screen; its
  screen-routing pre-checks (anonymous `login_hint`, `user_id_hint`,
  `oauthProviderAlias`, not-from-authz-endpoint) are unchanged (§0, §3.1).
- **Revised from an earlier draft**: whether to actually render "Continue as
  X" (and, separately, whether the POST `"continue"` action can complete at
  all) is no longer computed independently by the webapp from
  `session.GetSession(r.Context())` — it is read directly off a real `login`
  flow's `identify`-step response, created at **GET** time so the same flow
  instance is available for the **POST** `"continue"` action to advance
  (§3.4). `GetData` itself is unchanged (still session-based) — Part 1
  deliberately never exposes a raw user ID via `select_account`'s option
  data, so `GetData`'s inputs cannot come from the flow even in principle;
  only the *gating decision* moved.
- A mismatched `login_hint` is handled entirely as a pre-check ([§3.1](#31-the-get-handlers-pre-checks-unchanged-in-effect)) —
  it short-circuits before any `login` flow is created for the request, and
  requires **no** changes to `AuthflowController`/`webapp.Session` (an
  earlier draft's `SuppressIDPSessionCookie`-forcing plumbing idea was
  dropped as unnecessary, §0).
- Every project's *default* login/signup_login flow gains a leading
  `select_account` entry (§2). **Revised from an earlier draft**: customized
  flows that don't declare `select_account` do **not** get any
  behind-the-scenes compatibility carve-out — that project simply doesn't
  support account continuation, full stop, same as a Custom UI would see
  (§0, §6). A project that wants to keep "Continue as X" must add
  `select_account` to its own config.
- The engine-based path fires the `AuthenticationPostIdentifiedBlockingEventPayload`
  blocking event only at flow completion, never merely from creating the
  flow at GET time (§4). Since a customized flow lacking `select_account`
  now never completes via `select_account` at all (§6), this event only
  ever fires for `select_account` completions when the flow genuinely
  declares it — there is no longer an asymmetric "fires for the engine path,
  silent for the fallback" case for this scenario (only the unrelated,
  pre-existing `userIDHint` path still uses `continueWithCurrentAccountLegacy`
  silently, §3.5).
- No step-up (nested `authenticate`) is added to the generated
  `select_account` entry — the default UI's behavior stays "continue
  silently," matching today exactly. Step-up remains an opt-in customization
  (Part 1's UC3), not a default-UI feature.
- `continueWithCurrentAccountLegacy` is a verbatim extraction of existing
  logic, not a rewrite — it must produce byte-for-byte identical behavior to
  today's `continueWithCurrentAccount` for the one case it still handles
  (the pre-existing `userIDHint` continue-without-interaction branch, §3.5)
  — it is no longer reachable from either the GET or POST `select_account`
  branches.
- The POST `"continue"` action submits the `select_account` option's *real*
  index within the flow's `one_of` list (via `selectAccountOptionFromScreen`),
  never a hardcoded `0` — a hand-authored flow may declare `select_account`
  anywhere in that list, unlike the generated default flow, which always
  prepends it first (§3.4 revision note 4).

---

## 10. Implementation order

1. `generate_config_login_flow.go` + `generate_config_signup_login_flow.go` +
   updated golden tests (§2). Independently testable/mergeable — the
   generated config now includes `select_account`, but nothing in the webapp
   yet knows how to render/advance it differently from any other
   unrecognized-but-now-technically-valid identify option (harmless: the
   built-in UI doesn't call the JSON API directly, so this alone changes
   nothing user-visible yet).
2. `routes.go`'s navigator case (§3.2) — small, isolated, prevents a future
   panic once step 4 can produce `select_account` as a pending action.
3. `select_account.go`'s field rename + `Controller` addition + wire
   regeneration (§3.3) — compiles, still behaviorally identical (new field
   unused until step 4).
4. `select_account.go`'s GET-branch rewrite (`selectAccountOptionFromScreen`
   + flow-based render gating), `"continue"` rewrite, and
   `continueWithCurrentAccountLegacy` extraction (§3.4, §3.5) — the actual
   behavior change, gated by the fallback (at both layers) for safety.
5. E2E tests (§8.2) once steps 1-4 are complete.
6. `review-pr` skill pass before considering the change complete (mandatory
   per repo convention). The nested action-dispatch shape for the POST
   `"continue"` action (§3.4) has already been confirmed by reading
   `makeHTTPHandler` directly (not an open risk) — still worth a local
   manual click-through per §8.3, but not a code-review focus area on its
   own merits.

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
   - Files: `select_account.go` (GET-branch rewrite with
     `selectAccountOptionFromScreen`, `"continue"` action rewrite,
     `continueWithCurrentAccountLegacy` extraction).
   - This is the actual behavior change, at both the GET (rendering) and
     POST (completion) layer. Include the fallback for both layers (§3.4,
     §3.5) in the same commit — do not ship the new path without it, since
     that would regress customized-flow projects (§6).

5. **`chore: Update .vettedpositions`**
   - Run `make update-vettedpositions` after all line-number-affecting
     commits land.

Each commit is independently compilable; behavior only changes at commit 4,
and only for the flows that already declare (by commit 1, or by customer
choice) a matching `select_account` entry — bisect-safe by construction.
