# Plan: Account Recovery by Username

## Goal

Add `username` as a valid value for `authentication_flow.account_recovery_flows.steps[].one_of[].identification`. Username identifies the user but is not itself a delivery address, so the `select_destination` step needs a new behavior path.

The behavior of `select_destination` depends on the `enumerate_destinations` flag:

- `enumerate_destinations: true` — keep the existing enumeration behavior. After identifying the user by username, the matched user's email/phone identities are listed as destination options. Code is sent to the picked destination.
- `enumerate_destinations: false` — derive the options purely from `allowed_channels` (one option per allowed channel). `masked_display_name` is always the username itself. The flow proceeds without error regardless of whether the user actually has an identity matching the selected channel.

## Scope

In scope:

- `pkg/lib/config` schema and Go constants
- `pkg/lib/authenticationflow/declarative` for the identify step, input schema, and select_destination step
- Unit tests for the new derivation logic
- E2E test for the new flow under `e2e/tests/`

Out of scope:

- Auth UI updates (forgot password page, view models, templates, i18n) — only custom UI for now
- `GenerateAccountRecoveryFlowConfig` changes — username is not auto-generated into the default flow
- New JSON-schema if/then validation rules for `enumerate_destinations`. The new behavior is well-defined for both `true` and `false`, so neither value is invalid.

## Implementation Plan

Land as five atomic commits in this order. Each commit must build and its tests must pass before the next one starts. After all commits, run `make update-vettedpositions` if any line-numbered references in `.vettedpositions` shifted. Use the repo's commit convention from `CLAUDE.md` (imperative subject line, no trailing period). All commits land on a single feature branch and squash-merge into `main`.

### Commit 1 — Add username to the account recovery identification enum

- Title: `[Authflow] Add username to account recovery identification enum`
- Implements: §1 (Files to Modify)
- Files:
  - `pkg/lib/config/authentication_flow.go`
  - `pkg/lib/config/authentication_flow_test.go`
- Tests added in this commit:
  - Extend the existing `TestAuthenticationFlowAccountRecoveryFlow` suite with config cases where `identification: username` parses successfully, and where the value alongside `email` and `phone` is accepted.
- Verifies: `go test ./pkg/lib/config/...`

This commit is a no-op for runtime behavior — the new enum value is not yet used anywhere — but it unlocks the next commit.

### Commit 2 — Route username identification through the identify step

- Title: `[Authflow] Route username identification through account recovery identify step`
- Implements: §2 and §3
- Files:
  - `pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_identify.go`
  - `pkg/lib/authenticationflow/declarative/input_step_account_recovery_identify.go`
- Tests added in this commit:
  - None needed in isolation — the routing is exercised end-to-end by Commit 3's `deriveAccountRecoveryDestinationOptions` tests (when a test sets `iden.Identification = username`, it implicitly requires this commit to have landed) and by Commit 5's e2e flows. If there is an existing test for `InputSchemaStepAccountRecoveryIdentify.SchemaBuilder`, extend it to assert that a `username` option requires `login_id`.
- Verifies: `go build ./...` and `go test ./pkg/lib/authenticationflow/...`

After this commit, a flow configured with `identification: username` reaches `select_destination` and dispatches into the existing email/phone derive path. For `enumerate=true`, that behaves correctly. For `enumerate=false`, options will be empty because `deriveAllowedAccountRecoveryDestinationOptions` has no `username` case — that's fixed in Commit 3. So the feature is *not yet usable* end-to-end; that is expected.

### Commit 3 — Generate username destination options when `enumerate_destinations` is false

- Title: `[Authflow] Generate username destination options when enumerate_destinations is false`
- Implements: §4
- Files:
  - `pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_select_destination.go`
  - `pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_select_destination_test.go` (new; create using the same pattern as nearby `*_test.go` files)
- Tests added in this commit (table-driven against `deriveAccountRecoveryDestinationOptions`):
  - Username + `enumerate=true` + user found with one email and one phone, allowed = `[email, sms]` → 2 options with the actual masked email/phone.
  - Username + `enumerate=true` + user not found, allowed = `[email, sms]` → empty list.
  - Username + `enumerate=false` + user found, allowed = `[email, sms]` → 2 options, both masked = username, both `TargetLoginID` = username.
  - Username + `enumerate=false` + user not found, allowed = `[email]` → 1 option, masked = username, `TargetLoginID` = username.
