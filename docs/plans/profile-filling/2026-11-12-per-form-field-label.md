# Per-form field label implementation plan

Spec: [docs/specs/user-profile/profile-filling.md](../../specs/user-profile/profile-filling.md#translations) (`v2.page.fill-form.{name}.field-label-{key}`).

## 1. Goal / scope

Let a `user_profile` field of a named form have its own label, so the same attribute can read
differently per form, for example a consent checkbox whose form label carries a Privacy Policy link
while the settings page shows the plain attribute label.

Label lookup order for a field in a form named `{name}`, where `{key}` is the field's `key` in
`fill_form_data` (`FillFormDataField.Key`; for a `user_profile` field, its pointer):

1. `v2.page.fill-form.{name}.field-label-{key}` — only when the form has a `name`.
2. `standard-attribute-label-{pointer}` or `custom-attribute-label-{pointer}` (unchanged).
3. `labelutil.Label` of the first pointer token (unchanged).

In scope: label resolution in the Authflow `fill_form_data` builder, unit tests, e2e test.

Out of scope: `enum` option labels and `pattern` messages (stay per attribute), the settings page
(keeps using the attribute label), portal UI for editing translations.

## 2. Config model and schema

None. The new key is a translation, not config. Form `name` already exists
(`config.AuthenticationFlowForm.Name`), and the JSON schema in `pkg/lib/config/authentication_flow.go`
already rejects the name `default`, so the built-in `v2.page.fill-form.default.*` keys
cannot collide with a field label key.

No built-in translation is added to `resources/authgear/templates/en/translation.json`: the key
exists only when the developer defines it.

## 3. Runtime flow

Both places that show a form build its data through one function, so one change covers the
`fill_form` step and forms in the `identify` step, and both Auth UI and the Authflow API:

- `pkg/lib/authenticationflow/declarative/intent_fill_form.go` (`fill_form` step) →
  `NewFillFormDataForm`
- `pkg/lib/authenticationflow/declarative/form.go` (`identify` step form) → `NewFillFormDataForm`
- Auth UI (`pkg/auth/handler/webapp/authflowv2/form.go`) copies `FillFormDataField.Label` into
  `Label` / `LabelHTML`; no template change.

File: `pkg/lib/authenticationflow/declarative/data_fill_form.go`.

Change the signature of `formFieldLabel` to take the form name and the field key. The key is passed
separately from the pointer because the translation key is defined on the field's `key` in
`fill_form_data`, not on the attribute; today the two values are equal.

```go
func formFieldLabel(ctx context.Context, deps *authflow.Dependencies, formName string, fieldKey string, pointer string, custom *config.CustomAttributesAttributeConfig) (string, error)
```

Call flow inside `formFieldLabel`:

1. If `formName != ""`: `key := "v2.page.fill-form." + formName + ".field-label-" + fieldKey`.
   Call `deps.Translations.HasKey(ctx, key)`.
   - error → return `"", err`.
   - `true` → return `deps.Translations.RenderText(ctx, key, nil)`.
   - `false` → continue.
2. Existing behaviour: `standard-attribute-label-` / `custom-attribute-label-` + pointer via
   `HasKey` + `RenderText`; otherwise `labelutil.Label(firstPointerToken(pointer))`.

In `NewFillFormDataForm`, the loop already computes `pointer := formFieldKey(f)`, which becomes
`FillFormDataField.Key`. Rename it to `key`, keep `pointer := key` for the attribute lookups, and
change the call to:

```go
label, err := formFieldLabel(ctx, deps, form.Name, key, pointer, custom)
```

`FillFormDataField.Key` is set from `key`, so the translation key and the API `key` come from the
same value.

The existing post-processing stays after the lookup: a `boolean` field's label, whichever key it
came from, is passed through `htmlutil.SanitizeLinksHTML`. Non-boolean labels stay plain text and
are escaped by Auth UI, as today.

A form without a `name` skips step 1, so its behaviour is unchanged.

## 4. Compatibility and deployment behavior

- Additive. Forms without a `name`, and named forms without the new key, render the same labels as
  before.
- No stored state, no API shape change: `fill_form_data.form.fields[].label` keeps its type and
  meaning ("localized label"); only its source can differ.
- Not a breaking change; no `docs/BREAKING-CHANGES.md` entry.

## 5. File-level change plan

| File | Change |
|---|---|
| `pkg/lib/authenticationflow/declarative/data_fill_form.go` | `formFieldLabel` takes `formName`, looks up the per-form key first; `NewFillFormDataForm` passes `form.Name`. |
| `pkg/lib/authenticationflow/declarative/data_fill_form_test.go` | New Convey cases, see below. |
| `e2e/tests/profile_filling/var_translations/templates/en/translation.json` | Add `v2.page.fill-form.consent.field-label-/x_newsletter`. |
| `e2e/tests/profile_filling/per_form_field_label.test.yaml` | New e2e test. |

## 6. Test plan

### Unit tests

`pkg/lib/authenticationflow/declarative/data_fill_form_test.go`, Convey style, inside
`TestNewFillFormDataForm` (its `build` helper uses form name `names`; the `fakeTranslationService`
map is reset per leaf):

- **form field label overrides the attribute label**: set
  `standard-attribute-label-/given_name` = "First name" and
  `v2.page.fill-form.names.field-label-/given_name` = "Legal first name"; also
  `v2.page.fill-form.names.field-label-/x_employee_id` = "Staff number" with no attribute label.
  Expect "Legal first name" and "Staff number".
- **form field label of another form is not used**: set
  `v2.page.fill-form.other.field-label-/given_name` = "Other"; call `NewFillFormDataForm` directly
  with `Name: "names"` and with `Name: ""`. Expect "Given Name" (the fallback) in both.
- **boolean form field label is sanitized**: set
  `v2.page.fill-form.names.field-label-/x_terms` to a label with `onclick` and `<script>`; expect the
  same sanitized output as the existing "boolean label is sanitized" case.

Run: `go test ./pkg/lib/authenticationflow/declarative/...`.

### E2E

New `e2e/tests/profile_filling/per_form_field_label.test.yaml`, `extra_files_directory: ./var_translations`,
Authflow API style as `boolean_terms.test.yaml`. Custom attributes `/x_terms` and `/x_newsletter`
(boolean, `end_user: readwrite`). Signup flow: `identify` (username) + `create_authenticator`
(password), then:

1. `fill_form` with `form.name: consent`, fields `/x_terms` (required) and `/x_newsletter`.
2. `fill_form` without `name`, field `/x_newsletter`.

Translation added to `var_translations/templates/en/translation.json`:

```json
"v2.page.fill-form.consent.field-label-/x_newsletter": "Send me news. See our <a href=\"https://example.com/privacy\">Privacy Policy</a>."
```

Cases:

- Page 1: `/x_newsletter` label is the per-form text, sanitized (`target="_blank" rel="noopener noreferrer"` added);
  `/x_terms` label is the attribute label from `custom-attribute-label-/x_terms`.
- Page 2 (unnamed form): `/x_newsletter` label is "X Newsletter", proving the override is per form.
- Submitting both pages finishes the flow.

The added key does not affect the other tests sharing `var_translations`: none of them has a form
named `consent`.

Run per the `write-e2e-test` skill (unset `HTTP(S)_PROXY`).

## 7. Fixed behavioral decisions

- The translation key uses the field's `key` from `fill_form_data`, not the attribute pointer, so it
  identifies the field rather than the attribute.
- Only forms with a `name` can override labels; there is no `default` field label key, because the
  attribute label already is the default for every form.
- An empty translation for the per-form key is used as is (an empty label); it does not fall back.
  This matches `HasKey` semantics used by the existing attribute-label lookup.
- `enum` option labels and `pattern` messages are not overridable per form.

## 8. Atomic commit plan

1. **`[Profile Filling] Support per-form field labels`**
   - `pkg/lib/authenticationflow/declarative/data_fill_form.go`
   - `pkg/lib/authenticationflow/declarative/data_fill_form_test.go`
   - Behaviour change in section 3 plus unit tests. No wiring or generated files.
2. **`[Profile Filling] Add e2e test for per-form field labels`**
   - `e2e/tests/profile_filling/per_form_field_label.test.yaml`
   - `e2e/tests/profile_filling/var_translations/templates/en/translation.json`
