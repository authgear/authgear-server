---
name: update-configs
description: Add or change a field/section in any of the three config surfaces in pkg/lib/config — project config (authgear.yaml), feature config (authgear.features.yaml), or secrets (authgear.secrets.yaml). Use when adding a new config field, a new top-level section, a new secret key, or touching an existing feature-config section's Merge implementation.
---

# Update Configs

`pkg/lib/config` backs three separate documents, each with its own Go types,
its own JSON schema, and its own required testdata coverage. Get the surface
right first — the rest of this skill is organized by surface.

| Surface | Document | Go types | Parsed/validated by |
|---|---|---|---|
| Project config | `authgear.yaml` | `AppConfig`, `pkg/lib/config/*.go` (non-`feature_`/`secret_` files) | `config.Parse` (schema + `AppConfig.Validate`) |
| Feature config | `authgear.features.yaml` | `FeatureConfig`, `pkg/lib/config/feature_*.go` | `config.ParseFeatureConfig` (schema) + `FeatureConfig.Merge` |
| Secrets | `authgear.secrets.yaml` | `SecretConfig`, `pkg/lib/config/secret_*.go` | `config.ParseSecret` (schema) + `SecretConfig.Validate` (cross-checked against `AppConfig`) |

A single feature often touches more than one surface at once (e.g. a new
app-config section gated by a feature-config flag, or an app-config section
backed by a secret) — apply the relevant part of each section below for
every surface the change touches, don't stop at the first one.

## Hard requirement: every surface has a required testdata file

**This is the single most commonly missed step.** Each surface's schema and
validation behavior is exercised by a shared, hand-maintained testdata
fixture — a YAML file of `name`/`error`/`config` cases, each parsed through
the real production parser and asserted against an exact expected error (or
`null`). Adding or changing a field's schema, a validation rule, or
anything that changes what does or doesn't parse **without** adding a case
to the matching file is an incomplete change, not a smaller one — it looks
done because the code compiles and existing tests pass, but nothing
actually exercises the new behavior.

| Surface | Required testdata file | Consuming test |
|---|---|---|
| Project config | `pkg/lib/config/testdata/config_tests.yaml` | `TestAppConfig` (`config_test.go`) |
| Feature config (schema) | `pkg/lib/config/testdata/parse_feature_tests.yaml` | `TestParseFeatureConfig` (`feature_test.go`) |
| Feature config (merge) | `pkg/lib/config/testdata/merge_feature.yaml` | see "Required test coverage" under Feature config below |
| Secrets (schema) | `pkg/lib/config/testdata/parse_secret_tests.yaml` | `TestParseSecret` (`secret_test.go`) |
| Secrets (cross-validation against app config) | `pkg/lib/config/testdata/secret_config_validate_tests.yaml` | `TestSecretConfigValidate` (`secret_test.go`) |

Each of these files follows the same shape — a sequence of `---`-separated
documents:

```yaml
name: some-case-name
error: |-
  invalid configuration:
  /path/to/field: reason
    map[actual:... expected:...]
config:
  id: test
  http:
    public_origin: http://test
  ... the field under test ...
```

(`secret_config_validate_tests.yaml` instead has `app_config:` and
`secret_config:` keys, since it validates the two together — see its
existing cases for the shape.) `error: null` asserts the input parses
cleanly; a non-null `error` asserts the exact error string.

**Do not guess the expected error string.** Write a throwaway test (or add
a temporary `t.Logf`/`fmt.Printf` case) that runs the actual input through
`config.Parse` / `config.ParseFeatureConfig` / `config.ParseSecret`, capture
the real error text, then delete the throwaway test. A guessed string that
doesn't match byte-for-byte fails the moment someone actually runs it — and
if it's wrong in a way that makes the test trivially pass (e.g. an empty
`ShouldBeError` matcher), it silently tests nothing.

At minimum, add a case for:

- the new field/section parsing successfully with valid input
- each new validation rule actually rejecting the input it's meant to reject, with the real error text
- each new schema constraint (`pattern`, `enum`, `minimum`/`maximum`, `format`, `required`) being enforced

