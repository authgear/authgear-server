# Authflow Session OTP Cooldown Plan

## Goal

Add a second cooldown layer for OTP/message sending in authflow:

- keep the existing per-target cooldown unchanged
- add a per-authflow-session cooldown
- require both cooldowns to have passed before another send is allowed
- keep the cooldown separated by channel so switching from SMS to WhatsApp can still send immediately

This is intended to close the gap where an attacker can bypass the current cooldown by rotating targets within one authflow session.

## Scope

In scope:

- declarative authflow OTP sends
- declarative account recovery verify-code step
- authflow resend state / `CanResendAt`
- authflow tests

Out of scope for this change:

- legacy `workflow` / `latte` flows that do not use `pkg/lib/authenticationflow.Session`
- changing the existing per-target cooldown semantics
- adding new user-facing config unless we decide later that the session cooldown duration must be independently configurable

Decisions:

- this update applies only to authflow paths
- session cooldown state is not persisted in `authenticationflow.Session`
- session cooldown bucket formatting follows the existing cooldown pattern for each OTP kind

## Current State

Today cooldown is enforced only by OTP target through `kind.RateLimitTriggerCooldown(target)` in [pkg/lib/authn/otp/service.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/service.go). Declarative authflow nodes then expose resend timing through `InspectState`, which currently only reflects that target cooldown:

- [pkg/lib/authenticationflow/declarative/node_verify_claim.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_verify_claim.go)
- [pkg/lib/authenticationflow/declarative/node_authn_oob.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_authn_oob.go)
- [pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go)

Authflow already has a persisted session object at [pkg/lib/authenticationflow/session.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/session.go). For this change, the key thing we need from it is the stable `FlowID`, which can be used as the per-session cooldown key.

## Design

### 1. Reuse the existing cooldown duration

Use the existing `trigger_cooldown` duration for the same purpose/channel.

That keeps the behavior simple:

- verification SMS target cooldown = verification SMS session cooldown duration
- verification WhatsApp target cooldown = verification WhatsApp session cooldown duration
- OOB OTP and forgot-password keep their own existing cooldown durations

This avoids a config migration and means “both cooldowns must pass” is easy to reason about.

### 2. Add a second cooldown bucket keyed by authflow session

For each existing target cooldown bucket, add a parallel session-scoped bucket. The bucket key should be derived from:

- the existing cooldown family
- the authflow `FlowID`
- the same argument layout pattern already used by the target cooldown bucket for that kind

The simpler implementation is to mirror the existing `Kind` methods with a second method, for example:

```go
RateLimitTriggerCooldownSession(flowID string) ratelimit.BucketSpec
```

That keeps purpose/channel-specific bucket naming inside each kind implementation instead of re-deriving it in `otp.Service`.

The important part is to follow the current bucket formatting style:

- verification cooldown today is `NewCooldownSpec(name, period, target)`
- forgot-password cooldown today is `NewCooldownSpec(name, period, target)`
- OOB OTP cooldown today is `NewCooldownSpec(name, period, purpose, target)`

The session cooldown should mirror that same structure, replacing the target portion with `flowID` instead of inventing a different shape. Concretely:

- verification session cooldown: `NewCooldownSpec(name, period, flowID)`
- forgot-password session cooldown: `NewCooldownSpec(name, period, flowID)`
- OOB OTP session cooldown: `NewCooldownSpec(name, period, purpose, flowID)`

This keeps the generated Redis keys consistent with the existing style from [pkg/lib/ratelimit/bucket.go](/Users/tung/repo/authgear-server/pkg/lib/ratelimit/bucket.go) and the current cooldown spec implementations in:

- [pkg/lib/authn/otp/kind_verification.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind_verification.go)
- [pkg/lib/authn/otp/kind_oob_otp.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind_oob_otp.go)
- [pkg/lib/authn/otp/kind_forgot_password.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind_forgot_password.go)

### 3. Pass authflow session identity into OTP generation and state inspection

Extend OTP APIs so authflow callers can opt into session cooldown enforcement:

- add `AuthenticationFlowID string` to `otp.GenerateOptions`
- add an `InspectStateOptions` struct with `AuthenticationFlowID string`
- change `InspectState` to accept that optional context

Behavior:

- if `AuthenticationFlowID` is empty, OTP behavior stays exactly as today
- if `AuthenticationFlowID` is present, `GenerateOTP` checks both:
  - `kind.RateLimitTriggerCooldown(target)`
  - `kind.RateLimitTriggerCooldownSession(flowID)`
- `InspectState` returns `CanResendAt = max(targetCooldown, sessionCooldown)`

This is the important UX piece; otherwise the server would block resend correctly but the frontend would still show the target-only timer.