- Verifies: `go test ./pkg/lib/authenticationflow/declarative/...`

After this commit, the feature is functional in a degraded "silent fail for everyone" form — picks at `select_destination` succeed but no recovery codes are delivered.

### Commit 4 — Resolve target login id for username destinations at input time

- Title: `[Authflow] Resolve target login id for username destinations at input time`
- Implements: §5
- Files:
  - `pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_select_destination.go`
  - `pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_select_destination_test.go`
- Tests added in this commit:
  - `firstMatchingLoginIDForChannel`: all seven cases listed in §6.
  - `resolveUsernameTarget` (or the wired-up `ReactTo`): all five cases listed in §6.
- Verifies: `go test ./pkg/lib/authenticationflow/declarative/...`

After this commit, users with a matching email/phone identity receive a real recovery code via the username flow.

### Commit 5 — Add e2e tests for account recovery by username

- Title: `[Authflow] Add e2e tests for account recovery by username`
- Implements: §7
- Files:
  - `e2e/tests/account_recovery_username/test_enumerate.yaml`
  - `e2e/tests/account_recovery_username/test_no_enumerate.yaml`
  - `e2e/tests/account_recovery_username/users.json`
- Tests: the e2e specifications themselves. Before writing, read at least one existing `e2e/tests/account_recovery_*/test.yaml` to copy the layout, asserter conventions (`[[arrayof]]`, `[[string]]`, etc.), and `user_import` setup. Follow the repo's `write-e2e-test` skill if invoking it is convenient.
- Verifies: the relevant `make -C e2e ...` target for the new directory.

### What review looks at, per commit

- Commit 1: schema change is small enough to eyeball; the new test exercises both accepted and unchanged shapes.
- Commit 2: a reviewer can confirm by reading the three small switch additions; no behavior is observable yet.
- Commit 3: derive logic is local to one function; the new test file shows generated options for every combination.
- Commit 4: the helper and resolver are pure functions with table-driven tests; reviewer can read the table to confirm every branch is covered.
- Commit 5: e2e is the integration check.

If a later commit forces a fix in an earlier one, prefer a follow-up commit over amending — the dependency chain stays valid either way, and history stays linear.

## Current State

Reference points in the existing code:

- Identification enum and Go constants: `pkg/lib/config/authentication_flow.go` lines 645–653 (JSON schema enum) and lines 1389–1394 (Go constants).
- Identify step option building and dispatch: `pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_identify.go`. The relevant switch is in `NewIntentAccountRecoveryFlowStepIdentify` (lines 67–76) and `ReactTo` (lines 141–150). Both handle `email` and `phone` only today.
- Login ID input schema: `pkg/lib/authenticationflow/declarative/input_step_account_recovery_identify.go`, switch in `SchemaBuilder` (lines 58–67).
- Username already supported by `makeLoginIDSpec`: `pkg/lib/authenticationflow/declarative/utils_common.go` lines 927–949 (case `model.AuthenticationFlowIdentificationUsername`). `IntentUseAccountRecoveryIdentity` builds the spec via this helper, so identifying by username already works once the surrounding switches accept it.
- Select destination derivation: `pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_select_destination.go`, function `deriveAccountRecoveryDestinationOptions` (lines 161–191) and helper `deriveAllowedAccountRecoveryDestinationOptions` (lines 193–242).
- Code sending: `NodeDoSendAccountRecoveryCode.Send` in `pkg/lib/authenticationflow/declarative/node_do_send_account_recovery_code.go`. It calls `deps.ForgotPassword.SendCode(ctx, TargetLoginID, ...)`. `forgotpassword.Service.SendCode` (`pkg/lib/feature/forgotpassword/service.go` lines 103–155) looks up identities by `ClaimEmail` / `ClaimPhoneNumber` matching `loginID`. If no identity is found, it generates a dummy OTP for rate-limiting purposes and returns `ErrUserNotFound`. `Send` already silently swallows `ErrUserNotFound`, so a `TargetLoginID` that does not match any email/phone identity is a no-op without erroring.

The "silent fail when target does not resolve to an email/phone" property of `SendCode` is what we rely on for the `enumerate_destinations: false` case to "always return success".

## Design

### Identifying the user