## Project config (`authgear.yaml`)

Schema lives in `Schema.Add(...)` blocks in the relevant `pkg/lib/config/*.go`
file; struct-level cross-field validation goes in `AppConfig.Validate` (see
`c.validateX` methods in `config.go`) — use this for checks the JSON Schema
vocabulary can't express (uniqueness across array items, cross-field
comparisons, etc.), not for anything a schema keyword already covers.

Both layers are exercised by `config_tests.yaml` in one pass, since
`config.Parse` runs schema validation and then `AppConfig.Validate` in
sequence — one case per interesting input is enough, you don't need to
duplicate schema-only vs. validate-only cases.

## Feature config (`authgear.features.yaml`)

Feature config is merged across layers — code default ← cluster ← plan ←
app override — via `FeatureConfig.Merge` (`pkg/lib/config/feature.go`),
which reflects over every top-level field and dispatches to that section's
own `Merge` in `pkg/lib/config/feature_*.go`.

### Hard requirement: merge must be field-level, never whole-section replace

Historically, most sections implemented `Merge` as a wholesale swap:

```go
// WRONG for any section with more than one leaf field
func (c *XFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.X == nil {
		return c
	}
	return layer.X
}
```

This is a real, reachable bug class, not a theoretical one: if a lower layer
(e.g. a plan) sets field `A` and a higher layer (e.g. an app override) later
sets a sibling field `B` — without repeating `A` — the whole-section swap
silently resets `A` back to its code default. Plan/app documents are
routinely partial, hand-authored YAML, so this triggers in practice, not just
in edge cases. It was found and fixed across `identity`, `authentication`,
`authenticator`, `ui`, `hook`, `collaborator`, `messaging.rate_limits`, and
`test_mode` — do not reintroduce it in a new section or a new field on an
existing section.

**Every new field on an existing multi-field section, and every new
top-level section, must merge field-level:**

- Follow the reference pattern in `OAuthClientFeatureConfig.Merge`
  (`pkg/lib/config/feature_oauth.go`): nil-safe guards first
  (`if c == nil && layer == nil { return nil }`, `if c == nil { return layer }`,
  `if layer == nil { return c }`), then per-field
  `if layer.X != nil { c.X = layer.X }` for every field.
- If a section/sub-object genuinely has only **one** field, a whole-object
  replace at that level is fine — there's nothing else to lose. But if that
  one field is itself an object with siblings further down, cascade the
  field-level merge all the way down to where the real siblings are, even
  through single-field wrapper levels. See `feature_authenticator.go`'s
  `Authenticator → Password → Policy` cascade: `Authenticator` and `Password`
  each have only one field, but `Policy` has three siblings that must merge
  independently — so the cascade goes three levels deep, not stopping at the
  first single-field level.
- Never write `if layer.Section == nil { return c }; return layer.Section`
  for a section with more than one leaf field, directly or transitively.

### Required test coverage

Every new/changed `Merge` implementation needs a case in
`pkg/lib/config/testdata/merge_feature.yaml`: one layer sets field `A` only,
a later layer sets a sibling field `B` only (never repeating `A`) — assert
the final effective config has **both** `A` (from the first layer) and `B`
(from the second), not `A` reset to its default. Pick values that are **not**
the code default for the field being tested — otherwise a whole-section
regression would silently produce the "right" value by accident and the test
wouldn't catch it. See the existing `hook`/`collaborator`/`identity` cases in
that file for the pattern.

This is in addition to, not instead of, the schema-level cases required in
`parse_feature_tests.yaml` by the hard requirement above — merge behavior
and schema validation are different code paths and need separate coverage.

### Schema/runtime consistency

Before adding a JSON schema constraint on a feature config field
(`minItems`, `minLength`, `enum`, `required`, etc. in the
`FeatureConfigSchema.Add(...)` block), check what the field's actual
*consumer* code does with edge-case values (nil, empty, zero) — grep for
where the field is read at runtime. A constraint that's stricter than the
runtime semantics can silently make a legitimately meaningful value
unreachable. Concrete case: `PhoneInputFeatureConfig.allowlist` had
`"minItems": 1`, but `IntersectAllowlist` (`pkg/lib/config/utils.go`) already
treated an empty allowlist as "no restriction" — the schema blocked the one
input (`allowlist: []`) that would have cleanly expressed "clear this
override," forcing an awkward, undiscoverable workaround
(`phone_input: {}` with the field omitted) instead. Don't add a schema
constraint "for safety" without confirming the runtime already needs it.