### 4. Scope the new cooldown by channel, not across channels

The session cooldown must remain channel-specific.

That means:

- SMS send should not block an immediate WhatsApp send
- WhatsApp send should not block an immediate SMS send
- email cooldown remains independent from phone channels

Using channel-specific session buckets preserves the current “switch channel to send immediately” behavior.

### 5. Keep session cooldown semantics aligned with existing target cooldown semantics

The new session cooldown should be consumed at the same point the current target cooldown is consumed: during OTP generation, before message delivery.

That matches current behavior:

- if generation is blocked, nothing is sent
- if generation succeeds but downstream delivery fails, the cooldown still applies

Changing that semantic only for the session cooldown would make behavior inconsistent and harder to reason about.

## Method Call Plan

### Verification / claim verification flow

Current call flow:

1. [node_verify_claim.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_verify_claim.go) calls `GenerateCode`.
2. `GenerateCode` calls `deps.OTPCodes.GenerateOTP(...)`.
3. `OutputData` calls `deps.OTPCodes.InspectState(...)`.

Planned call flow:

1. `GenerateCode` reads `flowID := authflow.GetSession(ctx).FlowID`.
2. `GenerateCode` calls `deps.OTPCodes.GenerateOTP(..., &otp.GenerateOptions{..., AuthenticationFlowID: flowID})`.
3. `OutputData` reads the same `flowID`.
4. `OutputData` calls `deps.OTPCodes.InspectState(..., &otp.InspectStateOptions{AuthenticationFlowID: flowID})`.

### Authentication OOB flow

Current call flow:

1. [node_authn_oob.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_authn_oob.go) calls `GenerateCode`.
2. `GenerateCode` calls `deps.OTPCodes.GenerateOTP(...)`.
3. `OutputData` calls `deps.OTPCodes.InspectState(...)`.

Planned call flow:

1. `GenerateCode` reads `flowID := authflow.GetSession(ctx).FlowID`.
2. `GenerateCode` calls `deps.OTPCodes.GenerateOTP(..., &otp.GenerateOptions{..., AuthenticationFlowID: flowID})`.
3. `OutputData` reads the same `flowID`.
4. `OutputData` calls `deps.OTPCodes.InspectState(..., &otp.InspectStateOptions{AuthenticationFlowID: flowID})`.

### Account recovery verify-code step

Current call flow:

1. [intent_account_recovery_flow_step_verify_account_recovery_code.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go) enters the step.
2. It constructs [NodeDoSendAccountRecoveryCode](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_do_send_account_recovery_code.go).
3. `NodeDoSendAccountRecoveryCode.Send(...)` calls `deps.ForgotPassword.SendCode(...)`.
4. `OutputData` calls `deps.ForgotPassword.InspectState(...)`.

Planned call flow:

1. The intent reads `flowID := authflow.GetSession(ctx).FlowID`.
2. On initial entry, it passes `AuthenticationFlowID: flowID` into `NodeDoSendAccountRecoveryCode.Send(...)`.
3. On resend, it passes the same `AuthenticationFlowID: flowID` into `NodeDoSendAccountRecoveryCode.Send(...)`.
4. `OutputData` passes the same `flowID` into `deps.ForgotPassword.InspectState(...)`.
5. `forgotpassword.Service` forwards `AuthenticationFlowID` into `deps.OTPCodes.GenerateOTP(...)` and `deps.OTPCodes.InspectState(...)`.

### OTP service

Current call flow:

1. `GenerateOTP` checks only `kind.RateLimitTriggerCooldown(target)`.
2. `InspectState` reads only the target cooldown time.

Planned call flow:

1. `GenerateOTP` checks `kind.RateLimitTriggerCooldown(target)`.
2. If `AuthenticationFlowID != ""`, it also checks `kind.RateLimitTriggerCooldownSession(flowID)`.
3. `InspectState` reads the target cooldown time.
4. If `AuthenticationFlowID != ""`, it also reads the session cooldown time.
5. `InspectState` returns the later of the two timestamps as `CanResendAt`.

## Implementation Steps

### 1. OTP kind and service changes

Files:

- [pkg/lib/authn/otp/kind.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind.go)
- [pkg/lib/authn/otp/kind_verification.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind_verification.go)
- [pkg/lib/authn/otp/kind_oob_otp.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind_oob_otp.go)
- [pkg/lib/authn/otp/kind_forgot_password.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind_forgot_password.go)
- [pkg/lib/authn/otp/service.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/service.go)
- [pkg/lib/authn/otp/state.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/state.go)

Work:

- add session cooldown bucket generation to each kind
- extend generate/inspect APIs to optionally carry `AuthenticationFlowID`
- in `GenerateOTP`, enforce both cooldowns when `AuthenticationFlowID` is set
- in `InspectState`, compute the combined resend time as the later of the two cooldowns
- keep callers with empty `AuthenticationFlowID` unchanged

### 2. Authflow callers

Files:

- [pkg/lib/authenticationflow/declarative/node_verify_claim.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_verify_claim.go)
- [pkg/lib/authenticationflow/declarative/node_authn_oob.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_authn_oob.go)
- [pkg/lib/authenticationflow/declarative/node_do_send_account_recovery_code.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_do_send_account_recovery_code.go)
- [pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go)
- [pkg/lib/feature/forgotpassword/service.go](/Users/tung/repo/authgear-server/pkg/lib/feature/forgotpassword/service.go)

Work:

- pass `authflow.GetSession(ctx).FlowID` into OTP generation from declarative authflow nodes
- pass the same flow ID into `InspectState`
- thread authflow flow ID into forgot-password authflow sends so account recovery also participates in the per-session cooldown
- update [intent_account_recovery_flow_step_verify_account_recovery_code.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go) so initial entry, resend, and output-data state all use the authflow session-aware forgot-password path

For forgot-password, the cleanest shape is likely to add `AuthenticationFlowID` to `forgotpassword.CodeOptions`, then forward it into `otp.GenerateOptions`.

### 3. Resend state

Files:

- [pkg/lib/authn/otp/service.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/service.go)
- [pkg/lib/authenticationflow/declarative/node_verify_claim.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_verify_claim.go)
- [pkg/lib/authenticationflow/declarative/node_authn_oob.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_authn_oob.go)
- [pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go)

Work:

- ensure `VerifyOOBOTPData.CanResendAt` reflects the max of target cooldown and session cooldown
- ensure account recovery `CanResendAt` reflects the same combined cooldown
- confirm the existing resend button/countdown UI does not need schema changes because it already consumes `CanResendAt`

## Testing Plan

### Unit tests

Add/update tests for:

- `otp.Service.GenerateOTP`
  - target cooldown alone still works
  - session cooldown blocks a second send in the same authflow with a different target
  - switching channel does not hit the same session cooldown bucket
  - non-authflow callers with empty `AuthenticationFlowID` remain unchanged
- `otp.Service.InspectState`
  - returns the later of target cooldown and session cooldown
### Authflow integration tests

Add declarative authflow tests that cover:

- initial SMS send in a flow succeeds
- resend in the same flow before cooldown expires is blocked even after changing to a different phone number
- switching from SMS to WhatsApp in the same flow still sends immediately
- existing per-target cooldown still blocks resending to the same target after restarting authflow or using a different flow
- account recovery authflow follows the same rules on the verify-account-recovery-code step

Good locations to extend:

- [pkg/lib/authenticationflow/session_test.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/session_test.go)
- existing authflow declarative tests around OOB OTP / claim verification
- account recovery authflow tests around [intent_account_recovery_flow_step_verify_account_recovery_code.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go)

### E2E tests

Add end-to-end coverage that exercises the actual authflow API surface:

- sign-up / verification flow
  - first SMS send succeeds
  - re-send in the same authflow with a different phone target is blocked by session cooldown
  - switching from SMS to WhatsApp in the same authflow still sends immediately
- authentication OOB flow
  - same-flow, different-target resend is blocked
  - channel switch still works immediately
- account recovery flow
  - first code send succeeds
  - resend in the same authflow is blocked by session cooldown even if the target changes
  - channel switch still works immediately

Good locations to extend:

- existing authflow e2e coverage in the repo's authflow/API test suites
- any testrunner scenarios that already cover verification, OOB OTP, or account recovery resend behavior

## Rollout Notes

- This should be low-risk for non-authflow paths because the new API remains opt-in on `AuthenticationFlowID`, and this plan only updates authflow callers.
- The main regression risk is resend UX showing the wrong timer; that is why `InspectState` must be updated together with enforcement.
- If product later wants a different duration for per-session cooldown, that can be added as config after this change. This plan keeps the first version aligned with existing cooldown durations.

## Atomic Commit Plan

### Commit 1: Add session-cooldown support to OTP kinds and service

Goal:

- introduce the OTP-layer capability to enforce and inspect a session cooldown without changing any callers yet

Changes:

- add `RateLimitTriggerCooldownSession(flowID string) ratelimit.BucketSpec` to [pkg/lib/authn/otp/kind.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind.go)
- implement it in:
  - [pkg/lib/authn/otp/kind_verification.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind_verification.go)
  - [pkg/lib/authn/otp/kind_oob_otp.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind_oob_otp.go)
  - [pkg/lib/authn/otp/kind_forgot_password.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/kind_forgot_password.go)