When `identification: username`, `IntentUseAccountRecoveryIdentity` resolves the user via `deps.Identities.SearchBySpec(ctx, spec)` (with `on_failure: ignore`) or `findExactOneIdentityInfo` (with `on_failure: error`). `makeLoginIDSpec` already produces the correct username login-id spec, so no change is needed there. The result is a `NodeDoUseAccountRecoveryIdentity` with `Identification = username` and `MaybeIdentity` either set or nil.

### Select destination — `enumerate_destinations: true`

Existing path already works for any identification:

```go
if iden.MaybeIdentity != nil && step.EnumerateDestinations {
    userIdens, _ := deps.Identities.ListByUser(ctx, iden.MaybeIdentity.UserID)
    for _, channel := range allowedChannels {
        opts := enumerateAllowedAccountRecoveryDestinationOptions(channel, userIdens)
        options = append(options, opts...)
    }
}
```

`enumerateAllowedAccountRecoveryDestinationOptions` operates only on the user's `LoginIDType` (email/phone) and `allowedChannel.Channel`, so passing in a user resolved by username works without modification. No change.

If the username does not match any user (`MaybeIdentity == nil`), the code falls through to the non-enumerate branch and produces an empty options list. **This is existing behavior and we keep it as-is** — we do not add a special-case fallback. Customers who want a privacy-preserving flow when the user is not found should use `enumerate_destinations: false`.

### Select destination — `enumerate_destinations: false` with `username`

This is the new behavior path. The key idea: at *derive time* the options carry only the username; the actual destination lookup happens at *input time*, keyed on the channel the user just picked.

#### Derive-time behavior

For each `allowed_channel` in the step config, produce exactly one `AccountRecoveryDestinationOptionInternal`:

- `MaskedDisplayName = <typed username>` — taken from `iden.IdentitySpec.LoginID.Value.TrimSpace()`. Always the username, never the actual email/phone, so the UI doesn't reveal what address the user has.
- `Channel = allowedChannel.Channel` (email / sms / whatsapp).
- `OTPForm = allowedChannel.OTPForm`.
- `TargetLoginID = <typed username>` — placeholder. The real target is resolved later when the user picks an option.

The resulting list is exactly `len(allowed_channels)` options, regardless of which channels the user actually has. No identity lookup is performed at this stage, so the `select_destination` response is fast and uniform.

#### Input-time behavior — resolve `TargetLoginID` for the picked channel

When `IntentAccountRecoveryFlowStepSelectDestination.ReactTo` receives the index input picking one option:

1. Take the picked option's `Channel`.
2. If the upstream `MilestoneDoUseAccountRecoveryIdentity` has a `MaybeIdentity` (the username matched a real user) **and** the step has `EnumerateDestinations == false` **and** identification is `username`:
   - Call `deps.Identities.ListByUser(ctx, MaybeIdentity.UserID)`.
   - Find the user's first identity matching the picked option's channel:
     - `email` channel → first `LoginIDKeyTypeEmail` identity.
     - `sms` / `whatsapp` channel → first `LoginIDKeyTypePhone` identity.
   - If found, override the picked option's `TargetLoginID` with that identity's `LoginID` value.
   - If not found, leave `TargetLoginID = username` (silent-fail at `SendCode`).
3. Pass the (possibly modified) option into `NodeUseAccountRecoveryDestination`.

For all other cases (email/phone identifications, or username with `enumerate=true`), the picked option already has a correct `TargetLoginID` from the existing derive paths, and no override happens.

This per-pick resolution avoids running `ListByUser` for channels the user never selects.

#### What happens next

After `ReactTo` produces `NodeUseAccountRecoveryDestination` with the resolved option:

1. `IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode.ReactTo` constructs `NodeDoSendAccountRecoveryCode` with the resolved `TargetLoginID`.
2. `Send` calls `deps.ForgotPassword.SendCode(ctx, TargetLoginID, ...)`:
   - **Real email/phone**: `ListByClaim` finds the identity, code is sent for real.
   - **Username fallback**: `ListByClaim` finds nothing, `generateDummyOTP` runs for rate-limit accounting, returns `ErrUserNotFound`, silently swallowed by `Send`.