### Don't tag a scalar field `omitempty` if its zero value is a real default

`SetFieldDefaults` (`pkg/lib/config/default.go`) already makes every section
pointer non-nil via its generic reflection walk, regardless of whether that
section implements `SetDefaults()` — a section is never actually "absent" in
a parsed/defaulted `FeatureConfig`. If a plain (non-pointer) `bool`/`string`/
`int`/`float` field's zero value (`false`/`""`/`0`) *is* that field's real,
intended default — not a stand-in for "not set" — tagging it `omitempty`
doesn't skip anything meaningful when parsing, but it does hide that value
from JSON *output*: `encoding/json` treats the zero value as "empty" and
omits the key, so a fully-resolved section marshals as `{}` instead of e.g.
`{"disabled": false}`. This makes the Site Admin API's
`effective_plan_feature_config`/`effective_app_feature_config` (and any
other JSON consumer of `FeatureConfig`) show a resolved section as if it
were empty/unset. Fix: drop `omitempty` — plain `json:"disabled"`. This is
always safe, because `Merge()` for these fields already operates on the
*section's* pointer-nil-ness (see the field-level merge rule above), never
the leaf scalar's zero value — removing `omitempty` never changes merge or
validation behavior, only what the field looks like once marshaled.

This is a different situation from the slice case in "Schema/runtime
consistency" above (`PhoneInputFeatureConfig.allowlist`): there, `nil` and an
explicit empty slice are two *different* meaningful values (inherit vs.
explicitly cleared), so the fix was `omitzero` (which only omits the true
zero value, `nil`), not simply dropping the tag. A plain scalar only has one
value to begin with, so just remove `omitempty` entirely — don't reach for
`omitzero` there, it would be a no-op.

