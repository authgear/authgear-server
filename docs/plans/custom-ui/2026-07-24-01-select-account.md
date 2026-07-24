# Implementation Plan: Custom UI Select Account

Spec: [docs/specs/custom-ui-select-account.md](../specs/custom-ui-select-account.md)

This plan adds a new `select_account` identification kind to the `identify` step of
`login` and `signup_login` authentication flows. No new endpoints, step types, or
config top-level concepts are introduced — this is entirely additive within the
existing `identify` step's `one_of` mechanism.

---

## 1. Goal / scope

Implement:

- A new `model.AuthenticationFlowIdentification` value `select_account`.
- Config support for `select_account` in `authentication_flow.login_flows` and
  `authentication_flow.signup_login_flows` `identify` steps only (not `signup_flows`,
  `reauth_flows`, `promote_flows`, `account_recovery_flows`).
- Runtime resolution of the option (session eligibility rules — scoped to
  session presence, `suppress_idp_session_cookie`, and `prompt=login` only;
  `login_hint`/`id_token_hint` filtering is explicitly deferred, see
  [§3.1](#31-session-eligibility--shared-option-constructor) and
  [§9](#9-fixed-behavioral-decisions)) and of the input (session-still-valid
  re-check, immediate completion or switch into a named `login_flow`).
- The `SelectAccountSessionChanged` 401 error.
- E2E test infrastructure to inject an IDP session cookie into `create`/`input`
  authentication-flow API calls (does not exist today — see [§7](#7-e2e-test-infrastructure-gap)).

Explicitly out of scope, already implemented, no changes needed anywhere:

- CORS allow-listing of `x_custom_ui_uri` origins for
  `/api/v1/authentication_flows` and its `states/input` endpoint — already covers
  every OAuth client's `x_custom_ui_uri` (`pkg/lib/infra/middleware/cors_matcher.go`,
  `pkg/lib/infra/middleware/cors.go`) and already applies to these routes via
  `apiChain`/`authenticationFlowChain` in `pkg/auth/routes.go`.
- Portal UI — there is no structured portal form for `authentication_flow.*_flows`
  content (verified: no `login_flows`/`signup_login_flows` references under
  `portal/src`); it is edited as raw YAML. No portal changes.

In scope overall, but deferred to the companion plan
[2026-07-24-02-select-account-default-ui.md](2026-07-24-02-select-account-default-ui.md)
(Part 3), which depends on this plan's engine-level work landing first — **not**
skipped, just sequenced into the other plan file:

- Default/auto-generated flow config (`GenerateLoginFlowConfig`,
  `GenerateSignupLoginFlowConfig` in `pkg/lib/authenticationflow/declarative/generate_config_*.go`)
  currently iterates only `cfg.Authentication.Identities` (login_id/oauth/passkey/ldap),
  which `select_account` is not a member of, so the generators need an
  unconditional addition to give the *default* login/signup_login flows a
  `select_account` entry too (Part 3 §2). This plan (Part 1) only makes
  `select_account` a valid, hand-authorable config value — it does not by
  itself change what the generators emit.
- The built-in Auth UI (`pkg/auth/handler/webapp/authflowv2/select_account.go`
  and friends) — reusing this plan's engine work there is Part 3's subject
  matter in full.

---

## 2. Config model and schema

File: `pkg/lib/config/authentication_flow.go`. No new Go structs — `select_account`
reuses the existing `AuthenticationFlowLoginFlowOneOf` and
`AuthenticationFlowSignupLoginFlowOneOf` shapes.

### 2.1 `model.AuthenticationFlowIdentification` (new constant)

File: `pkg/api/model/identification.go`

```go
const (
    ...
    AuthenticationFlowIdentificationLDAP          AuthenticationFlowIdentification = "ldap"
    AuthenticationFlowIdentificationSelectAccount AuthenticationFlowIdentification = "select_account"
)
```

Add a `case AuthenticationFlowIdentificationSelectAccount: return nil` arm to both
`PrimaryAuthentications()` and `SecondaryAuthentications()` (mirroring `OAuth`/
`Passkey`/`LDAP` treatment — `select_account` never requires primary or secondary
authentication by default; step-up is opt-in via nested `steps`, see [§2.3](#23-nested-steps-uc3)).
This is required to keep the exhaustive switch from panicking — both methods
`panic(fmt.Errorf("unknown identification: %v", m))` on an unhandled case, and
while today's only callers (`preview_authflow_branch.go`,
`generate_config_{login,signup}_flow.go`) invoke these with hardcoded
non-`select_account` values, the switch must stay exhaustive per existing
convention.

### 2.2 JSON schema enum additions

Three `Schema.Add(...)` string literals in `pkg/lib/config/authentication_flow.go`
gain `"select_account"` in their `identification` enum array:

- `AuthenticationFlowLoginFlowIdentify` (~line 314-338)
- `AuthenticationFlowSignupLoginFlowIdentify` (~line 418-440)

Do **not** add it to `AuthenticationFlowSignupFlowIdentify` (~159),
`AuthenticationFlowReauthFlowIdentify` (~510), or
`AuthenticationFlowAccountRecoveryIdentification` (~645) — matches [Scope](../specs/custom-ui-select-account.md#scope)
in the spec exactly.

### 2.3 Nested `steps` (UC3)

`AuthenticationFlowLoginFlowOneOf` already has a generic `Steps
[]*AuthenticationFlowLoginFlowStep` field and the `AuthenticationFlowLoginFlowIdentify`
schema already has a generic `"steps"` property (~line 332-335) available to every
`identification` value. No struct or schema change needed for UC3 — a
`select_account` entry with a nested `authenticate` step for `secondary_totp` is
already representable once `select_account` is a valid enum value.

`AuthenticationFlowSignupLoginFlowOneOf.GetSteps()` returns `nil`
unconditionally (line ~1136) and its schema has no `steps` property — this is
correct and unchanged: per spec, `signup_login`'s `select_account` never has its
own nested steps; step-up is configured on the *target* `login_flow`'s own
`select_account` entry.

### 2.4 `signup_login`'s `select_account` has no `signup_flow` (UC2/UC4)

Today, `AuthenticationFlowSignupLoginFlowIdentify`'s schema unconditionally
requires `["identification", "signup_flow", "login_flow"]` for every entry
regardless of `identification` value (confirmed: there is no existing
conditional exemption in the JSON Schema; the only existing passkey exemption is
in `generate_config_signup_login_flow.go`'s Go-authored struct construction,
which bypasses schema validation entirely, not a schema-level exemption). Since
the spec's UC2/UC4 examples omit `signup_flow` for `select_account` in
hand-authored YAML, the schema itself must become conditional. Change:

```json
var _ = Schema.Add("AuthenticationFlowSignupLoginFlowIdentify", `
{
	"type": "object",
	"required": ["identification"],
	"properties": {
		"identification": {
			"type": "string",
			"enum": [
				"email",
				"phone",
				"username",
				"oauth",
				"passkey",
				"ldap",
				"id_token",
				"select_account"
			]
		},
		"bot_protection": { "$ref": "#/$defs/AuthenticationFlowBotProtection" },
		"signup_flow": { "$ref": "#/$defs/AuthenticationFlowObjectName" },
		"login_flow": { "$ref": "#/$defs/AuthenticationFlowObjectName" }
	},
	"allOf": [
		{
			"if": {
				"properties": { "identification": { "const": "select_account" } }
			},
			"then": {
				"required": ["login_flow"],
				"not": { "required": ["signup_flow"] }
			},
			"else": {
				"required": ["signup_flow", "login_flow"]
			}
		}
	]
}
`)
```

This moves `signup_flow`/`login_flow` out of the top-level `required` array and
into a conditional `allOf`/`if`/`then`/`else` (same technique already used
elsewhere in this file, e.g. `AuthenticationFlowLoginFlowStep`'s `type`-conditioned
`allOf`). The `"not": {"required": ["signup_flow"]}` branch enforces the spec's
"never `signup_flow`" wording as a hard config error, not just an omission
allowance.

No Go struct change: `AuthenticationFlowSignupLoginFlowOneOf.SignupFlow` stays
`omitempty` and is simply left unset for `select_account` entries.

### 2.5 Config-load-time validation — none added

There is no existing config-load-time check that a `signup_flow`/`login_flow`
name resolves to a real, defined flow (confirmed:
`flowRootObjectForLoginFlow`/`flowRootObjectForSignupFlow` in
`pkg/lib/authenticationflow/declarative/utils_common.go` raise `ErrFlowNotFound`
lazily at runtime). The spec's requirement that "the referenced `login_flow`
must itself declare a matching `select_account` `one_of` entry" is *not* given a
dedicated config-time validator — this matches how every other identification
kind's `login_flow`/`signup_flow` reference is already unvalidated at load time.
A misconfigured target (no matching `select_account` entry) surfaces the same
way a misconfigured `email` target would: `authflow.ErrIncompatibleInput` when
the replayed synthetic input doesn't match any `one_of` entry in the target
flow. This is a deliberate non-change, consistent with existing behavior.

### 2.6 Config test data

`pkg/lib/config/testdata/authentication_flow_type_identify_tests.yaml` only
asserts the JSON-schema-error *kind* (`"enum"`) for an unrecognized
`identification` value, not the literal list of allowed values — no update
needed. No other test fixture enumerates the full identification list (verified
via repo-wide search for `"oauth"` + `"id_token"`/`"ldap"` co-occurrence in
`pkg/lib/config/testdata/`).

---

## 3. Runtime flow

All new/changed files are in `pkg/lib/authenticationflow/declarative/` unless
stated otherwise.

### 3.1 Session eligibility — shared option constructor

New function in `data_identification.go`, returning **parallel slices** rather
than a single option, even though at most one entry is produced today — this
is the forward-compatible shape the user asked for: when multiple concurrent
accounts are supported later, only this function's *body* grows a real
enumeration loop; its signature, and every caller of it, stay unchanged:

```go
func NewIdentificationOptionsSelectAccount(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	authflowCfg *config.AuthenticationFlowBotProtection,
	appCfg *config.BotProtectionConfig,
) (options []IdentificationOption, userIDs []string, err error)
```

`options[i]` and `userIDs[i]` always refer to the same account —
`options[i]` is what's returned to the client, `userIDs[i]` is the
server-only value recorded for the session-freshness re-check ([§3.6](#36-shared-session-freshness-check)).
Today, `len(options) == len(userIDs)` is always `0` or `1`.

Call sequence (mirrors the eligibility rules already implemented for the
built-in Auth UI's account-selection screen,
`pkg/auth/handler/webapp/authflowv2/select_account.go`, adapted from the webapp
`Session`/cookie-resolved-session pair to the Authentication Flow API's
`authflow.Session` (query-derived OIDC params, `pkg/lib/authenticationflow/session.go`)
and `session.GetSession(ctx)` (cookie-resolved IDP/offline-grant session,
`pkg/lib/session/context.go`) — **both already populated in the request context
for `POST /api/v1/authentication_flows`**, since `newAllSessionMiddleware` runs
inside `apiChain` before the handler, per `pkg/auth/routes.go`):

1. `sess := session.GetSession(ctx)` (import `"github.com/authgear/authgear-server/pkg/lib/session"`,
   unaliased — this package's exported name is `session`; do not confuse with
   `authflow.GetSession(ctx)`, a different type from `pkg/lib/authenticationflow`).
   If `sess == nil` → return `(nil, nil, nil)`.
2. If `authflow.GetSuppressIDPSessionCookie(ctx)` → return `(nil, nil, nil)`.
3. If `slice.ContainsString(authflow.GetSession(ctx).Prompt, "login")` → return
   `(nil, nil, nil)`. (`authflow.GetSession(ctx).Prompt` already has
   `max_age`-implied `"login"` folded in server-side by
   `oauth.PromptResolver.ResolvePrompt` well before the flow is created — no
   separate `max_age` check needed here, matching the spec's edge case table.)
4. `userID := sess.GetAuthenticationInfo().UserID`.
5. `identities, err := deps.Identities.ListByUser(ctx, userID)` (if `err != nil`,
   return `(nil, nil, err)`).
6. `displayName := selectAccountDisplayName(identities)` — new unexported
   helper in `data_identification.go`, functionally identical to
   `pkg/auth/handler/webapp.IdentitiesDisplayName` (same
   `identitiesDisplayNamePriorities` logic: prefer `LoginID` over `OAuth`
   identity, `DisplayID()`/provider-type formatting). Duplicated rather than
   imported because `pkg/lib/authenticationflow/declarative` cannot import
   `pkg/auth/handler/webapp` (that package already imports
   `pkg/lib/authenticationflow`; importing it back would cycle).
7. Return
   `([]IdentificationOption{{Identification:
   model.AuthenticationFlowIdentificationSelectAccount, BotProtection:
   GetBotProtectionData(flows, authflowCfg, appCfg), DisplayName: displayName}},
   []string{userID}, nil)` — one-element slices; this single `return` statement
   is the only place that would grow into a real loop if multiple concurrent
   accounts are ever supported.

**Deliberate deviation from the spec, scoped for this plan**: the spec's
[Session and account resolution](../specs/custom-ui-select-account.md#session-and-account-resolution)
section additionally omits the option when `login_hint`/`id_token_hint`
identify a *different* user than the session. This plan's first iteration
**does not implement that check** — the option is shown whenever a usable
session exists, regardless of `login_hint`/`id_token_hint`, per explicit
instruction to ignore them for now ("the options always show all logged in
accounts"). `authflow.GetUserIDHint(ctx)`/`authflow.GetLoginHint(ctx)` are
therefore **not read** by this function at all in this plan. This is a
tracked simplification, not an oversight — see [§9](#9-fixed-behavioral-decisions)
and treat re-adding the hint-mismatch checks as follow-up work, not something
silently dropped forever; the spec document itself is not being amended by
this plan, only this implementation's first cut is narrower than it.

`IdentificationOption` (in `data_identification.go`) gains one field:

```go
type IdentificationOption struct {
    ...
    // DisplayName is specific to SelectAccount. Unmasked — see spec's
    // "Unlike masked_display_name elsewhere" note.
    DisplayName string `json:"display_name,omitempty"`
}
```

This field must never carry a user ID or other unmasked-but-sensitive value —
`display_name` is the only `select_account`-specific field returned to the
client, matching the spec's HTTP API section exactly.

### 3.2 Option construction call sites

Both `NewIntentLoginFlowStepIdentify` (`intent_login_flow_step_identify.go`)
and `NewIntentSignupLoginFlowStepIdentify` (`intent_signup_login_flow_step_identify.go`)
gain, in their `switch b.Identification` option-building loop:

```go
case model.AuthenticationFlowIdentificationSelectAccount:
    selectAccountOptions, selectAccountUserIDs, err := NewIdentificationOptionsSelectAccount(ctx, deps, flows, b.BotProtection, deps.Config.BotProtection)
    if err != nil {
        return nil, err
    }
    for idx, opt := range selectAccountOptions {
        if i.SelectAccountUserIDs == nil {
            i.SelectAccountUserIDs = map[int]string{}
        }
        // Keyed by this option's position in the full `options` slice being
        // built (the same position the client will later echo back as
        // "index" — see §3.3), not by its position within
        // selectAccountOptions. This is what the client's "index" input
        // actually looks up at dispatch time (§3.4).
        i.SelectAccountUserIDs[len(options)] = selectAccountUserIDs[idx]
        options = append(options, opt)
    }
```

Both `IntentLoginFlowStepIdentify` and `IntentSignupLoginFlowStepIdentify`
structs gain one new field:

```go
// SelectAccountUserIDs maps a position in this intent's Options slice (the
// same position the client echoes back as the identify input's "index"
// field, §3.3) to the user ID resolved for that select_account entry when
// the option was constructed. Never serialized to the API response
// (OutputData only marshals i.Options, not the whole intent) — looked up
// again at dispatch time (§3.4) and re-checked against the current session
// on submission to detect a session change between option-construction and
// input (see resolveSelectAccountSession, §3.6).
//
// Keyed by index (not a single field) precisely so that if multiple
// concurrent accounts are supported later, only NewIdentificationOptionsSelectAccount's
// body needs to enumerate more than one entry — the dispatch/lookup code in
// §3.4/§3.9/§3.10 already handles an arbitrary number of entries today.
SelectAccountUserIDs map[int]string `json:"select_account_user_ids,omitempty"`
```

### 3.3 Input schema: `index` field on the identify step (§ HTTP API changes)

File: `input_step_identify.go`.

1. Change the option-building loop from `for _, option := range i.Options` to
   `for index, option := range i.Options` (needed to `Const`-validate the
   index).
2. Add a new `case model.AuthenticationFlowIdentificationSelectAccount:` in
   `InputSchemaStepIdentify.SchemaBuilder()`'s switch:
   ```go
   case model.AuthenticationFlowIdentificationSelectAccount:
       required = append(required, "index")
       b.Properties().Property("index", validation.SchemaBuilder{}.Type(validation.TypeInteger).Const(index))
       setRequiredAndAppendOneOf()
   ```
   (bot-protection requirement is already handled generically above the
   switch, unchanged.)
3. Add a field to `InputStepIdentify`:
   ```go
   Index int `json:"index,omitempty"`
   ```
   and implement the getter the dispatch code in [§3.4](#34-how-index-is-used-for-dispatch) needs:
   ```go
   var _ inputTakeIdentificationOptionIndex = &InputStepIdentify{}

   func (i *InputStepIdentify) GetIdentificationOptionIndex() int {
       return i.Index
   }
   ```
   (`inputTakeIdentificationOptionIndex` is defined once, in [§3.7](#37-new-shared-input-schema-inputschematakeidentificationoptionindex),
   and implemented by both `InputStepIdentify` here and the new
   `InputTakeIdentificationOptionIndex` type in §3.7 — two different concrete
   types satisfying the same interface, matching how `inputTakeIdentificationMethod`
   is already implemented by `InputStepIdentify` alongside several other
   `inputTakeXxx` interfaces.)

### 3.4 How `index` is used for dispatch

`checkIdentificationMethod(deps, step, im)` (in both
`IntentLoginFlowStepIdentify` and `IntentSignupLoginFlowStepIdentify`) still
resolves the **config** `one_of` position (used for `JSONPointerForOneOf`, to
locate nested `steps`) purely by matching `Identification` — unaffected and
unchanged, because there is still at most one `select_account` config entry
per `identify` step, so identification-based matching is already unambiguous
for it (unlike `oauth`, where multiple *options* can share
`Identification == oauth` from one `one_of` entry and need `alias` to
disambiguate).

`index`, however, **is** used — to resolve *which recorded user ID* a
`select_account` submission refers to, via `i.SelectAccountUserIDs` (§3.2), but
**only within the flow whose own identify step produced that index**.
`IntentSignupLoginFlowStepIdentify.ReactTo` (§3.9) uses it exactly this way —
a real client call is always dispatched against the same flow instance that
built the map, so the index is unambiguous there. `IntentLoginFlowStepIdentify.ReactTo`
(§3.10) does the same for a **direct** real client call, but must also accept
a **second** input shape: a `SyntheticInputSelectAccount` replayed from a
`signup_login` switch, which carries the already-verified user ID directly
instead of an index (§3.8) — because that index would have been computed
against the *source* flow's options, not this (target) flow's own,
independently-computed ones. Carrying an index across that specific boundary
was the original design and was wrong; see §3.8's revision note for why, and
`e2e/tests/select_account/signup_login_switch_index_mismatch.test.yaml` for
the regression test that would fail without the fix.

This is also why `checkIdentificationMethod`'s *config* `one_of` lookup and
the `index`-based *options* lookup must not be confused with each other: the
former answers "which config branch (and therefore which nested `steps`, if
any) applies", the latter answers "which specific account was chosen" — two
different questions that happen to coincide (always resolve to the same
single config branch) only because there's exactly one `select_account`
config entry per `identify` step. Multiple *options* from that one entry
(future) would still resolve to the *same* config branch via
`checkIdentificationMethod`, disambiguated only by `index`/`SelectAccountUserIDs`
— exactly parallel to how `oauth`'s multiple provider options resolve to the
same-or-different config branches disambiguated by `alias` instead. This
per-flow scoping is exactly why a *cross-flow* index (§3.8) never made sense
in the first place — `index` was never meant to be a globally meaningful
identifier, only a same-flow-instance one.

### 3.5 Login flow: `IntentUseIdentitySelectAccount` + `NodeDoUseIdentitySelectAccount`

New file: `intent_use_identity_select_account.go`. Mirrors
`intent_identify_with_id_token.go`'s `IntentIdentifyWithIDToken` shape exactly
— `select_account`, like `id_token`, resolves to a user without resolving to a
specific `identity.Info`, so it implements `MilestoneFlowUseIdentity` returning
whatever `MilestoneDoUseIdentity` is already in scope (none), never producing
one itself:

```go
type IntentUseIdentitySelectAccount struct {
	JSONPointer     jsonpointer.T                          `json:"json_pointer,omitempty"`
	Identification  model.AuthenticationFlowIdentification `json:"identification,omitempty"`
	ExpectedUserID  string                                 `json:"expected_user_id,omitempty"`
}

var _ authflow.Intent = &IntentUseIdentitySelectAccount{}
var _ authflow.Milestone = &IntentUseIdentitySelectAccount{}
var _ MilestoneIdentificationMethod = &IntentUseIdentitySelectAccount{}
var _ MilestoneFlowUseIdentity = &IntentUseIdentitySelectAccount{}
var _ authflow.InputReactor = &IntentUseIdentitySelectAccount{}

func (*IntentUseIdentitySelectAccount) Kind() string { return "IntentUseIdentitySelectAccount" }
func (*IntentUseIdentitySelectAccount) Milestone() {}
func (n *IntentUseIdentitySelectAccount) MilestoneIdentificationMethod() model.AuthenticationFlowIdentification {
	return n.Identification
}
func (*IntentUseIdentitySelectAccount) MilestoneFlowUseIdentity(flows authflow.Flows) (MilestoneDoUseIdentity, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoUseIdentity](flows)
}
```

`CanReactTo`: **must** check `authflow.FindMilestoneInCurrentFlow[MilestoneDoUseUser](flows)` first and
return `(nil, authflow.ErrEOF)` if already satisfied — exactly like
`IntentIdentifyWithIDToken.CanReactTo` does — before building the schema. Omitting
this check is a real bug, not a hypothetical one: once `NodeDoUseIdentitySelectAccount`
(and then `NodePostIdentified`) finish reacting, `FindInputReactorForFlow` falls
through to asking `IntentUseIdentitySelectAccount.CanReactTo` again; without the
guard it happily returns the same schema again, `ReactTo` fires again, and the
`doAccept` loop only stops at `MAX_LOOP=100` by panicking — which then gets
retried indefinitely by the caller, hanging the request. Confirmed by
implementation: this exact omission caused every e2e test touching this code
path to hang the server in a tight loop until fixed.

```go
func (n *IntentUseIdentitySelectAccount) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, userIdentified := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseUser](flows)
	if userIdentified {
		return nil, authflow.ErrEOF
	}

	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}
	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, n.JSONPointer, n)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeIdentificationOptionIndex{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
		BotProtectionCfg:        deps.Config.BotProtection,
	}, nil
}
```

`IntentLookupIdentitySelectAccount` (signup_login flow, [§3.8](#38-signup_login-flow-intentlookupidentityselectaccount))
does **not** need this guard — its `ReactTo` always ends in either an error
(`ErrIncompatibleInput`) or `errors.Join(bpSpecialErr, &authflow.ErrorSwitchFlow{...})`,
never in a normal completed node, so it can never be found and re-entered the
way `IntentUseIdentitySelectAccount` can.

`ReactTo`:
```go
func (n *IntentUseIdentitySelectAccount) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputTakeIdentificationOptionIndex inputTakeIdentificationOptionIndex
	if !authflow.AsInput(input, &inputTakeIdentificationOptionIndex) {
		return nil, authflow.ErrIncompatibleInput
	}

	var bpSpecialErr error
	bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
	if err != nil {
		return nil, err
	}

	result, err := NewNodeDoUseIdentitySelectAccount(ctx, deps, flows, n.ExpectedUserID)
	if err != nil {
		return nil, err
	}
	return result, bpSpecialErr
}
```

New file: `node_do_use_identity_select_account.go`. Mirrors
`node_do_use_id_token.go`'s `NodeDoUseIDToken`:

```go
type NodeDoUseIdentitySelectAccount struct {
	UserID string `json:"user_id,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseIdentitySelectAccount{}
var _ authflow.Milestone = &NodeDoUseIdentitySelectAccount{}
var _ MilestoneDoUseUser = &NodeDoUseIdentitySelectAccount{}
var _ authflow.InputReactor = &NodeDoUseIdentitySelectAccount{}

func (n *NodeDoUseIdentitySelectAccount) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return nil, nil
}

func (n *NodeDoUseIdentitySelectAccount) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	return NewNodePostIdentified(ctx, deps, flows, &NodePostIdentifiedOptions{
		Identification: model.Identification{
			Identification: model.AuthenticationFlowIdentificationSelectAccount,
			Identity:       nil,
			IDToken:        nil,
		},
	})
}

func NewNodeDoUseIdentitySelectAccount(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, expectedUserID string) (authflow.ReactToResult, error) {
	userID, err := resolveSelectAccountSession(ctx, expectedUserID)
	if err != nil {
		return nil, err
	}
	n := &NodeDoUseIdentitySelectAccount{UserID: userID}
	return authflow.NewNodeSimple(n), nil
}

func (*NodeDoUseIdentitySelectAccount) Kind() string { return "NodeDoUseIdentitySelectAccount" }
func (*NodeDoUseIdentitySelectAccount) Milestone() {}
func (n *NodeDoUseIdentitySelectAccount) MilestoneDoUseUser() string { return n.UserID }
```

### 3.6 Shared session-freshness check

New unexported function, placed in `node_do_use_identity_select_account.go`
(only two call sites, both introduced by this plan — see [§3.8](#38-signup_login-flow-intentlookupidentityselectaccount)):

```go
func resolveSelectAccountSession(ctx context.Context, expectedUserID string) (userID string, err error) {
	sess := session.GetSession(ctx)
	if sess == nil {
		return "", ErrSelectAccountSessionChanged
	}
	userID = sess.GetAuthenticationInfo().UserID
	if userID != expectedUserID {
		return "", ErrSelectAccountSessionChanged
	}
	return userID, nil
}
```

This implements the spec's "the session cookie must still resolve to the same
user recorded when the option was computed" check. It is called both when
`select_account` completes directly in a `login` flow, and — independently,
against the `expectedUserID` resolved from that flow's own
`SelectAccountUserIDs[index]` (§3.4) — when `signup_login` is about to switch
into the target `login_flow` (redundant-but-harmless double check within the
same HTTP request when nothing changed in between; see [§3.8](#38-signup_login-flow-intentlookupidentityselectaccount)).

### 3.7 New shared input schema: `InputSchemaTakeIdentificationOptionIndex`

New file: `input_take_identification_option_index.go`. Mirrors
`input_take_login_id.go`'s `InputSchemaTakeLoginID`/`InputTakeLoginID` shape:

```go
type InputSchemaTakeIdentificationOptionIndex struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	IsBotProtectionRequired bool
	BotProtectionCfg        *config.BotProtectionConfig
}

var _ authflow.InputSchema = &InputSchemaTakeIdentificationOptionIndex{}

func (i *InputSchemaTakeIdentificationOptionIndex) GetJSONPointer() jsonpointer.T { return i.JSONPointer }
func (i *InputSchemaTakeIdentificationOptionIndex) GetFlowRootObject() config.AuthenticationFlowObject { return i.FlowRootObject }

func (i *InputSchemaTakeIdentificationOptionIndex) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.Type(validation.TypeObject).Required("index")
	b.Properties().Property("index", validation.SchemaBuilder{}.Type(validation.TypeInteger))
	if i.IsBotProtectionRequired && i.BotProtectionCfg != nil {
		b = AddBotProtectionToExistingSchemaBuilder(b, i.BotProtectionCfg)
	}
	return b
}

func (i *InputSchemaTakeIdentificationOptionIndex) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeIdentificationOptionIndex
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeIdentificationOptionIndex struct {
	Index         int                         `json:"index,omitempty"`
	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeIdentificationOptionIndex{}
var _ inputTakeIdentificationOptionIndex = &InputTakeIdentificationOptionIndex{}
var _ inputTakeBotProtection = &InputTakeIdentificationOptionIndex{}

func (*InputTakeIdentificationOptionIndex) Input() {}
func (i *InputTakeIdentificationOptionIndex) GetIdentificationOptionIndex() int { return i.Index }
func (i *InputTakeIdentificationOptionIndex) GetBotProtectionProvider() *InputTakeBotProtectionBody { return i.BotProtection }
func (i *InputTakeIdentificationOptionIndex) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil { return "" }
	return i.BotProtection.Type
}
func (i *InputTakeIdentificationOptionIndex) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil { return "" }
	return i.BotProtection.Response
}
```

New interface in `input_interface.go` (alongside
`inputTakeAccountRecoveryDestinationOptionIndex`/`inputTakeAuthenticationOptionIndex`):

```go
type inputTakeIdentificationOptionIndex interface {
	GetIdentificationOptionIndex() int
}
```

This is a distinct HTTP-cascade step: the client's single `{"identification":
"select_account", "index": 0}` call is first parsed by
`InputSchemaStepIdentify` (top-level `identify` step, produces `*InputStepIdentify`),
which dispatches into the `select_account` sub-flow; the engine then re-parses
the *same* raw JSON body against the sub-flow's own `CanReactTo`-declared
schema (`InputSchemaTakeIdentificationOptionIndex`, producing a *different* Go
value, `*InputTakeIdentificationOptionIndex`) before calling `ReactTo` — this
two-level re-parse is the same mechanism `IntentLookupIdentityLoginID`/
`IntentLookupIdentityPasskey` already rely on (confirmed by reading those
files: `assertion_response`/`login_id` are required by both the outer
`InputSchemaStepIdentify` and the inner `InputSchemaTakeLoginID`/
`InputSchemaTakePasskeyAssertionResponse`), all within one client HTTP call —
no second round trip is introduced.

### 3.8 Signup_login flow: `IntentLookupIdentitySelectAccount`

**Revised from the original design below** — the first implementation carried
the client's array-position `index` unchanged into the target `login_flow`'s
synthetic input. This is wrong: the target flow computes its **own**,
independent `Options`/`SelectAccountUserIDs` from its **own** `one_of` config,
so `select_account` can legitimately sit at a different position there than
in the source `signup_login` flow's options (different identification lists,
different provider/server counts before it, etc.). `AcceptSyntheticInput`
feeds the synthetic input through *without* re-validating it against the
target's schema, so a position mismatch silently produced
`ErrIncompatibleInput` for an otherwise perfectly valid, session-backed
continuation. Caught by e2e-testing a config where the two flows disagree on
position (not by static reading) — see
`e2e/tests/select_account/signup_login_switch_index_mismatch.test.yaml`.

The fix follows the existing precedent set by `SyntheticInputPasskey`/
`SyntheticInputOAuth` (`synthetic_input_passkey.go`, `synthetic_input_oauth.go`):
those carry the actual resolved value (an assertion response, an identity
spec) across a flow switch, never a position. `select_account`'s synthetic
input must do the same — carry the already-verified user ID directly, since
that is the one thing that is meaningful independent of either flow's option
layout.

New file: `synthetic_input_select_account.go`, mirroring `SyntheticInputPasskey`'s
shape exactly (including carrying `BotProtection` forward for interface-shape
parity with sibling synthetic input types, even though — matching those same
siblings — it is not actually populated when constructed, since bot
protection is verified once, at the source flow, against the real client
input):

```go
type SyntheticInputSelectAccount struct {
	Identification model.AuthenticationFlowIdentification `json:"identification,omitempty"`
	UserID         string                                 `json:"user_id,omitempty"`
	BotProtection  *InputTakeBotProtectionBody             `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &SyntheticInputSelectAccount{}
var _ inputTakeIdentificationMethod = &SyntheticInputSelectAccount{}
var _ inputTakeSelectAccountUserID = &SyntheticInputSelectAccount{}
var _ inputTakeBotProtection = &SyntheticInputSelectAccount{}

func (*SyntheticInputSelectAccount) Input() {}
func (i *SyntheticInputSelectAccount) GetIdentificationMethod() model.AuthenticationFlowIdentification {
	return i.Identification
}
func (i *SyntheticInputSelectAccount) GetSelectAccountUserID() string { return i.UserID }
// GetBotProtectionProvider/Type/Response mirror SyntheticInputPasskey's, omitted here.
```

New interface in `input_interface.go`:

```go
type inputTakeSelectAccountUserID interface {
	GetSelectAccountUserID() string
}
```

`intent_lookup_identity_select_account.go` no longer stores a `SyntheticInput`
field at all — it builds the synthetic input fresh in `ReactTo`, from the
`userID` that `resolveSelectAccountSession` just verified (which, on success,
equals `n.ExpectedUserID`):

```go
type IntentLookupIdentitySelectAccount struct {
	JSONPointer    jsonpointer.T                          `json:"json_pointer,omitempty"`
	Identification model.AuthenticationFlowIdentification `json:"identification,omitempty"`
	ExpectedUserID string                                 `json:"expected_user_id,omitempty"`
}

var _ authflow.Intent = &IntentLookupIdentitySelectAccount{}
var _ authflow.Milestone = &IntentLookupIdentitySelectAccount{}
var _ MilestoneIdentificationMethod = &IntentLookupIdentitySelectAccount{}
var _ authflow.InputReactor = &IntentLookupIdentitySelectAccount{}

func (*IntentLookupIdentitySelectAccount) Kind() string { return "IntentLookupIdentitySelectAccount" }
func (*IntentLookupIdentitySelectAccount) Milestone() {}
func (n *IntentLookupIdentitySelectAccount) MilestoneIdentificationMethod() model.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *IntentLookupIdentitySelectAccount) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}
	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, n.JSONPointer, n)
	if err != nil {
		return nil, err
	}
	return &InputSchemaTakeIdentificationOptionIndex{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
		BotProtectionCfg:        deps.Config.BotProtection,
	}, nil
}

func (n *IntentLookupIdentitySelectAccount) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}
	oneOf := n.oneOf(current)

	var inputTakeIdentificationOptionIndex inputTakeIdentificationOptionIndex
	if authflow.AsInput(input, &inputTakeIdentificationOptionIndex) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
		if err != nil {
			return nil, err
		}

		userID, err := resolveSelectAccountSession(ctx, n.ExpectedUserID)
		if err != nil {
			return nil, err
		}

		// select_account never switches to signup: it can only ever
		// continue an existing login. The synthetic input carries the
		// already-verified user ID directly, not this flow's option index.
		return nil, errors.Join(bpSpecialErr, &authflow.ErrorSwitchFlow{
			FlowReference: authflow.FlowReference{
				Type: authflow.FlowTypeLogin,
				Name: oneOf.LoginFlow,
			},
			SyntheticInput: &SyntheticInputSelectAccount{
				Identification: n.Identification,
				UserID:         userID,
			},
		})
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentLookupIdentitySelectAccount) oneOf(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupLoginFlowOneOf {
	oneOf, ok := o.(*config.AuthenticationFlowSignupLoginFlowOneOf)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}
	return oneOf
}
```

Note the explicit `resolveSelectAccountSession` re-check here even though the
switched-to `login_flow`'s own `IntentUseIdentitySelectAccount` will
independently re-check against `n.ExpectedUserID` again — this source-side
check exists so a session that changed between "option shown" and "input
submitted" fails with the precise `SelectAccountSessionChanged` error at the
point closest to where the mismatch is known, rather than deferring to the
target flow, where a changed/absent session could otherwise silently omit the
option entirely and fail with a generic `ErrIncompatibleInput` instead of the
documented 401.

### 3.9 Wiring into `IntentSignupLoginFlowStepIdentify.ReactTo`

The client-supplied `index` (§3.3/§3.4) is used **only** to look up
`i.SelectAccountUserIDs` in *this* (source) flow — it never travels any
further. `idx` (the existing `checkIdentificationMethod`-resolved **config**
`one_of` position) is still used for `JSONPointerForOneOf`, unchanged. In the
existing dispatch switch:

```go
case model.AuthenticationFlowIdentificationSelectAccount:
    var inputTakeIdentificationOptionIndex inputTakeIdentificationOptionIndex
    if !authflow.AsInput(input, &inputTakeIdentificationOptionIndex) {
        return nil, authflow.ErrIncompatibleInput
    }
    optionsIndex := inputTakeIdentificationOptionIndex.GetIdentificationOptionIndex()
    expectedUserID, ok := i.SelectAccountUserIDs[optionsIndex]
    if !ok {
        return nil, authflow.ErrIncompatibleInput
    }

    return authflow.NewSubFlow(&IntentLookupIdentitySelectAccount{
        JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
        Identification: identification,
        ExpectedUserID: expectedUserID,
    }), nil
```

This case no longer touches the shared `syntheticInput := &InputStepIdentify{Identification: identification}`
variable that every other case in this switch builds — `select_account`
doesn't need it; `IntentLookupIdentitySelectAccount` builds its own
`SyntheticInputSelectAccount` internally (§3.8) once the switch actually
fires.

### 3.10 Wiring into `IntentLoginFlowStepIdentify.ReactTo`

This is the one dispatch site that must accept **two** different input
shapes, because it's reachable both directly (a real client call against
this `login` flow's own identify step) and as the target of a `signup_login`
switch (a `SyntheticInputSelectAccount` carrying a user ID, not an index):

```go
case model.AuthenticationFlowIdentificationSelectAccount:
    var expectedUserID string
    var inputTakeSelectAccountUserID inputTakeSelectAccountUserID
    if authflow.AsInput(input, &inputTakeSelectAccountUserID) {
        // Replayed via a signup_login switch: the source flow already
        // resolved and verified this exact user ID — this flow's own
        // option positions never come into it.
        expectedUserID = inputTakeSelectAccountUserID.GetSelectAccountUserID()
    } else {
        var inputTakeIdentificationOptionIndex inputTakeIdentificationOptionIndex
        if !authflow.AsInput(input, &inputTakeIdentificationOptionIndex) {
            return nil, authflow.ErrIncompatibleInput
        }
        optionsIndex := inputTakeIdentificationOptionIndex.GetIdentificationOptionIndex()
        var ok bool
        expectedUserID, ok = i.SelectAccountUserIDs[optionsIndex]
        if !ok {
            return nil, authflow.ErrIncompatibleInput
        }
    }

    return authflow.NewSubFlow(&IntentUseIdentitySelectAccount{
        JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
        Identification: identification,
        ExpectedUserID: expectedUserID,
    }), nil
```

`IntentUseIdentitySelectAccount.ReactTo`'s own input-shape gate must accept
either shape too, for the same reason (it receives the same input object on
the next cascade iteration): `authflow.AsInput(input,
&inputTakeIdentificationOptionIndex) || authflow.AsInput(input,
&inputTakeSelectAccountUserID)` — neither branch's extracted value is used at
that level (`n.ExpectedUserID` is already resolved by the caller above), the
check only gates that *some* select_account-shaped input arrived.

### 3.11 Error definition

File: `pkg/lib/authenticationflow/declarative/error.go`, add:

```go
var ErrSelectAccountSessionChanged = apierrors.Unauthorized.
	WithReason("SelectAccountSessionChanged").
	New("session no longer matches the selected account")
```

Matches the spec's exact response body:
```json
{ "error": { "name": "Unauthorized", "reason": "SelectAccountSessionChanged", "message": "session no longer matches the selected account", "code": 401 } }
```

### 3.12 `target_step` error message (no change, noted)

`utils_common.go`'s `findIdentity` helper returns `InvalidTargetStep.NewWithInfo("the
referenced target_step does not associate with an identity, perhaps the taken
branch is an ID token.", ...)` when a later `authenticate` step's `target_step`
points at a `select_account`-completed identify step (no `MilestoneDoUseIdentity`,
same as `id_token`). This message literally says "ID token" and will now also
apply to `select_account`, but this is a pre-existing generic fallback message,
not part of this feature's scope — not changed by this plan.

---

## 4. Event / delivery flow

`NodeDoUseIdentitySelectAccount.ReactTo` calls the existing
`NewNodePostIdentified` (`node_post_identified.go`), which prepares/dispatches
the existing `AuthenticationPostIdentifiedBlockingEventPayload` blocking event.
No changes to the event payload shape — `model.Identification.Identification`
will now also carry the value `"select_account"` for this event, same as every
other identification kind. No webhook/event schema doc found enumerating
identification values that would need updating (repo search found none).

---

## 5. Compatibility and deployment behavior

- **Backward compatibility**: `select_account` is a new enum value; any
  existing `login`/`signup_login` config without it is schema-unaffected and
  behavior-unchanged (per spec's own Backward Compatibility section). The one
  schema *tightening* is `AuthenticationFlowSignupLoginFlowIdentify`'s
  `required` restructuring (§2.4) — this loosens (not tightens) requirements
  for existing configs, since `signup_flow`+`login_flow` are still required
  for every non-`select_account` identification (the `else` branch), identical
  to today. No existing config can newly fail validation.
- **No storage/migration involved.** Flow state is ephemeral (Redis-backed,
  keyed by `state_token`); no persisted schema changes, no Redis key format
  changes, no rollout ordering concerns. A rolling deploy where old and new
  binaries briefly coexist is safe: old binaries simply don't recognize
  `select_account` as a valid `identification` (schema rejects it) until they
  are replaced; no in-flight flow state references a `select_account` node
  kind created by a newer binary in a way an older binary would need to read
  (blue/green or rolling deploy of this server component doesn't share
  in-flight flow state across binary versions in this repo's deployment
  model).
- **No dual-read/dual-write needed.**

---

## 6. File-level change plan

| File | Change |
|---|---|
| `pkg/api/model/identification.go` | Add `AuthenticationFlowIdentificationSelectAccount` const; add case to `PrimaryAuthentications()`/`SecondaryAuthentications()` |
| `pkg/lib/config/authentication_flow.go` | Add `"select_account"` to `AuthenticationFlowLoginFlowIdentify` and `AuthenticationFlowSignupLoginFlowIdentify` enums; restructure `AuthenticationFlowSignupLoginFlowIdentify`'s `required` into a conditional `allOf` (§2.4) |
| `pkg/lib/authenticationflow/declarative/data_identification.go` | Add `DisplayName` field to `IdentificationOption`; add `NewIdentificationOptionsSelectAccount` (slice-returning); add unexported `selectAccountDisplayName` helper |
| `pkg/lib/authenticationflow/declarative/input_step_identify.go` | Loop-index capture; add `select_account` case to `SchemaBuilder()`; add `Index` field + `GetIdentificationOptionIndex()` to `InputStepIdentify` |
| `pkg/lib/authenticationflow/declarative/input_take_identification_option_index.go` | **New.** `InputSchemaTakeIdentificationOptionIndex` / `InputTakeIdentificationOptionIndex` |
| `pkg/lib/authenticationflow/declarative/input_interface.go` | Add `inputTakeIdentificationOptionIndex` and `inputTakeSelectAccountUserID` interfaces |
| `pkg/lib/authenticationflow/declarative/synthetic_input_select_account.go` | **New.** `SyntheticInputSelectAccount` — carries the verified user ID (not an index) across a signup_login switch, mirroring `SyntheticInputPasskey`/`SyntheticInputOAuth` (§3.8) |
| `pkg/lib/authenticationflow/declarative/intent_login_flow_step_identify.go` | Add option-construction case; add `SelectAccountUserIDs map[int]string` field; add `ReactTo` dispatch case accepting either a real index-carrying input or a synthetic user-ID-carrying input |
| `pkg/lib/authenticationflow/declarative/intent_signup_login_flow_step_identify.go` | Add option-construction case; add `SelectAccountUserIDs map[int]string` field; add index-lookup `ReactTo` dispatch case (index is only ever used within this flow, never carried onward) |
| `pkg/lib/authenticationflow/declarative/intent_use_identity_select_account.go` | **New.** `IntentUseIdentitySelectAccount` (login flow) |
| `pkg/lib/authenticationflow/declarative/node_do_use_identity_select_account.go` | **New.** `NodeDoUseIdentitySelectAccount`, `NewNodeDoUseIdentitySelectAccount`, `resolveSelectAccountSession` |
| `pkg/lib/authenticationflow/declarative/intent_lookup_identity_select_account.go` | **New.** `IntentLookupIdentitySelectAccount` (signup_login flow) — builds `SyntheticInputSelectAccount` for the flow switch |
| `pkg/lib/authenticationflow/declarative/error.go` | Add `ErrSelectAccountSessionChanged` |
| `e2e/pkg/testrunner/models.go` | Add `session_cookie` property/field to the `Step` schema/struct (§7) |
| `e2e/pkg/testrunner/testcase.go` | Call `client.InjectSession(...)` in `StepActionCreate` and `StepActionInput` when `step.SessionCookie != nil` (§7) |
| `e2e/tests/select_account/*.test.yaml` | **New.** E2E tests (§8.2) |
| Various `_test.go` files in `pkg/lib/authenticationflow/declarative/` | Unit tests (§8.1) |

No changes to: CORS middleware, portal frontend, authui frontend (Custom UI is
a customer-hosted, non-Authgear-frontend concept — no authui changes),
`generate_config_*.go`, GraphQL schemas, admin API.

---

## 7. E2E test infrastructure gap

Verified: `action: create` / `action: input` steps in the E2E YAML format
(`e2e/pkg/testrunner/models.go`, `Step` struct) have **no** way to attach an
IDP session cookie to the request today. `session_cookie` (referencing the
existing `SessionCookie` schema, `models.go` ~line 639) currently only exists
for `saml_request_session_cookie` and `http_request_session_cookie`, wired via
`client.InjectSession(idpSessionID, idpSessionToken)`
(`e2e/pkg/e2eclient/client.go:405`) at `testcase.go` lines ~339 and ~458. Since
`select_account` fundamentally depends on an existing session being visible to
`POST /api/v1/authentication_flows`, this must be added:

1. `e2e/pkg/testrunner/models.go`:
   - Add `"session_cookie": { "$ref": "#/$defs/SessionCookie" }` to the `Step`
     JSON schema's `properties` (alongside `http_request_session_cookie`).
   - Add `SessionCookie *SessionCookie \`json:"session_cookie"\`` to the `Step`
     struct, next to `HTTPRequestSessionCookie`.
2. `e2e/pkg/testrunner/testcase.go`:
   - In `case StepActionCreate:` (~line 161), before `client.CreateFlow(input)`:
     ```go
     if step.SessionCookie != nil {
         client.InjectSession(step.SessionCookie.IDPSessionID, step.SessionCookie.IDPSessionToken)
     }
     ```
   - In `case StepActionInput: fallthrough case "":` (~line 544), before
     `client.InputFlow(...)`, the same guard.

This reuses `client.InjectSession` exactly as `http_request`/`saml_request`
already do — `client` is a single `*authflowclient.Client` per test case with
one shared `CookieJar`, so injecting a session once makes it visible to every
subsequent HTTP call from that client, including `CreateFlow`/`InputFlow`
calls that don't themselves reference `SessionCookie`. Injecting a *different*
session before a later `input` step (or omitting `session_cookie` to simulate
the cookie disappearing) is how the `SelectAccountSessionChanged` test case
(§8.2) is exercised.

The underlying session row itself is created via the existing `before: -
type: create_session` hook (`BeforeHookCreateSession`, already implemented,
already used by `e2e/tests/promote/promote_with_session.test.yaml`) — no
changes needed there.

---

## 8. Test plan

### 8.1 Unit tests (Go, Convey BDD style — matches this package's existing convention, e.g. `input_step_identify_test.go`)

- `data_identification_test.go` (new or extend if a file for this already
  exists under a different name — check first): table-test
  `NewIdentificationOptionsSelectAccount` for each omission rule this plan
  actually implements: no session, `SuppressIDPSessionCookie`, `prompt`
  contains `login`. Also assert `login_hint`/`id_token_hint` (`UserIDHint`) are
  **not** consulted at all — e.g. a session present alongside a mismatched
  `login_hint`/`id_token_hint` in the context still yields the option (locks
  in the §3.1 deviation so it isn't silently reintroduced by accident).
  Requires constructing a minimal `authflow.Dependencies` with a fake
  `Identities` and a `context.Context` carrying a fake `session.ResolvedSession`
  via `session.WithSession` plus the relevant `authflow.Session.MakeContext`
  fields — follow whatever fake/mock pattern this package's existing tests use
  (check `node_check_login_hint_test.go` if present, else the nearest analog
  under this package, before inventing a new mock style).
- `input_step_identify_test.go`: extend existing `TestInputSchemaStepIdentify`
  with a `select_account` option case, asserting the generated JSON schema
  requires `index` with the correct `const`.
- `input_take_identification_option_index_test.go` (new): schema-builder test
  mirroring `input_take_login_id_test.go`'s shape, if one exists — else a
  minimal schema-round-trip test.
- New test for `resolveSelectAccountSession`: matching user → no error;
  mismatched user → `ErrSelectAccountSessionChanged`; nil session →
  `ErrSelectAccountSessionChanged`.
- `pkg/lib/config`: extend `authentication_flow_test.go` (or wherever
  `AuthenticationFlowConfig` schema validation is tested) with a case
  asserting a `signup_login_flows` `select_account` entry without
  `signup_flow` passes validation, and one *with* `signup_flow` fails
  validation (the new `not: {required: [signup_flow]}` branch).

Before writing, use the `add-go-test` skill to confirm exact fake/mock
conventions already used in this package (this package appears to construct
`*authflow.Dependencies` by hand per test rather than via a generated mock
framework — confirm before adding new test helpers).

### 8.2 E2E tests (YAML, `e2e/tests/select_account/`)

New directory `e2e/tests/select_account/`, following the `write-e2e-test`
skill's format. All tests use `before: - type: create_session` to seed a
known IDP session (mirroring `e2e/tests/promote/promote_with_session.test.yaml`),
then a `session_cookie` on the `create`/`input` step (§7) to present it.

1. **`login_direct.test.yaml`** (UC1 skeleton, without the two-client OAuth
   plumbing, which is out of e2e scope): `login_flows: default` with
   `one_of: [select_account, email→primary_password]`. With an injected
   session: `create` returns `identify` with an `options` array containing
   `{"identification": "select_account", "display_name": "[[string]]"}`.
   `input {"identification": "select_account", "index": 0}` → `finished`.
2. **`login_no_session.test.yaml`**: same config, no `before: create_session`
   → `create`'s `options` array does not contain a `select_account` entry
   (assert via `"[[arrayof]]"` absence or exact array match without it).
3. **`login_prompt_login.test.yaml`**: injected session, `url_query` includes
   `prompt=login` → option omitted.
4. **`login_step_up_totp.test.yaml`** (UC3): `select_account` with nested
   `authenticate: secondary_totp`. Injected session for a user with TOTP
   enrolled → `input select_account` → `action.type == "authenticate"`,
   `authentication == "secondary_totp"`; submit TOTP code → `finished`.
5. **`login_session_changed.test.yaml`**: `create` with session A injected
   (option shown) → re-inject session B (or no session) via a second
   `session_cookie` on the `input` step → `input select_account` → `error:
   {"reason": "SelectAccountSessionChanged", "code": 401}`.
6. **`signup_login_switch.test.yaml`** (UC2/UC4 skeleton):
   `signup_login_flows: default` with `select_account` (`login_flow:
   default_login`, no `signup_flow`) + `email` (`signup_flow`/`login_flow`);
   `login_flows: default_login` with matching `select_account` entry +
   `terminate_other_sessions` after `identify`. Injected session → `create`
   signup_login → `input select_account` → single response shows
   `"type": "login"`, `"name": "default_login"`,
   `"action": {"type": "terminate_other_sessions"}` (mirrors the cascading
   assertion style already used in `e2e/tests/signup_login/login.test.yaml`).
7. **`signup_login_decline.test.yaml`** (UC2): injected session, `identify`
   options include `select_account`; `input` with `{"identification":
   "email", "login_id": "new-account@example.com"}` instead (not
   `select_account`) → switches to `signup` as normal, unaffected by
   `select_account`'s presence.

### 8.3 Manual/local verification before marking complete

- `make test` scoped to `./pkg/lib/config/...` and
  `./pkg/lib/authenticationflow/...` (per repo convention: run the narrowest
  relevant test first).
- `cd e2e && make teardown && make setup`, then
  `go test ./pkg/testrunner/ -count 1 -v -timeout 10m -run
  "TestAuthflow/select_account"`.
- `make export-schemas` is optional/debug-only (confirmed: output goes to
  `tmp/`, not a checked-in artifact) — not required, but running it locally is
  a reasonable sanity check that the schema changes still produce valid JSON
  Schema.

---

## 9. Fixed behavioral decisions

(To avoid re-litigating during implementation — these are settled, not open
questions.)

- `select_account` never requires primary or secondary authentication by
  default (`PrimaryAuthentications()`/`SecondaryAuthentications()` return
  `nil`), matching `oauth`/`passkey`/`ldap`. Step-up is opt-in per-entry via
  nested `steps`, never a shared mechanism.
- `select_account` **does** support the generic `bot_protection` config field
  (unlike `id_token`, which explicitly opts out because it "inherently does
  not support user interaction") — `select_account` is a real button click by
  a real user, so bot protection is a meaningful, if optional, control.
- No `MilestoneDoUseIdentity` is produced — `select_account` resolves to a
  user (`MilestoneDoUseUser`), not to a specific `identity.Info`, exactly like
  `id_token`. A later step's `target_step` pointing back at a
  `select_account`-completed identify step gets the same (slightly
  ID-token-worded) `InvalidTargetStep` error as it would for `id_token` today
  — not fixed by this plan.
- The session re-check (`resolveSelectAccountSession`) runs at **both** the
  `signup_login` switch point and the final `login`-flow resolution point,
  independently, each against its own flow instance's own recorded
  `SelectAccountUserIDs` entry. This is intentional redundancy, not a bug —
  see §3.8's rationale.
- No config-load-time validation is added for "does the referenced
  `login_flow` actually declare a matching `select_account` entry" — consistent
  with how every other identification kind's flow references are (not)
  validated today.
- `index` in the input **is** used for server-side dispatch (§3.4) — it looks
  up `i.SelectAccountUserIDs[index]` to resolve which recorded user ID a
  submission refers to, but **only within the flow that produced that
  index**. `index` is never carried across a `signup_login` → `login_flow`
  switch — `IntentLookupIdentitySelectAccount` resolves the user ID itself
  (already verified by `resolveSelectAccountSession`) and passes it onward via
  `SyntheticInputSelectAccount.UserID`, not via `index` (§3.8). This is
  deliberately real, working lookup code, not a validated-but-ignored
  placeholder, specifically so that supporting multiple concurrent accounts
  later requires no changes to the *within-one-flow* dispatch/`ReactTo` code —
  only `NewIdentificationOptionsSelectAccount`'s enumeration body would need
  to grow. `checkIdentificationMethod`'s separate, unchanged *config* `one_of`
  lookup (by `Identification` match) remains what resolves nested
  `steps`/`JSONPointerForOneOf` — the two lookups answer different questions
  (§3.4).
- **Scoped simplification, deferred, not dropped**: `login_hint`/`id_token_hint`
  eligibility filtering (the spec's "a mismatched hint omits the option" rule)
  is intentionally **not implemented** in this plan's first iteration — the
  option is shown whenever a usable session exists, independent of any hint.
  `NewIdentificationOptionsSelectAccount` does not read
  `authflow.GetLoginHint(ctx)`/`authflow.GetUserIDHint(ctx)` at all (§3.1).
  Re-adding this is expected follow-up work once multiple-concurrent-account
  support is designed, not a permanent decision — do not read this omission as
  the spec being wrong or superseded.

---

## 10. Implementation order

1. `pkg/api/model/identification.go` — new constant + exhaustive switch cases.
   (Compiles standalone; nothing depends on runtime code yet.)
2. `pkg/lib/config/authentication_flow.go` — schema enum + conditional
   `required` restructuring. (Compiles standalone; config tests can be written
   and run immediately after this commit.)
3. `pkg/lib/authenticationflow/declarative/input_interface.go` +
   `input_take_identification_option_index.go` — new shared input-schema
   plumbing used by both the login and signup_login paths.
4. `pkg/lib/authenticationflow/declarative/data_identification.go` +
   `input_step_identify.go` — option construction + top-level input schema
   `index` support.
5. `pkg/lib/authenticationflow/declarative/error.go` — `ErrSelectAccountSessionChanged`.
6. `pkg/lib/authenticationflow/declarative/node_do_use_identity_select_account.go` +
   `intent_use_identity_select_account.go` — login-flow resolution path.
7. `pkg/lib/authenticationflow/declarative/intent_login_flow_step_identify.go` —
   wire option construction + dispatch for `login` flows. (`login`-flow UC1/UC3
   e2e tests can now pass.)
8. `pkg/lib/authenticationflow/declarative/intent_lookup_identity_select_account.go` —
   signup_login-flow switch path.
9. `pkg/lib/authenticationflow/declarative/intent_signup_login_flow_step_identify.go` —
   wire option construction + dispatch for `signup_login` flows. (UC2/UC4 e2e
   tests can now pass.)
10. E2E infra (`e2e/pkg/testrunner/models.go`, `testcase.go`) — can be done in
    parallel with steps 1-9 since it has no dependency on the authflow
    changes; needed before any e2e test in §8.2 can run.
11. Unit tests (§8.1) alongside each corresponding source change above, not
    deferred to the end.
12. E2E tests (§8.2) once steps 1-10 are complete.
13. `review-pr` skill pass (mandatory per repo convention) before considering
    the change complete.

---

## 11. Atomic commit plan

1. **`Add select_account identification constant`**
   - Files: `pkg/api/model/identification.go`
   - Scope: new const + exhaustive switch arms only. No behavior reachable yet
     (nothing produces this value in config).
   - Tests: none new (existing `identification_test.go` if any, unaffected).

2. **`Allow select_account in login and signup_login identify config`**
   - Files: `pkg/lib/config/authentication_flow.go`
   - Scope: schema enum additions + `AuthenticationFlowSignupLoginFlowIdentify`
     conditional `required` restructuring.
   - Tests: config validation unit tests (§8.1, config case) in the same
     commit.
   - Note: config now accepts `select_account`, but no runtime code
     interprets it yet — parsing succeeds, the identify step simply never
     produces a `select_account` option (falls through no `case` in the
     option-building `switch`, silently ignored) until commit 4/6 lands. This
     is safe to merge/deploy independently.

3. **`Add shared identification-option-index input schema`**
   - Files: `input_interface.go`, `input_take_identification_option_index.go`
   - Scope: new, currently-unused (until commit 4) input type.
   - Tests: `input_take_identification_option_index_test.go`.

4. **`Resolve select_account identification options from the session`**
   - Files: `data_identification.go`, `input_step_identify.go`
   - Scope: `NewIdentificationOptionsSelectAccount`, `selectAccountDisplayName`,
     `IdentificationOption.DisplayName`, `InputStepIdentify.Index` +
     `GetIdentificationOptionIndex()` + schema case. Still not wired into
     either `IntentLoginFlowStepIdentify` or `IntentSignupLoginFlowStepIdentify`'s
     option-building switch — this commit is inert until commit 6/7.
   - Tests: `data_identification_test.go` eligibility-rule table test;
     extended `input_step_identify_test.go`.

5. **`Add SelectAccountSessionChanged error`**
   - Files: `error.go`
   - Scope: one new `var`.

6. **`Complete select_account identification in login flows`**
   - Files: `node_do_use_identity_select_account.go`,
     `intent_use_identity_select_account.go`,
     `intent_login_flow_step_identify.go`
   - Scope: full login-flow path wired end-to-end (option construction →
     dispatch → resolution → `NewNodePostIdentified`). `select_account` is
     now usable in any `login_flows` config.
   - Tests: `resolveSelectAccountSession` unit test; e2e
     `login_direct.test.yaml`, `login_no_session.test.yaml`,
     `login_prompt_login.test.yaml`, `login_step_up_totp.test.yaml`,
     `login_session_changed.test.yaml` (requires commit 9's e2e infra to run,
     but the YAML files themselves can be added here).

7. **`Complete select_account identification in signup_login flows`**
   - Files: `intent_lookup_identity_select_account.go`,
     `intent_signup_login_flow_step_identify.go`
   - Scope: full signup_login-flow switch path wired end-to-end.
   - Tests: e2e `signup_login_switch.test.yaml`, `signup_login_decline.test.yaml`.

8. **`Support injecting an IDP session cookie into e2e create/input flow steps`**
   - Files: `e2e/pkg/testrunner/models.go`, `e2e/pkg/testrunner/testcase.go`
   - Scope: e2e infra only, independent of the authflow changes — can land
     any time relative to commits 1-7, but must land before the e2e test
     files in commits 6-7 can be run in CI.
   - No generated-file regeneration needed (this is hand-written Go, not a
     generator target).

9. **`chore: Update .vettedpositions`**
   - Run `make update-vettedpositions` (or the equivalent skill) after all
     line-number-affecting commits above land, per repo convention — do not
     hand-edit `.vettedpositions`.

Each commit above is independently compilable and independently safe to
deploy (later commits activate behavior earlier commits only made
representable/parseable, never the reverse) — bisect-safe by construction.