3. The verify step's data shows `MaskedDisplayName = username`, `Channel = <picked>`, `OTPForm = <picked>` — masked display stays as the username regardless of what TargetLoginID resolved to.
4. If a real code was sent, the user can submit it and proceed to reset password. Otherwise, the verify step exists but no code can satisfy it — the user sees a generic "invalid code" message.

This satisfies the requirement: step transitions never error on "invalid channel" selection. Users with a matching channel get a working recovery flow; users without one see the flow proceed in a way that is indistinguishable from a successful send.

#### Privacy properties

- **UI display is uniform**: number of options = `len(allowed_channels)`, masked display always shows the username. The `select_destination` response is identical for any (user × channel selection) input combination, so the UI never reveals which channels the user actually has.

#### Multiple matching identities

If the user has two emails and the user picks the email option, only the *first* email returned by `ListByUser` receives a code. With `enumerate_destinations: true`, the user picks among all of them; with `false`, the system picks the first.

## Files to Modify

### 1. Config schema and Go constant

**`pkg/lib/config/authentication_flow.go`**

- Add `"username"` to the `AuthenticationFlowAccountRecoveryIdentification` JSON enum (currently lines 645–653, two values: `email`, `phone`).
- Add the matching Go constant after `AuthenticationFlowAccountRecoveryIdentificationPhone` (currently lines 1391–1394):

```go
const (
    AuthenticationFlowAccountRecoveryIdentificationEmail    = AuthenticationFlowAccountRecoveryIdentification(model.AuthenticationFlowIdentificationEmail)
    AuthenticationFlowAccountRecoveryIdentificationPhone    = AuthenticationFlowAccountRecoveryIdentification(model.AuthenticationFlowIdentificationPhone)
    AuthenticationFlowAccountRecoveryIdentificationUsername = AuthenticationFlowAccountRecoveryIdentification(model.AuthenticationFlowIdentificationUsername)
)
```

### 2. Identify step option building and dispatch

**`pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_identify.go`**

- In `NewIntentAccountRecoveryFlowStepIdentify` (lines 67–76), add `case config.AuthenticationFlowAccountRecoveryIdentificationUsername:` to the switch so a `username` branch in the config is included in the offered identification options.
- In `ReactTo` (lines 141–150), add the same case so picking `username` dispatches to `IntentUseAccountRecoveryIdentity` like email/phone does.

### 3. Input schema validation

**`pkg/lib/authenticationflow/declarative/input_step_account_recovery_identify.go`**

- In `SchemaBuilder` (lines 58–67), add:

```go
case config.AuthenticationFlowAccountRecoveryIdentificationUsername:
    requireString("login_id")
    setRequiredAndAppendOneOf()
```

### 4. Select destination — derive-time options

**`pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_select_destination.go`**

Modify `deriveAccountRecoveryDestinationOptions` (lines 161–191) to add a username branch *only for the `enumerate_destinations: false` case*. The `enumerate_destinations: true` and email/phone non-enumerate paths are unchanged.

```go
isUsername := iden.Identification == config.AuthenticationFlowAccountRecoveryIdentificationUsername

switch {
case iden.MaybeIdentity != nil && step.EnumerateDestinations:
    // existing enumerate path, unchanged
    userID := iden.MaybeIdentity.UserID
    userIdens, err := deps.Identities.ListByUser(ctx, userID)
    if err != nil {
        return nil, err
    }
    for _, channel := range allowedChannels {
        opts := enumerateAllowedAccountRecoveryDestinationOptions(channel, userIdens)
        options = append(options, opts...)
    }
case isUsername && !step.EnumerateDestinations:
    // new username path: one option per allowed channel.
    // No identity lookup here — TargetLoginID is the typed username and will be
    // resolved to the user's actual email/phone at ReactTo time when the user
    // picks one of the options.
    username := iden.IdentitySpec.LoginID.Value.TrimSpace()
    for _, channel := range allowedChannels {
        options = append(options, &AccountRecoveryDestinationOptionInternal{
            AccountRecoveryDestinationOption: AccountRecoveryDestinationOption{
                MaskedDisplayName: username,
                Channel:           AccountRecoveryChannel(channel.Channel),
                OTPForm:           AccountRecoveryOTPForm(channel.OTPForm),
            },
            TargetLoginID: username,
        })
    }
default:
    // existing email/phone non-enumerate path, unchanged.
    // Also covers username + enumerate=true + user not found (returns empty options).
    for _, channel := range allowedChannels {
        opts := deriveAllowedAccountRecoveryDestinationOptions(channel, iden)
        options = append(options, opts...)
    }
}
```