- add `AuthenticationFlowID` to `otp.GenerateOptions`
- add `otp.InspectStateOptions` with `AuthenticationFlowID`
- change `InspectState` to accept options
- update [pkg/lib/authn/otp/service.go](/Users/tung/repo/authgear-server/pkg/lib/authn/otp/service.go) so:
  - `GenerateOTP` checks target cooldown and, when `AuthenticationFlowID` is present, session cooldown
  - `InspectState` returns the max of target cooldown and session cooldown

Tests:

- unit tests for `GenerateOTP`
  - empty `AuthenticationFlowID` preserves current behavior
  - session cooldown blocks same-flow re-send with different target
  - SMS and WhatsApp use different session buckets
- unit tests for `InspectState`
  - returns target cooldown when no flow ID
  - returns max(target, session) when flow ID is present

Why atomic:

- this commit changes only the OTP layer and remains backward-compatible because all existing callers can pass no flow ID

### Commit 2: Thread authflow FlowID through verification and authn OOB

Goal:

- enable per-authflow-session cooldown for the main declarative OTP authflow paths

Changes:

- update [pkg/lib/authenticationflow/declarative/node_verify_claim.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_verify_claim.go)
  - pass `authflow.GetSession(ctx).FlowID` into `GenerateOTP`
  - pass the same flow ID into `InspectState`
- update [pkg/lib/authenticationflow/declarative/node_authn_oob.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_authn_oob.go)
  - pass `authflow.GetSession(ctx).FlowID` into `GenerateOTP`
  - pass the same flow ID into `InspectState`

Tests:

- declarative authflow tests for verification / OOB paths
  - resend blocked in same flow even when target changes
  - switching SMS to WhatsApp still sends immediately
  - target cooldown still works independently
  - `CanResendAt` reflects the combined cooldown
- e2e coverage for verification / OOB authflow resend behavior

Why atomic:

- this commit affects only declarative verification / OOB flows, with no forgot-password or account-recovery changes mixed in

### Commit 3: Thread authflow FlowID through forgot-password service

Goal:

- make forgot-password service capable of session-aware cooldown when called from authflow

Changes:

- add `AuthenticationFlowID` to `forgotpassword.CodeOptions`
- add a matching optional parameter to forgot-password inspect-state path if needed
- update [pkg/lib/feature/forgotpassword/service.go](/Users/tung/repo/authgear-server/pkg/lib/feature/forgotpassword/service.go) so:
  - `SendCode` forwards `AuthenticationFlowID` into `otp.GenerateOptions`
  - `InspectState` forwards `AuthenticationFlowID` into `otp.InspectStateOptions`

Tests:

- forgot-password service unit tests
  - no flow ID preserves current behavior
  - flow ID enables session cooldown
  - inspect-state returns combined resend time when flow ID is present

Why atomic:

- this commit prepares the forgot-password service without yet changing the account recovery authflow call sites

### Commit 4: Enable session cooldown in account recovery authflow

Goal:

- apply the new cooldown to the authflow account recovery verify-code step

Changes:

- update [pkg/lib/authenticationflow/declarative/node_do_send_account_recovery_code.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/node_do_send_account_recovery_code.go)
  - accept `AuthenticationFlowID`
  - pass it into `deps.ForgotPassword.SendCode(...)`
- update [pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go](/Users/tung/repo/authgear-server/pkg/lib/authenticationflow/declarative/intent_account_recovery_flow_step_verify_account_recovery_code.go)
  - pass `authflow.GetSession(ctx).FlowID` on initial entry
  - pass the same flow ID on resend
  - pass the same flow ID into `deps.ForgotPassword.InspectState(...)`

Tests:

- account recovery authflow tests
  - initial send works
  - resend in same flow is blocked even after changing target
  - switching SMS to WhatsApp still sends immediately
  - `CanResendAt` reflects the combined cooldown
- e2e coverage for account recovery resend behavior

Why atomic:

- this commit finishes the authflow-only scope by wiring the prepared forgot-password service into the account recovery authflow path

### Commit 5: Cleanup and regression test pass

Goal:

- make the implementation easy to maintain and ensure nothing outside authflow regressed

Changes:

- clean up naming/comments if needed after the functional commits
- add or refine any missing table-driven tests discovered during implementation
- verify no legacy non-authflow callers were accidentally switched to session-aware mode

Tests:

- targeted OTP tests
- targeted declarative authflow tests
- targeted forgot-password/account-recovery tests
- targeted e2e authflow tests

Why atomic:

- keeps follow-up refactors and test hardening separate from behavior-changing commits
