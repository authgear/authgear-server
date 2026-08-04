---
name: update-feature-config
description: Add or change a feature config field/section in pkg/lib/config (authgear.features.yaml). Use when adding a new feature config field, adding a new top-level section, or touching an existing section's Merge implementation.
---

# Update Feature Config

Feature config (`authgear.features.yaml`) is merged across layers — code
default ← cluster ← plan ← app override — via `FeatureConfig.Merge`
(`pkg/lib/config/feature.go`), which reflects over every top-level field and
dispatches to that section's own `Merge` in `pkg/lib/config/feature_*.go`.

## Hard requirement: merge must be field-level, never whole-section replace

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

## Required test coverage

Every new/changed `Merge` implementation needs a case in
`pkg/lib/config/testdata/merge_feature.yaml`: one layer sets field `A` only,
a later layer sets a sibling field `B` only (never repeating `A`) — assert
the final effective config has **both** `A` (from the first layer) and `B`
(from the second), not `A` reset to its default. Pick values that are **not**
the code default for the field being tested — otherwise a whole-section
regression would silently produce the "right" value by accident and the test
wouldn't catch it. See the existing `hook`/`collaborator`/`identity` cases in
that file for the pattern.

## Schema/runtime consistency

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

## Don't tag a scalar field `omitempty` if its zero value is a real default

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

## References

- `pkg/lib/config/feature.go` — top-level `FeatureConfig.Merge` dispatcher
- `pkg/lib/config/feature_*.go` — per-section `Merge` implementations
- `pkg/lib/config/testdata/merge_feature.yaml` — shared merge test fixture
- `pkg/lib/config/testdata/parse_feature_tests.yaml` — schema validation test fixture
- `pkg/lib/config/default.go` — `SetFieldDefaults`, the generic reflection walk
- `pkg/lib/config/feature_test.go`'s `TestFeatureConfigDisabledFieldsSerializeExplicitly` — marshal-based test pattern for the omitempty rule above