No changes are needed to `deriveAllowedAccountRecoveryDestinationOptions` or `enumerateAllowedAccountRecoveryDestinationOptions`.

### 5. Select destination — input-time resolution

**`pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_select_destination.go`**

Modify `IntentAccountRecoveryFlowStepSelectDestination.ReactTo` (lines 111–123) to resolve `TargetLoginID` for username + enumerate=false picks. After the user submits the option index but before constructing `NodeUseAccountRecoveryDestination`:

```go
func (i *IntentAccountRecoveryFlowStepSelectDestination) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
    if len(flows.Nearest.Nodes) == 0 {
        var inputTakeAccountRecoveryDestinationOptionIndex inputTakeAccountRecoveryDestinationOptionIndex
        if authflow.AsInput(input, &inputTakeAccountRecoveryDestinationOptionIndex) {
            optionIdx := inputTakeAccountRecoveryDestinationOptionIndex.GetAccountRecoveryDestinationOptionIndex()
            option := i.Options[optionIdx]

            resolved, err := i.resolveUsernameTarget(ctx, deps, flows, option)
            if err != nil {
                return nil, err
            }

            return authflow.NewNodeSimple(&NodeUseAccountRecoveryDestination{
                Destination: resolved,
            }), nil
        }
    }
    return nil, authflow.ErrIncompatibleInput
}

// resolveUsernameTarget overrides TargetLoginID with the user's first identity
// matching the picked option's Channel, but only for the
// "username + enumerate_destinations=false + user found" sub-case.
// All other flows return the option unchanged.
func (i *IntentAccountRecoveryFlowStepSelectDestination) resolveUsernameTarget(
    ctx context.Context,
    deps *authflow.Dependencies,
    flows authflow.Flows,
    option *AccountRecoveryDestinationOptionInternal,
) (*AccountRecoveryDestinationOptionInternal, error) {
    current, err := i.currentFlowObject(deps, flows, i)
    if err != nil {
        return nil, err
    }
    step := i.step(current)
    if step.EnumerateDestinations {
        return option, nil
    }

    ms := authflow.FindAllMilestones[MilestoneDoUseAccountRecoveryIdentity](flows.Root)
    if len(ms) == 0 {
        return option, nil
    }
    accIden := ms[0].MilestoneDoUseAccountRecoveryIdentity()
    if accIden.Identification != config.AuthenticationFlowAccountRecoveryIdentificationUsername {
        return option, nil
    }
    if accIden.MaybeIdentity == nil {
        return option, nil
    }

    userIdens, err := deps.Identities.ListByUser(ctx, accIden.MaybeIdentity.UserID)
    if err != nil {
        return nil, err
    }
    if target := firstMatchingLoginIDForChannel(userIdens, option.Channel); target != "" {
        // Mutate a copy, not the slice element.
        copied := *option
        copied.TargetLoginID = target
        return &copied, nil
    }
    return option, nil
}
```

Add the helper next to the existing destination helpers in the same file:

```go
// firstMatchingLoginIDForChannel returns the first login-id value among `userIdens`
// whose login-id type maps to the requested account-recovery channel.
// email → LoginIDKeyTypeEmail. sms / whatsapp → LoginIDKeyTypePhone.
// Returns "" when no matching identity is present.
func firstMatchingLoginIDForChannel(
    userIdens []*identity.Info,
    channel AccountRecoveryChannel,
) string {
    var wantType model.LoginIDKeyType
    switch channel {
    case AccountRecoveryChannelEmail:
        wantType = model.LoginIDKeyTypeEmail
    case AccountRecoveryChannelSMS, AccountRecoveryChannelWhatsapp:
        wantType = model.LoginIDKeyTypePhone
    default:
        return ""
    }
    for _, ui := range userIdens {
        if ui.Type != model.IdentityTypeLoginID {
            continue
        }
        if ui.LoginID.LoginIDType == wantType {
            return ui.LoginID.LoginID
        }
    }
    return ""
}
```

The mutate-a-copy step matters because `i.Options` is persisted as part of the intent state; we do not want a bare option mutation to be observable in OutputData on later inspections.

### 6. Unit tests