**Test with a real marshal, not `ShouldResemble` on parsed structs.** Every
existing test in this package compares parsed Go *structs*, which can't tell
`omitempty` apart from no tag at all — that's exactly why this went
unnoticed for nine fields across five sections. Marshal with
`encoding/json.Marshal` (or `sigs.k8s.io/yaml.Marshal`, which calls it
internally — this is what `viewEffectiveResource`'s merge fold does) and
assert on the resulting shape, e.g.
`TestFeatureConfigDisabledFieldsSerializeExplicitly` in `feature_test.go`.

**When auditing for this, grep the whole package by field, not file by
file.** A file having a `SetDefaults()` for one field doesn't mean every
field in that file is covered — `feature_identity.go` has one for
`BiometricFeatureConfig` (a pointer-scalar field) while
`LoginIDPhoneFeatureConfig.Disabled` and
`OAuthSSOProviderFeatureConfig.Disabled` (plain-bool, single-field sections
in that same file) still had the bug. Use:

```
grep -nE '^\s*[A-Z][A-Za-z0-9_]*\s+(bool|string|int|int32|int64|float32|float64)\s+`json:"[^"]*,omitempty"`' pkg/lib/config/feature_*.go
```

## Secrets (`authgear.secrets.yaml`)

Secret item schema/types live in `pkg/lib/config/secret*.go`, registered via
`SecretConfigSchema.Add(...)` and `secretItemKeys` (`secret.go`) the same
way project-config/feature-config sections register with `Schema`/
`FeatureConfigSchema`. Two independent things need coverage, per the table
above:

- **Schema-level parsing** (`config.ParseSecret`) — does the item's JSON
  shape validate on its own. Cases go in `parse_secret_tests.yaml`.
- **Cross-validation against `AppConfig`** (`SecretConfig.Validate`) — does
  the secret make sense *given* the current project config (e.g. does a
  `stream_name` on a `telemetry.audit_logs.streams.tls` item correspond to a
  configured stream; is a referenced key actually used anywhere). Cases go
  in `secret_config_validate_tests.yaml`, with both `app_config:` and
  `secret_config:` set up together. Only needed if the change involves a
  relationship between a secret and app config — a self-contained secret
  field only needs the schema-level case.

### Pitfall: `SetFieldDefaults` materializes every omitted struct field

`SetFieldDefaults` (`pkg/lib/config/default.go`) walks every field of every
parsed config/secret recursively and, for any `nil` pointer-to-struct field,
allocates a zero-value struct in its place — this runs on secret item data
too (`SecretItem.parse` calls it, same as `config.Parse` does for
`AppConfig`). This means an **optional** struct field that was genuinely
omitted from the YAML is *not* `nil` by the time your code reads it — it's a
non-nil struct with all-zero-value fields, indistinguishable from `nil` by a
naive `!= nil` check.

This caused a real bug: a `*TelemetryAuditLogStreamClientCertificate` field
left out of a secret item because only `certificate_authority` was
configured got materialized into an empty (non-nil) struct, so
`item.ClientCertificate != nil` was true even when nothing was configured,
and the code tried to parse an empty certificate and failed.

**The fix is `nullable:"true"` on the field tag**, not a deeper "is this
struct actually empty" check downstream:

```go
type TelemetryAuditLogStreamTLSMaterialsItem struct {
	StreamName           string                                     `json:"stream_name,omitempty"`
	ClientCertificate    *TelemetryAuditLogStreamClientCertificate  `json:"client_certificate,omitempty" nullable:"true"`
	CertificateAuthority *X509Certificate                           `json:"certificate_authority,omitempty" nullable:"true"`
}
```

`nullable:"true"` tells `SetFieldDefaults` to skip recursing into that field
entirely, so a genuinely-omitted field stays `nil`. This is already the
established pattern elsewhere — see `Usage *UsageConfig` in `config.go`,
`BotProtection *AuthenticationFlowBotProtection` in `authentication_flow.go`,
and the fields in `bot_protection.go`/`fraud_protection.go`. **Any time you
add an optional `*Struct` field to project config, feature config, or a
secret item where "omitted" and "present but zero-valued" are meaningfully
different states to your code, tag it `nullable:"true"` and check `!= nil`
downstream** — don't discover the need for it by checking a substring/zero
value instead, and don't skip it because "it doesn't have `SetDefaults()` so
there's nothing to default" (the generic reflection walk materializes it
regardless of whether the type has its own `SetDefaults()`).

## References

- `pkg/lib/config/config.go` — `AppConfig`, `Schema`, `AppConfig.Validate`
- `pkg/lib/config/feature.go` — top-level `FeatureConfig.Merge` dispatcher
- `pkg/lib/config/feature_*.go` — per-section feature-config `Merge` implementations
- `pkg/lib/config/secret.go` — `SecretConfig`, `SecretConfigSchema`, `secretItemKeys`, `SecretItem.parse`
- `pkg/lib/config/secret_*.go` — per-key secret types
- `pkg/lib/config/default.go` — `SetFieldDefaults`, the generic reflection walk, and the `nullable` tag
- `pkg/lib/config/testdata/config_tests.yaml` — project config parse+validate test fixture
- `pkg/lib/config/testdata/parse_feature_tests.yaml` — feature config schema validation test fixture
- `pkg/lib/config/testdata/merge_feature.yaml` — feature config merge test fixture
- `pkg/lib/config/testdata/parse_secret_tests.yaml` — secret schema validation test fixture
- `pkg/lib/config/testdata/secret_config_validate_tests.yaml` — secret × app-config cross-validation test fixture
- `pkg/lib/config/config_test.go`'s `TestAppConfig` / `pkg/lib/config/feature_test.go`'s `TestParseFeatureConfig` / `pkg/lib/config/secret_test.go`'s `TestParseSecret` and `TestSecretConfigValidate` — the tests that consume the fixtures above
- `pkg/lib/config/feature_test.go`'s `TestFeatureConfigDisabledFieldsSerializeExplicitly` — marshal-based test pattern for the omitempty rule above
