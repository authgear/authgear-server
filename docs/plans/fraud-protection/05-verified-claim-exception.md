# Fraud Protection: Verified Claim Exception & Allow Reason Field

## Context

The fraud protection system currently warns or blocks SMS sending when any configured warning rule is triggered. This plan makes two related changes:

1. **Allow reason field** — add `AllowReason` to `FraudProtectionDecisionRecord` so observers can see *why* a send was allowed even when warnings fired.
2. **Verified claim exception** — if the destination phone number is a verified claim in `_auth_verified_claim`, the send is never blocked (but buckets and metrics still run normally).
3. **Always-allow paths fully participate in pipeline** — currently `isAlwaysAllowed` exits before bucket/metrics logic. That early-return is removed; whitelist matches only affect the final verdict, not whether the pipeline runs.

---

## Full set of AllowReason values

| Value | When set |
|---|---|
| `""` (empty) | No warnings triggered — normal happy path, nothing to explain |
| `"always_allow_ip"` | IP CIDR or IP geolocation whitelist matched |
| `"always_allow_phone"` | Phone geolocation or phone regex whitelist matched |
| `"verified_claim"` | Warnings fired but phone is a verified claim — block suppressed |
| `"record_only"` | Warnings fired but action is `record_only` — block suppressed by config |

`AllowReason` is only set when a potential block was suppressed by something. When the decision is `"blocked"` or when allowed simply because no warnings fired, it is empty (`omitempty` in JSON).

---

## Behaviour change to `CheckAndRecord`

### Before (current)

```
enabled? → parse phone → isAlwaysAllowed? → early return nil (no event, no buckets)
                       ↓
                  compute thresholds → fill buckets → eval warnings → decision → event
```

### After

```
enabled? → parse phone → compute thresholds → fill buckets → eval warnings
                                                                     ↓
                                               alwaysAllowReason? isVerifiedClaim?
                                                                     ↓
                                                          decision + allowReason → event
```

The two silent early-returns (disabled, unparseable) remain as-is — no event in those cases. The `isAlwaysAllowed` check is moved to the decision step and no longer causes an early return.

---

## Decision logic (after the change)

```go
alwaysAllowReason := s.alwaysAllowReason(s.Config, ip, phoneNumber, phoneCountry)

isVerifiedClaim, err := s.VerifiedClaims.ExistsByClaimNameAndValue(
    ctx, string(model.ClaimPhoneNumber), phoneNumber,
)
if err != nil {
    return err
}

var allowReason model.FraudProtectionAllowReason
decision := model.FraudProtectionDecisionAllowed
action := s.Config.Decision.Action

if alwaysAllowReason != "" {
    allowReason = alwaysAllowReason
} else if isVerifiedClaim && len(warnings) > 0 {
    allowReason = model.FraudProtectionAllowReasonVerifiedClaim
} else if action == config.FraudProtectionDecisionActionRecordOnly && len(warnings) > 0 {
    allowReason = model.FraudProtectionAllowReasonRecordOnly
} else if action == config.FraudProtectionDecisionActionDenyIfAnyWarning && len(warnings) > 0 {
    decision = model.FraudProtectionDecisionBlocked
}
```

Priority: `always_allow` > `verified_claim` > `record_only` > normal block/allow.

---

## Files to change

### 1. `pkg/api/model/fraud_protection.go`

Add type and constants:

```go
type FraudProtectionAllowReason string

const (
    FraudProtectionAllowReasonAlwaysAllowIP    FraudProtectionAllowReason = "always_allow_ip"
    FraudProtectionAllowReasonAlwaysAllowPhone FraudProtectionAllowReason = "always_allow_phone"
    FraudProtectionAllowReasonVerifiedClaim    FraudProtectionAllowReason = "verified_claim"
    FraudProtectionAllowReasonRecordOnly       FraudProtectionAllowReason = "record_only"
)
```

Add field to `FraudProtectionDecisionRecord`:

```go
AllowReason FraudProtectionAllowReason `json:"allow_reason,omitempty"`
```

### 2. `pkg/lib/feature/verification/store_pq.go`

Add a new method to `StorePQ`:

```go
// ExistsByClaimNameAndValue returns true if any verified claim row matches
// the given claim name and value, regardless of user ID.
func (s *StorePQ) ExistsByClaimNameAndValue(ctx context.Context, claimName, claimValue string) (bool, error) {
    q := s.selectQuery().Where("name = ? AND value = ?", claimName, claimValue).Limit(1)
    row, err := s.SQLExecutor.QueryRowWith(ctx, q)
    if err != nil {
        return false, err
    }
    _, err = s.scan(row)
    if errors.Is(err, ErrClaimUnverified) {
        return false, nil
    } else if err != nil {
        return false, err
    }
    return true, nil
}
```

`ErrClaimUnverified` is what `scan` returns for `sql.ErrNoRows`, so this correctly maps "no row" → `false`.

### 3. `pkg/lib/fraudprotection/service.go`

**Add interface:**

```go
type VerifiedClaimChecker interface {
    ExistsByClaimNameAndValue(ctx context.Context, claimName, claimValue string) (bool, error)
}
```

**Add field to `Service`:**

```go
VerifiedClaims VerifiedClaimChecker
```

**Refactor `isAlwaysAllowed` → `alwaysAllowReason`** — change the private helper from returning `bool` to returning `FraudProtectionAllowReason`:

```go
func (s *Service) alwaysAllowReason(cfg *config.FraudProtectionConfig, ip, phoneNumber, phoneCountry string) model.FraudProtectionAllowReason {
    if cfg.Decision == nil || cfg.Decision.AlwaysAllow == nil {
        return ""
    }
    alwaysAllow := cfg.Decision.AlwaysAllow
    if isIPAlwaysAllowed(alwaysAllow.IPAddress, ip) {
        return model.FraudProtectionAllowReasonAlwaysAllowIP
    }
    if isPhoneAlwaysAllowed(alwaysAllow.PhoneNumber, phoneNumber, phoneCountry) {
        return model.FraudProtectionAllowReasonAlwaysAllowPhone
    }
    return ""
}
```

**Remove the early-return block in `CheckAndRecord`:**

```go
// REMOVE this:
if s.isAlwaysAllowed(s.Config, ip, phoneNumber, phoneCountry) {
    return nil
}
```

**Replace the decision block** with the logic in the "Decision logic" section above, and pass `allowReason` into the event payload.

### 4. `pkg/lib/deps/deps_common.go`

In the `fraudprotection.DependencySet` block, add:

```go
wire.Bind(new(fraudprotection.VerifiedClaimChecker), new(*verification.StorePQ)),
```

Both `fraudprotection` and `verification` are already imported.

### 5. Regenerate wire

```
make generate
```

---

## Tests

In `pkg/lib/fraudprotection/` (mock `VerifiedClaimChecker` and move existing `isAlwaysAllowed` tests to cover `alwaysAllowReason`):

- `isAlwaysAllowed` IP match → buckets filled, event dispatched, `allow_reason = "always_allow_ip"`, no error returned
- `isAlwaysAllowed` phone match → buckets filled, event dispatched, `allow_reason = "always_allow_phone"`, no error returned
- Verified claim, warnings triggered → buckets filled, event dispatched, `allow_reason = "verified_claim"`, no error returned
- `record_only`, warnings triggered → buckets filled, event dispatched, `allow_reason = "record_only"`, no error returned
- `deny_if_any_warning`, warnings triggered, not verified → `decision = "blocked"`, `ErrBlockedByFraudProtection` returned
- No warnings → `allow_reason = ""`, no error returned
- `VerifiedClaims.ExistsByClaimNameAndValue` returns error → error propagated

---

## Commit steps

1. `[Fraud Protection] Add AllowReason field to FraudProtectionDecisionRecord`
2. `[Fraud Protection] Add ExistsByClaimNameAndValue to verification StorePQ`
3. `[Fraud Protection] Restructure CheckAndRecord: always-allow participates in pipeline`
4. `[Fraud Protection] Never block verified phone number claims in fraud protection`
5. `[Fraud Protection] Wire VerifiedClaimChecker into fraudprotection.Service`
6. `chore: Regenerate wire after VerifiedClaimChecker wiring`