**`pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_select_destination_test.go`** (new file if it does not exist, otherwise extend)

Tests for `deriveAccountRecoveryDestinationOptions`:

- Username + `enumerate=true` + user found with one email and one phone, allowed channels = [email, sms]
  - expected: 2 options, masked_display_name = the actual masked email/phone (existing enumerate behavior, no change).
- Username + `enumerate=true` + user not found, allowed channels = [email, sms]
  - expected: empty options list (existing behavior, kept as-is).
- Username + `enumerate=false` + user found, allowed channels = [email, sms]
  - expected: 2 options, both masked_display_name = username, both TargetLoginID = username (resolution happens later in ReactTo, not here).
- Username + `enumerate=false` + user not found, allowed channels = [email]
  - expected: 1 option, masked_display_name = username, TargetLoginID = username.

Tests for `firstMatchingLoginIDForChannel`:

- Empty `userIdens` → returns "".
- `userIdens` containing one email + one phone, channel = email → returns the email's LoginID value.
- Same input, channel = sms → returns the phone's LoginID value.
- Same input, channel = whatsapp → returns the phone's LoginID value.
- `userIdens` containing only emails, channel = sms → returns "".
- `userIdens` containing two emails, channel = email → returns the *first* email's LoginID (documents "first match wins").
- `userIdens` containing a non-login-id identity (e.g., oauth) → that one is skipped.

Tests for `IntentAccountRecoveryFlowStepSelectDestination.ReactTo` (or `resolveUsernameTarget` directly if extracted):

- Username + `enumerate=false` + user found with email, picks email option → returned option has TargetLoginID = user's email.
- Username + `enumerate=false` + user found with only phone, picks email option → returned option has TargetLoginID = username (silent-fail fallback).
- Username + `enumerate=false` + user not found, picks any option → returned option has TargetLoginID = username.
- Username + `enumerate=true` (any) → returned option is unchanged (no override applied).
- Email identification + `enumerate=false`, picks the option → returned option is unchanged (no override applied).

### 7. E2E test

**`e2e/tests/account_recovery_username/`** (new directory)

Two test files:

- `test_enumerate.yaml`: identify by username; the matched user has email + phone; assert select_destination shows enumerated email/phone options.
- `test_no_enumerate.yaml`: identify by username; allowed channels include email + sms; assert select_destination shows two options both masked as the username, regardless of user's actual identities. Continue past select to verify the next step transition succeeds.

`users.json` defines a user with `username` + `email` login IDs to support both tests.

Use the existing E2E format already used in `e2e/tests/account_recovery_*` directories (this codebase already has account recovery e2e tests; copy the conventions).

## Verification

1. Schema: try saving a config with `identification: username` — should pass schema validation.
2. Custom-UI flow with `enumerate_destinations: true`:
   - Username matches a user with email → select_destination shows masked email; picking it sends the recovery code.
   - Username does not match any user → select_destination returns an empty options list (existing behavior, kept as-is).
3. Custom-UI flow with `enumerate_destinations: false`:
   - Username matches a user with email, allowed = [email] → 1 option masked as username; picking it actually sends a code to the user's email; the verify step shows masked_display_name = username; the user can submit the real code and reset password.
   - Username matches a user with only phone, allowed = [email, sms] → 2 options; picking email is a silent no-op (no code sent), picking sms sends to the user's phone.
   - Username does not match any user, allowed = [email] → 1 option masked as username; picking it is a silent no-op; the flow still advances to verify_code with no real code in existence.
4. Run unit tests: `go test ./pkg/lib/authenticationflow/declarative/...`.
5. Run E2E: `make -C e2e test` (or whatever the repo's e2e make target is) for the new directory.

## Non-goals / Follow-ups

- Auth UI: the built-in forgot-password page does not yet render a username input. A separate plan can add view-model support, template rendering, and i18n. Until then, only the API/JSON flow exposes username recovery.
- `GenerateAccountRecoveryFlowConfig` is not modified, so apps with `username` login-id keys do not automatically get a username recovery branch in the default flow. Customers must opt in by writing the flow config.
- Multiple-identity selection under `enumerate_destinations: false` is hard-coded to "first match wins". If we later want a deterministic preference (e.g., primary identity, last-used, oldest), that's a follow-up change in `firstMatchingLoginIDForChannel`.
